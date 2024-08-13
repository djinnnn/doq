#!/bin/bash

# 定义变量
echo "Starting read input files"
DATA_DIR_PATH='/home/lcl/wy/Projects/DoQ/data/20240701/ports_2'
INPUT_FILE='/home/lcl/wy/Projects/DoQ/data/20240701/ports_2/shard_6.csv'
#INPUT_FILE='/home/lcl/wy/Projects/DoQ/data/20240701/ports_2/test.csv'
NUM_PARTS=7
LOG_DIRECTORY="$DATA_DIR_PATH/scan_logs"


# 创建日志目录（如果不存在）
mkdir -p "$LOG_DIRECTORY"

# 分割输入文件为NUM_PARTS个小文件
split -n l/7  -d --additional-suffix=.txt "$INPUT_FILE" "$DATA_DIR_PATH/part_"

# 获取分割文件列表
PART_FILES=($(ls $DATA_DIR_PATH/part_*.txt))
echo "Start Scanning"
# 遍历每个分割文件执行扫描
for PART_FILE in "${PART_FILES[@]}"; do
    SEGMENT_NUMBER=$(echo "$PART_FILE" | grep -oP '(?<=part_)\d+(?=.txt)')
    echo "$PART_FILE"
    echo "$SEGMENT_NUMBER"
    LOG_FILE="$LOG_DIRECTORY/segment_${SEGMENT_NUMBER}_log.txt"
    OUTPUT_FILE="output_segment_${SEGMENT_NUMBER}.txt"
    OUTPUT_FILE_PATH="$DATA_DIR_PATH/$OUTPUT_FILE"

    echo "Starting scan for segment $SEGMENT_NUMBER from file $PART_FILE" > "$LOG_FILE"
    echo "Scan started at $(date '+%Y-%m-%d %H:%M:%S')" >> "$LOG_FILE"
    echo "Starting scan for segment $SEGMENT_NUMBER from file $PART_FILE"
    
    # 使用 zmap 执行扫描并将输出重定向到临时文件
    SCAN_COMMAND="sudo zmap -I $PART_FILE -M udp -p 853 --probe-args=file:initial_qscanner_1a1a1a1a.pkt -o $OUTPUT_FILE_PATH -B 5M"
    eval "$SCAN_COMMAND" 2>&1 | tail -n 10 -f >> "$LOG_FILE" &

    wait $!

    echo "Scan completed at $(date '+%Y-%m-%d %H:%M:%S')" >> "$LOG_FILE"
done

echo "All scans completed."
