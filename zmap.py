import os
import subprocess
from datetime import datetime

# 输入文件和分割文件的数量
data_dir_path = '/home/lcl/wy/Projects/DoQ/data/20240701/ports_2'
input_file = '/home/lcl/wy/Projects/DoQ/data/20240701/ports_2/shard_6.csv'  # 你的输入文件
num_parts = 7
log_directory = "/home/lcl/wy/Projects/DoQ/data/20240701/ports_2/scan_logs"


if not os.path.exists(log_directory):
    os.makedirs(log_directory)

# 读取输入文件的所有行
with open(input_file, 'r') as file:
    lines = file.readlines()

# 计算每个小文件的行数
lines_per_part = len(lines) // num_parts
extra_lines = len(lines) % num_parts

# 分割输入文件
part_files = []
start = 0
for i in range(num_parts):
    end = start + lines_per_part + (1 if i < extra_lines else 0)
    part_file = f'part_{i+1}.txt'
    with open(part_file, 'w') as part:
        part.writelines(lines[start:end])
    part_files.append(part_file)
    start = end

# 执行扫描并记录日志
for part_file in part_files:
    segment_number = part_files.index(part_file) + 1
    log_file = os.path.join(log_directory, f"segment_{segment_number}_log.txt")
    output_file = f"output_segment_{segment_number}.txt"
    output_file_path = os.path.join(data_dir_path, output_file)
    with open(log_file, "w") as log:
        log.write(f"Starting scan for segment {segment_number} from file {part_file}\n")
        log.write(f"Scan started at {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")

        # 使用 zmap 执行扫描
        scan_command = f"sudo zmap -I {part_file} -M udp -p 853 --probe-args=file:initial_qscanner_1a1a1a1a.pkt -o {output_file_path} -B 5M"
        process = subprocess.Popen(scan_command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

        # 捕获和记录输出
        stdout, stderr = process.communicate()
        log.write(stdout.decode())
        log.write(stderr.decode())
        log.write(f"Scan completed at {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")

print("All scans completed.")
