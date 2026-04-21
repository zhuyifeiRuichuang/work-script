#!/usr/bin/env bash
set -e

usage() {
  echo "用法："
  echo "  全集群全部 Pod        : ./podOwner.sh -A"
  echo "  某命名空间全部 Pod    : ./podOwner.sh -n <namespace>"
  echo "  全集群查单个 Pod      : ./podOwner.sh -p <pod-name>"
  echo "  指定命名空间查 Pod    : ./podOwner.sh -p <pod-name> -n <namespace>"
  exit 1
}

POD_NAME=""
NAMESPACE=""
ALL=false

while getopts ":Ap:n:" opt; do
  case "$opt" in
    A) ALL=true ;;
    p) POD_NAME="$OPTARG" ;;
    n) NAMESPACE="$OPTARG" ;;
    *) usage ;;
  esac
done

# 至少传一个参数
[[ "$ALL" == false && -z "$POD_NAME" && -z "$NAMESPACE" ]] && usage

# 构造 kubectl 参数
if [[ "$ALL" == true ]]; then
  NS_FILTER="-A"
elif [[ -n "$NAMESPACE" ]]; then
  if ! kubectl get ns "$NAMESPACE" &>/dev/null; then
    echo "❌ 命名空间 \"$NAMESPACE\" 不存在！"
    kubectl get ns -o name | sed 's|namespace/|  |'
    exit 1
  fi
  NS_FILTER="-n $NAMESPACE"
else
  NS_FILTER="-A"
fi

json=$(kubectl get pods $NS_FILTER -o json)

rows=$(echo "$json" |
  jq -r --arg p "$POD_NAME" '
    def final_name(owner):
      if owner == null or owner.kind == null or owner.name == null then
        "<none>"
      else
        (owner.kind + "/" + owner.name
         | sub("ReplicaSet/(?<n>.*)-[^-]+$";  "Deployment/\(.n)")
         | sub("Job/(?<n>.*)-[^-]+$";         "CronJob/\(.n)")
         | sub("^<none>/";                    "<none>"))
      end;

    .items[]
    | if ($p != "") then select(.metadata.name == $p) else . end
    | (.metadata.ownerReferences // [])[0] as $o
    | [
        .metadata.namespace,
        .metadata.name,
        ($o.kind // "<none>") + "/" + ($o.name // "<none>"),
        final_name($o)
      ] | @tsv
  ')

[[ -z "$rows" ]] && {
  [[ -n "$POD_NAME" ]] && echo "未找到 Pod \"$POD_NAME\" ${NAMESPACE:+in namespace \"$NAMESPACE\"}"
  [[ -z "$POD_NAME" ]] && echo "命名空间 \"$NAMESPACE\" 中没有任何 Pod"
  exit 0
}

awk -F'\t' '
BEGIN {
  hdr[1]="NAMESPACE"; hdr[2]="POD"; hdr[3]="CONTROLLER"; hdr[4]="FINAL_CONTROLLER"
  for (i=1;i<=4;i++) w[i]=length(hdr[i])
}
{
  for (i=1;i<=4;i++) if (length($i) > w[i]) w[i]=length($i)
  line[NR]=$0
}
END {
  # 上边框
  printf "+"; for (i=1;i<=4;i++) printf "%*s+", w[i]+2, gensub(/./,"-","g",sprintf("%*s",w[i],"")); printf "\n"
  # 表头
  printf "|"; for (i=1;i<=4;i++) printf " %-*s |", w[i], hdr[i]; printf "\n"
  # 分隔线
  printf "+"; for (i=1;i<=4;i++) printf "%*s+", w[i]+2, gensub(/./,"-","g",sprintf("%*s",w[i],"")); printf "\n"
  # 数据
  for (j=1;j<=NR;j++) { split(line[j],f,"\t"); printf "|"; for (i=1;i<=4;i++) printf " %-*s |", w[i], f[i]; printf "\n" }
  # 下边框
  printf "+"; for (i=1;i<=4;i++) printf "%*s+", w[i]+2, gensub(/./,"-","g",sprintf("%*s",w[i],"")); printf "\n"
}' <<< "$rows"
