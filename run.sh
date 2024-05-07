#!/bin/bash

# 获取命令行参数
sni=$1
input=$2
output=$3
cert=$4


# 定义执行命令并记录时间的函数
run_command() {
    local command=$1
    local start_time=$(date +%s)
    echo "Running command: $command"
    eval $command
    local end_time=$(date +%s)
    echo "Execution time for '$command': $((end_time - start_time)) seconds"
    echo "-------------------------------------"
}

# 根据 sni 参数决定是否添加 -sni true
sni_option=""
if [ "$sni" = "true" ]; then
    sni_option="-sni"
fi

mkdir -p $output $cert

# 执行命令并记录时间
run_command "./doq -port 8853 $sni_option $input/8853.csv $output/port8853.csv $cert/8853.csv"
run_command "./doq -port 784 $sni_option $input/784.csv $output/port784.csv $cert/784.csv"
run_command "./doq -port 853 $sni_option $input/853.csv $output/port853.csv $cert/853.csv"
run_command "./doq -port 443 $sni_option $input/443.csv $output/port443.csv $cert/443.csv"
