#!/bin/bash

# 定义变量
DATA_DIR_PATH='/home/lcl/wy/Projects/DoQ/data/20240701/ports_2'
INPUT_FILE='/home/lcl/wy/Projects/DoQ/data/20240701/ports_2/shard_7.csv'
NUM_PARTS=7
LOG_DIRECTORY="$DATA_DIR_PATH/scan_logs"

# 创建日志目录（如果不存在）
mkdir -p "$LOG_DIRECTORY"

# 分割输入文件为NUM_PARTS个小文件
split -l/7  -d --additional-suffix=.txt "$INPUT_FILE" "$DATA_DIR_PATH/part_"

# 获取分割文件列表
PART_FILES=($(ls $DATA_DIR_PATH/part_*.txt))

# 遍历每个分割文件执行扫描
for PART_FILE in "${PART_FILES[@]}"; do
    SEGMENT_NUMBER=$(echo "$PART_FILE" | grep -o '[0-9]*')
    LOG_FILE="$LOG_DIRECTORY/segment_${SEGMENT_NUMBER}_log.txt"
    OUTPUT_FILE="output_segment_${SEGMENT_NUMBER}.txt"
    OUTPUT_FILE_PATH="$DATA_DIR_PATH/$OUTPUT_FILE"

    echo "Starting scan for segment $SEGMENT_NUMBER from file $PART_FILE" > "$LOG_FILE"
    echo "Scan started at $(date '+%Y-%m-%d %H:%M:%S')" >> "$LOG_FILE"

    # 使用 zmap 执行扫描
    SCAN_COMMAND="sudo zmap -I $PART_FILE -M udp -p 853 --probe-args=file:initial_qscanner_1a1a1a1a.pkt -o $OUTPUT_FILE_PATH -B 5M"
    eval "$SCAN_COMMAND" >> "$LOG_FILE" 2>&1

    echo "Scan completed at $(date '+%Y-%m-%d %H:%M:%S')" >> "$LOG_FILE"
done

echo "All scans completed."
