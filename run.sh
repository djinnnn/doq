#!/bin/bash

# 定义一个函数来执行命令并记录时间
run_command() {
    start_time=$(date +%s) # 获取开始时间
    echo "Running command: $1"
    $1 # 执行命令
    end_time=$(date +%s) # 获取结束时间

    echo "Execution time for '$1': $((end_time - start_time)) seconds"
    echo "-------------------------------------"
}
#./doq [infile] [outfile] [certfile]
# 执行命令并记录时间
run_command "./doq -port 8853 ../../data/20240505/alive/8853.csv ../../data/20240505/results/port8853.csv ../../data/20240505/certs-1/8853.csv"
run_command "./doq -port 784 ../../data/20240505/alive/784.csv ../../data/20240505/results/port784.csv ../../data/20240505/certs-1/784.csv"
run_command "./doq -port 853 ../../data/20240505/alive/853.csv ../../data/20240505/results/port853.csv ../../data/20240505/certs-1/853.csv"
run_command "./doq -port 443 ../../data/20240505/alive/443.csv ../../data/20240505/results/port443.csv ../../data/20240505/certs-1/443.csv"
