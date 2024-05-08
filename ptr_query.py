#ptr_query.py
import dns.resolver
import dns.reversename
from concurrent.futures import ThreadPoolExecutor
import pandas as pd
import time
import argparse

def parse_arguments():
    parser = argparse.ArgumentParser(description='Perform PTR queries on a list of IP addresses')
    parser.add_argument('input_file', type=str, help='Path to the input CSV file containing IP addresses')
    parser.add_argument('output_file', type=str, help='Path to the output CSV file to save PTR query results')
    return parser.parse_args()

# 要查询的IP地址列表
#ips = ['8.8.8.8', '8.8.4.4', '1.1.1.1', '192.0.2.1', '93.184.216.34']
def read_ips_from_csv(file_path):
    df = pd.read_csv(file_path, header=None)
    ip_column = df.iloc[:, 0]
    return ip_column.tolist()

# 查询函数
def query_ptr(ip):
    resolver = dns.resolver.Resolver()
    try:
        # 获取IP地址的反向DNS名称
        reverse_name = dns.reversename.from_address(ip)
        # 解析PTR记录
        answers = resolver.resolve(reverse_name, 'PTR')
        return ip, [answer.to_text() for answer in answers]
    except Exception as e:
        return ip, str(e)

# 使用 ThreadPoolExecutor 来并行执行查询
def perform_queries(ips, max_workers=16):
    with ThreadPoolExecutor(max_workers=max_workers) as executor:
        results = executor.map(query_ptr, ips)
        return list(results)

#read ip list
args = parse_arguments()
#ips = read_ips_from_csv('20240418_alive_server_443_port.csv')
# ips = read_ips_from_csv('test.csv')
ips = read_ips_from_csv(args.input_file)

# 执行查询并打印结果
start = time.time()
results = perform_queries(ips)
for result in results:
    print(f"{result[0]}: {result[1]}")
    
end = time.time()
df=pd.DataFrame(results, columns=['IP', 'PTR'])
df.to_csv(args.output_file, index=False)
print("Done! Spend ", end-start, " seconds.")
