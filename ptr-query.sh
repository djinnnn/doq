#ptr-query.sh
#记得修改参数
#!/bin/bash
echo "Start.."
mkdir -p /home/ubuntu/wy/Projects/DoQ/data/20240505/ptr_results
# 输入文件列表
input_files=(
    "/home/ubuntu/wy/Projects/DoQ/data/20240505/alive/443.csv"
    "/home/ubuntu/wy/Projects/DoQ/data/20240505/alive/784.csv"
    "/home/ubuntu/wy/Projects/DoQ/data/20240505/alive/853.csv"
    "/home/ubuntu/wy/Projects/DoQ/data/20240505/alive/8853.csv"
)

# 循环处理每个输入文件
for input_file in "${input_files[@]}"; do
    # 提取文件名（不带路径和扩展名）
    file_name=$(basename -- "$input_file")
    file_name_without_extension="${file_name%.*}"

    # 输出文件路径
    output_file="/home/ubuntu/wy/Projects/DoQ/data/20240505/ptr_results/${file_name_without_extension}_ptr.csv"

    # 调用 Python 脚本进行查询
    python3 ptr_query.py "$input_file" "$output_file"
done
