#!/bin/bash

# 获取传入的参数，用于确定 shard 范围的起始值
START_SHARD=$1

# 定义常量
NUM_SHARDS=14
LOG_DIRECTORY="/home/lcl/wy/Projects/DoQ/data/20240813/scan_logs"

# 创建日志目录（如果不存在）
mkdir -p "$LOG_DIRECTORY"

# 执行14个分片的扫描
for ((i=0; i<NUM_SHARDS; i++)); do
    SHARD=$((START_SHARD * NUM_SHARDS + i))
    LOG_FILE="$LOG_DIRECTORY/segment_${SHARD}_log.txt"
    OUTPUT_FILE="/home/lcl/wy/Projects/DoQ/data/20240813/ports/20240813_${SHARD}_ports.csv"
    
    echo "Starting scan for shard $SHARD" > "$LOG_FILE"
    echo "Scan started at $(date '+%Y-%m-%d %H:%M:%S')" >> "$LOG_FILE"
    echo "Starting scan for shard $SHARD"

    # 执行 zmap 扫描
    SCAN_COMMAND="sudo zmap --seed 20240813 --shards 84 --shard $SHARD -M udp -p 853 --probe-args=file:initial_qscanner_1a1a1a1a.pkt 0.0.0.0/0 -o $OUTPUT_FILE -B 5M"
    eval "$SCAN_COMMAND" 2>&1 | tail -n 10 -f >> "$LOG_FILE" &

    wait $!

    echo "Scan completed at $(date '+%Y-%m-%d %H:%M:%S')" >> "$LOG_FILE"
done

echo "All scans completed."
