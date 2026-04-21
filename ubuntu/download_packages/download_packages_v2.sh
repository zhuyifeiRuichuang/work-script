#!/bin/bash

# 初始化变量
package_name=""
download_dir=""
dist=""
package_list_file=""
archive_target_dir="/tmp"

# 显示帮助信息
usage() {
  echo "使用方法: $0 [选项]"
  echo "选项:"
  echo "  -n, --name <软件包名称>   指定要下载的软件包名称（与-f互斥）"
  echo "  -f, --file <清单文件>     指定包含软件包名称的文件（与-n互斥）"
  echo "  -d, --dir <下载目录>      指定下载目录（可选，默认: /tmp/<软件包名称>/）"
  echo "  -a, --archive <打包目录>   指定打包目标目录（可选，默认: /tmp）"
  echo "  -h, --help                显示此帮助信息"
  echo ""
  echo "示例:"
  echo "  $0 -n vim                    # 下载vim及其依赖"
  echo "  $0 -f packages.txt           # 从文件批量下载多个软件包"
  echo "  $0 -n vim -d /home/user/downloads  # 指定下载目录"
  echo "  $0 -f packages.txt -a /backup     # 指定打包目标目录"
  exit 1
}

# 获取当前系统的发行版版本代号
get_distribution() {
  # 检查是否存在os-release文件
  if [ -f /etc/os-release ]; then
    . /etc/os-release
    # 返回版本代号，如果不存在则返回发行版名称
    if [ -n "$VERSION_CODENAME" ]; then
      echo "$VERSION_CODENAME"
    else
      echo "$ID$VERSION_ID" | tr '[:upper:]' '[:lower:]'
    fi
  else
    #  fallback方案
    echo "$(lsb_release -cs 2>/dev/null || echo "unknown")"
  fi
}

# 下载单个软件包及其依赖
download_package() {
  local pkg_name="$1"
  local pkg_download_dir="$2"
  
  # 创建下载目录（静默）
  mkdir -p "$pkg_download_dir" 2>/dev/null || return 1

  # 切换到下载目录
  cd "$pkg_download_dir" 2>/dev/null || return 1

  # 检查软件包是否存在
  if ! apt-cache show "$pkg_name" >/dev/null 2>&1; then
    return 1
  fi
  
  # 获取直接依赖包列表（多种方法尝试）
  dependencies=$(apt-cache depends "$pkg_name" 2>/dev/null | grep -E "(Depends|PreDepends):" | cut -d: -f2 | sed 's/^ *//' | grep -v "^$" | sort -u)
  
  # 如果第一种方法失败，尝试更宽松的方法
  if [ -z "$dependencies" ]; then
    dependencies=$(apt-cache depends "$pkg_name" 2>/dev/null | grep "Depends" | cut -d: -f2 | sed 's/^ *//' | grep -v "^$" | sort -u)
  fi
  
  # 如果仍然失败，尝试获取所有可能的依赖
  if [ -z "$dependencies" ]; then
    dependencies=$(apt-cache depends "$pkg_name" 2>/dev/null | grep -v "^[A-Za-z]*:" | grep "^\w" | sort -u)
  fi
  
  # 最后兜底：尝试从apt-cache show中获取信息
  if [ -z "$dependencies" ]; then
    dependencies=$(apt-cache show "$pkg_name" 2>/dev/null | grep "^Depends:" | cut -d: -f2 | sed 's/^ *//' | sed 's/,//g' | sed 's/(.*)//g' | tr ' ' '\n' | grep -v "^$" | sort -u)
  fi
  
  if [ -z "$dependencies" ]; then
    dependencies="$pkg_name"
  else
    # 添加主包到依赖列表中（确保包含主包）
    dependencies="$dependencies $pkg_name"
  fi
  
  # 去重
  dependencies=$(echo "$dependencies" | tr ' ' '\n' | sort -u | tr '\n' ' ')
  
  # 验证依赖包是否存在，过滤掉不存在的包
  valid_dependencies=""
  for dep in $dependencies; do
    if apt-cache show "$dep" >/dev/null 2>&1; then
      valid_dependencies="$valid_dependencies $dep"
    fi
  done
  
  # 使用验证后的依赖列表
  dependencies="$valid_dependencies"
  
  # 计算总包数用于进度显示
  local total_deps=$(echo "$dependencies" | wc -w)
  local current=0
  
  if [ "$total_deps" -eq 0 ]; then
    echo "错误：无法找到有效的软件包依赖"
    return 1
  fi
  
  echo "正在处理软件包 $pkg_name 及依赖包下载（共 $total_deps 个包）..."
  
  # 分批下载并显示进度
  local failed_downloads=""
  for dep in $dependencies; do
    current=$((current + 1))
    local progress=$((current * 100 / total_deps))
    echo -ne "\r进度: $progress% ($current/$total_deps)"
    
    if ! sudo apt-get download "$dep" >/dev/null 2>&1; then
      failed_downloads="$failed_downloads $dep"
    fi
  done
  
  echo ""
  
  # 检查下载是否成功
  local deb_count=$(ls *.deb 2>/dev/null | wc -l)
  [ "$deb_count" -gt 0 ]
}

# 打包软件包目录
archive_packages() {
  local source_dir="$1"
  local archive_name="$2"
  
  # 检查缓存目录大小
  local cache_size=$(du -sh "$source_dir" 2>/dev/null | cut -f1)
  echo "正在打包缓存目录: $source_dir (大小: ${cache_size:-未知})"
  
  # 切换到源目录的上一级目录进行打包
  local parent_dir=$(dirname "$source_dir")
  local dir_name=$(basename "$source_dir")
  
  cd "$parent_dir" || {
    echo "错误：无法进入目录 $parent_dir" >&2
    return 1
  }
  
  # 固定使用tar.gz压缩
  echo "开始压缩，请稍候..."
  
  # 执行打包（后台运行以显示进度）
  tar -czf "$archive_name" "$dir_name" 2>/dev/null &
  local tar_pid=$!
  
  # 显示进度提示（每3秒一次）
  local count=0
  while kill -0 $tar_pid 2>/dev/null; do
    sleep 3
    count=$((count + 3))
    echo "打包进行中... (${count}秒)"
  done
  
  wait $tar_pid
  local tar_result=$?
  
  # 检查打包是否成功
  if [ $tar_result -eq 0 ] && [ -f "$archive_name" ]; then
    # 显示压缩包大小
    local archive_size=$(du -sh "$archive_name" 2>/dev/null | cut -f1)
    echo "压缩完成: $archive_name (大小: ${archive_size:-未知})"
    return 0
  else
    echo "错误：打包过程中出现错误" >&2
    return 1
  fi
}

# 清理缓存目录
clean_cache() {
  local cache_dir="$1"
  
  if [ -d "$cache_dir" ]; then
    echo "正在清理缓存目录: $cache_dir"
    rm -rf "$cache_dir"
    if [ $? -eq 0 ]; then
      echo "缓存目录清理完成"
    else
      echo "警告：缓存目录清理失败" >&2
    fi
  fi
}

# 从文件读取软件包列表
read_package_list() {
  local file="$1"
  
  if [ ! -f "$file" ]; then
    echo "错误：清单文件 $file 不存在" >&2
    exit 1
  fi
  
  # 读取文件中的软件包名称，过滤空行和注释行
  local packages=()
  while IFS= read -r line; do
    # 跳过空行和以#开头的注释行
    if [[ -n "$line" && ! "$line" =~ ^[[:space:]]*# ]]; then
      # 去除前后空格
      package=$(echo "$line" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
      if [ -n "$package" ]; then
        packages+=("$package")
      fi
    fi
  done < "$file"
  
  if [ ${#packages[@]} -eq 0 ]; then
    echo "错误：清单文件 $file 中没有找到有效的软件包名称" >&2
    exit 1
  fi
  
  echo "${packages[@]}"
}

# 解析命令行参数（支持长选项）
OPTIONS=$(getopt -o n:f:d:a:h --long name:,file:,dir:,archive:,help -n "$0" -- "$@")

if [ $? -ne 0 ]; then
  echo "解析参数错误，请使用 -h 查看帮助" >&2
  exit 1
fi

eval set -- "$OPTIONS"

while true; do
  case "$1" in
    -n|--name)
      package_name="$2"
      shift 2
      ;;
    -f|--file)
      package_list_file="$2"
      shift 2
      ;;
    -d|--dir)
      download_dir="$2"
      shift 2
      ;;
    -a|--archive)
      archive_target_dir="$2"
      shift 2
      ;;
    -h|--help)
      usage
      ;;
    --)
      shift
      break
      ;;
    *)
      echo "内部错误" >&2
      exit 1
      ;;
  esac
done

# 参数验证
if [ -n "$package_name" ] && [ -n "$package_list_file" ]; then
  echo "错误：-n 和 -f 参数互斥，只能使用其中一个" >&2
  usage
fi

if [ -z "$package_name" ] && [ -z "$package_list_file" ]; then
  echo "错误：必须指定软件包名称 (-n) 或清单文件 (-f)" >&2
  usage
fi

# 获取发行版信息
dist=$(get_distribution)
if [ -z "$dist" ] || [ "$dist" = "unknown" ]; then
  echo "警告：无法识别操作系统发行版，将使用'unknown'作为发行版标识" >&2
  dist="unknown"
fi

# 创建打包目标目录
mkdir -p "$archive_target_dir"
if [ $? -ne 0 ]; then
  echo "错误：无法创建打包目录 $archive_target_dir" >&2
  exit 1
fi

# 处理单个软件包
if [ -n "$package_name" ]; then
  # 设置默认下载目录（静默）
  if [ -z "$download_dir" ]; then
    download_dir="/tmp/$package_name/"
  fi
  
  # 下载软件包
  if download_package "$package_name" "$download_dir"; then
    echo "缓存目录: $(realpath "$download_dir")"
    
    # 打包
    archive_name="${package_name}_${dist}.tar.gz"
    if archive_packages "$download_dir" "$archive_target_dir/$archive_name"; then
      echo "下载成功！打包完成: $archive_target_dir/$archive_name"
      
      # 自动清理缓存目录
      clean_cache "$download_dir"
    else
      exit 1
    fi
  else
    echo "下载失败" >&2
    exit 1
  fi
fi

# 处理批量软件包
if [ -n "$package_list_file" ]; then
  packages=$(read_package_list "$package_list_file")
  
  # 设置默认下载目录（静默）
  if [ -z "$download_dir" ]; then
    download_dir="/tmp/batch_packages/"
  fi
  
  # 创建主下载目录（静默）
  mkdir -p "$download_dir"
  
  successful_packages=()
  failed_packages=()
  
  # 批量处理每个软件包
  for package in $packages; do
    pkg_download_dir="$download_dir$package/"
    
    if download_package "$package" "$pkg_download_dir"; then
      successful_packages+=("$package")
    else
      failed_packages+=("$package")
    fi
    echo ""
  done
  
  # 显示结果
  if [ ${#successful_packages[@]} -gt 0 ]; then
    if [ ${#failed_packages[@]} -eq 0 ]; then
      echo "全部下载成功！成功处理 ${#successful_packages[@]} 个软件包"
    else
      echo "部分成功。成功 ${#successful_packages[@]} 个，失败 ${#failed_packages[@]} 个"
      echo "失败的软件包: ${failed_packages[*]}"
    fi
    
    # 显示缓存目录绝对路径
    echo "缓存目录: $(realpath "$download_dir")"
    
    # 打包整个下载目录
    timestamp=$(date +"%Y%m%d_%H%M%S")
    archive_name="batch_packages_${dist}_${timestamp}.tar.gz"
    full_archive_path="$archive_target_dir/$archive_name"
    
    if archive_packages "$download_dir" "$archive_name"; then
      echo "打包完成: $full_archive_path"
      
      # 自动清理缓存目录
      clean_cache "$download_dir"
    fi
  else
    echo "错误：没有成功处理任何软件包" >&2
    exit 1
  fi
fi

echo "脚本执行完成！"