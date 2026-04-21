#!/usr/bin/env bash

# 遍历本地所有镜像
docker images --format "{{.Repository}}:{{.Tag}}" | while IFS= read -r old_tag; do
    # 跳过 <none>:<none> 这类无效镜像
    [[ "$old_tag" == *"<none>"* ]] && continue

    new_tag="zhuyifeiruichuang/$(basename "$old_tag")"
    echo "== Tag & Push: $old_tag  ->  $new_tag"
    docker tag "$old_tag" "$new_tag"
    docker push "$new_tag"
done

echo "All images pushed under zhuyifeiruichuang/  ✅"
