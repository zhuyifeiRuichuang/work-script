#!/bin/bash
# ubutnu24检测服务器配置
# 检查是否以root权限运行（某些信息需要root权限）
if [ "$(id -u)" -ne 0 ]; then
    echo "警告：部分硬件信息可能需要root权限才能完全显示"
    echo "建议使用sudo运行此脚本以获取完整信息"
    echo
fi

echo "==================== 服务器硬件信息 ===================="
echo

# CPU信息
echo "---------------------- CPU信息 -----------------------"
echo -n "CPU型号: "
grep -m1 'model name' /proc/cpuinfo | cut -d: -f2 | sed -e 's/^ *//'

echo -n "物理核心数: "
grep 'cpu cores' /proc/cpuinfo | uniq | cut -d: -f2 | sed -e 's/^ *//'

echo -n "逻辑核心数: "
grep -c '^processor' /proc/cpuinfo
echo

# 内存信息
echo "---------------------- 内存信息 ----------------------"
echo -n "总内存容量: "
free -h | awk '/Mem:/ {print $2}'

if [ "$(id -u)" -eq 0 ]; then
    echo -n "内存条数量: "
    dmidecode -t memory | grep -c '^Memory Device$'
    
    echo "内存详情:"
    dmidecode -t memory | awk '/Memory Device$/{print "\n插槽"++count; flag=1} flag && /Size: [0-9]/ {print "  容量: "$2" "$3; flag=0} /Speed: [0-9]/ {print "  主频: "$2" MHz"}' | grep -v 'No Module Installed'
else
    echo "提示: 要查看内存条数量和主频，请使用root权限运行脚本"
fi
echo

# 磁盘信息
echo "---------------------- 磁盘信息 ----------------------"
echo "磁盘列表及容量:"
lsblk -o NAME,SIZE,TYPE,MOUNTPOINT | grep -v 'loop'

echo -n "物理磁盘数量: "
lsblk -d | grep -c '^sd'
echo

# GPU信息
echo "---------------------- GPU信息 -----------------------"
gpu_count=$(lspci | grep -i 'vga\|3d\|display' | wc -l)
echo "GPU数量: $gpu_count"

if [ $gpu_count -gt 0 ]; then
    echo "GPU型号:"
    lspci | grep -i 'vga\|3d\|display' | cut -d: -f3- | sed -e 's/^ *//'
else
    echo "未检测到GPU"
fi

echo
echo "======================================================"
