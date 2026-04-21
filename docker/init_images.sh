#!/usr/bin/env bash
# --------------------------------------------------
#  1. 按行读取镜像列表
#  2. 拉取失败→等待3分钟→重试，直到成功
#  3. 统一改名为 zhuyifeiruichuang/ 前缀
#  4. 改名后自动 docker push
# --------------------------------------------------

set -euo pipefail

# ========== 用户参数 ==========
REGISTRY_DST="zhuyifeiruichuang"          # 最终前缀
IMAGE_LIST_FILE="${1:-image.list}"        # 镜像列表文件，默认 image.list
# ==============================

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

# 把列表里所有 registry.cn-beijing.aliyuncs.com/kubesphereio/ 去掉，得到短名
normalize_name() {
    local img="$1"
    img="${img#registry.cn-beijing.aliyuncs.com/kubesphereio/}"
    echo "$img"
}

while IFS= read -r img; do
    [[ -z "$img" || "$img" =~ ^# ]] && continue      # 跳过空行或注释
    short=$(normalize_name "$img")
    new_tag="${REGISTRY_DST}/${short}"

    # ---------- 拉取 ----------
    if ! docker image inspect "$img" &>/dev/null; then
        echo -e "${GREEN}>>> 拉取 $img${NC}"
        while true; do
            if docker pull "$img"; then
                break
            else
                echo -e "${RED}!!! 拉取失败，3 分钟后重试...${NC}"
                sleep 180
            fi
        done
    else
        echo ">>> $img 已存在，跳过拉取"
    fi

    # ---------- 改名 ----------
    if ! docker image inspect "$new_tag" &>/dev/null; then
        echo ">>> 改名 $img  ->  $new_tag"
        docker tag "$img" "$new_tag"
    else
        echo ">>> $new_tag 已存在，跳过改名"
    fi

    # ---------- 推送 ----------
    echo ">>> 推送 $new_tag"
    docker push "$new_tag"
    echo -e "${GREEN}✔ 完成 $new_tag${NC}\n"
done < "$IMAGE_LIST_FILE"
