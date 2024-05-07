import pandas as pd
'''
def save_valid_ptrs:
    valid_ptr_df = df.loc[df['Valid_PTR'] == True, ['IP', 'PTR']]
    print(valid_ptr_df)
    valid_ptr_df.to_csv('valid_ptr_records.csv', index=False)
    print("已保存Valid_PTR为True的数据到 'valid_ptr_records.csv'")
'''
import numpy as np

def save_valid_ptrs(df, output_filename='valid_ptr_records.csv'):
    # 复制 DataFrame 防止修改原始数据
    df_copy = df.copy()
    
   # 如果 Valid_PTR 为 False，则将 PTR 列设置为 IP 列的值
    df_copy['PTR'] = np.where(
    df_copy['Valid_PTR'] == False,
    df_copy['IP'],
    df_copy['PTR'].apply(lambda x: x.strip("[]'").split(',')[0].strip())
)

    
   # 将 PTR 列中的列表转换为字符串
    df_copy['PTR'] = df_copy['PTR'].apply(lambda x: x[0] if isinstance(x, list) else str(x))
   
   # 将这些行的 Valid_PTR 设置为 True
    df_copy['Valid_PTR'] = np.where(df_copy['Valid_PTR'] == False, True, df_copy['Valid_PTR'])
    
    # 筛选出 Valid_PTR 为 True 的记录
    valid_ptr_df = df_copy[df_copy['Valid_PTR'] == True][['IP', 'PTR']]
    
    # 打印出有效的 PTR 记录，用于验证或调试
    print(valid_ptr_df)
    
    # 将这些记录保存到 CSV 文件
    valid_ptr_df.to_csv(output_filename, index=False, header=False)
    
    # 打印一个确认消息
    print(f"已保存Valid_PTR为True的数据到 '{output_filename}'")


# 示例使用方法
# df = pd.read_csv('your_input_file.csv')  # 假设你已经加载了一个包含 'Valid_PTR', 'IP', 'PTR' 列的 DataFrame
# save_valid_ptrs(df, 'custom_filename.csv')  # 调用函数并指定输出文件名



def check_valid_ptr_records(df):     
    df['First Octet'] = df['IP'].str.split('.').str[0]
    df['Last Octet'] = df['IP'].str.split('.').str[-1]

    df['Valid_PTR'] = df.apply(lambda row: (row['First Octet'] not in row['PTR'] and
                                        'resolution' not in row['PTR'] and
                                        row['Last Octet'] not in row['PTR'] and
                                        'Resolver' not in row['PTR']
                                        ), axis=1)




file_path = [
    '/home/ubuntu/wy/Projects/DoQ/data/20240505/ptr_results/443_ptr.csv',
    '/home/ubuntu/wy/Projects/DoQ/data/20240505/ptr_results/784_ptr.csv',
    '/home/ubuntu/wy/Projects/DoQ/data/20240505/ptr_results/853_ptr.csv',
    '/home/ubuntu/wy/Projects/DoQ/data/20240505/ptr_results/8853_ptr.csv'
]

output_filepath = [
    '/home/ubuntu/wy/Projects/DoQ/data/20240505/ptr_results/443.csv',
    '/home/ubuntu/wy/Projects/DoQ/data/20240505/ptr_results/784.csv',
    '/home/ubuntu/wy/Projects/DoQ/data/20240505/ptr_results/853.csv',
    '/home/ubuntu/wy/Projects/DoQ/data/20240505/ptr_results/8853.csv'
]

#顺序为：443，784，853，8853
dfs = []

for file in file_path:
    df = pd.read_csv(file)
    dfs.append(df)
    
for df in dfs:
    count_not_exist = df['PTR'].str.contains('not exist', na=False).sum()
    print(f'The number of rows with "not exist" in the PTR column: {count_not_exist}')
    
    #df.to_csv('ptr.csv')

for df in dfs:
    check_valid_ptr_records(df)
    
for df in dfs:
    print(df)

for i in range(4):
    save_valid_ptrs(dfs[i], output_filepath[i])
    

