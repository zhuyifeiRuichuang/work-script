import importlib
import subprocess
import sys
import os
import time
import threading
import datetime
import random
import string
import configparser
import argparse
from minio import Minio
from minio.error import S3Error

# 获取脚本所在目录的绝对路径
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
# 日志目录路径（脚本所在目录下的log子目录）
LOG_DIR = os.path.join(SCRIPT_DIR, 'log')

# 确保日志目录存在，不存在则创建
def ensure_log_dir():
    if not os.path.exists(LOG_DIR):
        try:
            os.makedirs(LOG_DIR, exist_ok=True)  # exist_ok=True 避免目录已存在时报错
            print(f"已创建日志目录: {LOG_DIR}")
        except Exception as e:
            print(f"创建日志目录失败: {str(e)}")
            sys.exit(1)

# 生成日志文件名（包含完整路径）
def generate_log_filename():
    ensure_log_dir()  # 确保日志目录存在
    date_str = datetime.datetime.now().strftime("%Y%m%d%H%M%S")
    random_str = ''.join(random.choices(string.ascii_letters + string.digits, k=6))
    # 日志文件路径 = 日志目录 + 文件名
    return os.path.join(LOG_DIR, f"{date_str}_{random_str}_upload_minio.log")

# 全局日志文件路径
log_file = generate_log_filename()

# 写入日志
def log(message):
    timestamp = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    with open(log_file, 'a', encoding='utf-8') as f:
        f.write(f"[{timestamp}] {message}\n")

# 安全解析布尔值的函数
def safe_parse_boolean(value):
    """解析布尔值，支持多种格式"""
    cleaned = str(value).strip().strip('"').strip("'").lower()
    if cleaned in ['true', '1']:
        return True
    elif cleaned in ['false', '0']:
        return False
    else:
        raise ValueError(f"无法解析为布尔值: {value}（有效值应为true/false/1/0，可带引号）")

# 处理配置文件中的注释（保留引号内的#）
def remove_comments(line):
    """
    移除行尾注释，但保留引号内的#符号
    例如：
    'key = "value#1" # 这是注释' → 'key = "value#1"'
    'key = value; 这是注释' → 'key = value'
    """
    in_quote = None  # 记录当前是否在引号内，值为"或'，None表示不在引号内
    for i, char in enumerate(line):
        # 处理引号
        if char in ['"', "'"]:
            if in_quote == char:
                in_quote = None  # 退出引号
            elif in_quote is None:
                in_quote = char  # 进入引号
        # 处理注释符号（不在引号内时）
        elif char in ['#', ';'] and in_quote is None:
            return line[:i].rstrip()  # 返回注释前的内容（去除尾部空格）
    return line.rstrip()  # 没有注释时返回整行（去除尾部空格）

# 任务状态常量
STATUS_WAITING = "等待中"
STATUS_RUNNING = "执行中"
STATUS_SUCCESS = "成功"
STATUS_FAILED = "失败"

# 任务类
class UploadTask:
    def __init__(self, task_name, task_type, params):
        self.task_name = task_name
        self.task_type = task_type
        self.params = params
        self.total_files = 0
        self.uploaded_files = 0
        self.status = STATUS_WAITING
        self.error_msg = ""
        self.lock = threading.Lock()

    def update_status(self, status, error_msg=""):
        with self.lock:
            self.status = status
            if error_msg:
                self.error_msg = error_msg

    def update_progress(self, uploaded, total=None):
        with self.lock:
            self.uploaded_files = uploaded
            if total is not None:
                self.total_files = total

# 全局变量
tasks = []
task_threads = []
progress_complete = False
config_info = ""  # 新增：保存配置文件信息，用于持久显示

def update_console_display():
    """动态刷新控制台显示，保留配置文件信息"""
    global tasks, progress_complete, config_info
    sys.stdout.write("\033[H\033[J")  # 清屏
    
    # 持久显示配置文件信息和日志路径（每次刷新都重新打印）
    print(config_info)
    print(f"详细日志: {log_file}\n")
    print(f"任务总数: {len(tasks)} | 状态: {'运行中' if not progress_complete else '已完成'}\n")
    
    # 表头
    print(f"{'任务名称':<10} {'类型':<6} {'进度':<25} {'状态':<8} 详情")
    print("-" * 80)
    
    # 任务行
    for task in tasks:
        with task.lock:
            if task.task_type == "file":
                progress = 100 if task.uploaded_files > 0 else 0
                bar = "=" * 20 if progress == 100 else " " * 20
                progress_str = f"[{bar}] {progress}%"
                detail = task.params["local_file"]
            else:
                if task.total_files == 0:
                    progress = 0
                    bar = " " * 20
                else:
                    progress = int((task.uploaded_files / task.total_files) * 100)
                    bar_length = int(20 * progress / 100)
                    bar = "=" * bar_length + " " * (20 - bar_length)
                progress_str = f"[{bar}] {progress}% ({task.uploaded_files}/{task.total_files})"
                detail = task.params["local_dir"]
            
            # 状态颜色
            if task.status == STATUS_SUCCESS:
                status_str = "\033[92m成功\033[0m"
            elif task.status == STATUS_FAILED:
                status_str = f"\033[91m失败\033[0m"
            elif task.status == STATUS_RUNNING:
                status_str = "\033[93m执行中\033[0m"
            else:
                status_str = STATUS_WAITING
            
            if task.status == STATUS_FAILED:
                detail += f" (错误: {task.error_msg[:30]}...)"
            
            print(f"{task.task_name:<10} {task.task_type:<6} {progress_str:<25} {status_str:<8} {detail}")
    
    sys.stdout.flush()

def get_package_manager():
    """检测操作系统包管理器"""
    log("开始检测操作系统包管理器")
    if os.path.exists('/etc/redhat-release') or os.path.exists('/etc/centos-release'):
        log("检测到RedHat/CentOS系统，使用yum包管理器")
        return 'yum'
    elif os.path.exists('/etc/debian_version') or os.path.exists('/etc/lsb-release'):
        log("检测到Debian/Ubuntu系统，使用apt包管理器")
        return 'apt'
    else:
        log("无法识别操作系统类型")
        return None

def check_and_install_pip():
    """检查并安装pip"""
    log("开始检查pip是否已安装")
    try:
        importlib.import_module('pip')
        log("pip已安装")
        return True
    except ImportError:
        log("未找到pip，准备安装...")
    
    pkg_manager = get_package_manager()
    if not pkg_manager:
        log("无法识别操作系统，无法安装pip")
        return False
    
    try:
        log(f"使用{pkg_manager}执行更新操作")
        if pkg_manager == 'yum':
            result = subprocess.run(['sudo', 'yum', 'update', '-y'], capture_output=True, text=True)
            log(f"yum update stdout: {result.stdout[:500]}...")
            log(f"yum update stderr: {result.stderr}")
            if result.returncode != 0:
                log(f"yum update 失败，返回码: {result.returncode}")
                return False
                
            log("开始安装python3-pip")
            result = subprocess.run(['sudo', 'yum', 'install', 'python3-pip', '-y'], capture_output=True, text=True)
            log(f"yum install stdout: {result.stdout[:500]}...")
            log(f"yum install stderr: {result.stderr}")
            if result.returncode != 0:
                return False
        else:
            result = subprocess.run(['sudo', 'apt', 'update', '-y'], capture_output=True, text=True)
            log(f"apt update stdout: {result.stdout[:500]}...")
            log(f"apt update stderr: {result.stderr}")
            if result.returncode != 0:
                return False
                
            log("开始安装python3-pip")
            result = subprocess.run(['sudo', 'apt', 'install', 'python3-pip', '-y'], capture_output=True, text=True)
            log(f"apt install stdout: {result.stdout[:500]}...")
            log(f"apt install stderr: {result.stderr}")
            if result.returncode != 0:
                return False
        
        importlib.import_module('pip')
        log("pip安装成功并验证通过")
        return True
    except Exception as e:
        log(f"安装pip时发生异常: {str(e)}")
        return False

def check_and_install_minio():
    """检查并安装minio模块"""
    log("开始检查minio模块是否已安装")
    try:
        importlib.import_module('minio')
        log("minio模块已安装")
        return True
    except ImportError:
        log("未找到minio模块，开始安装...")
        try:
            result = subprocess.run([sys.executable, "-m", "pip", "install", "minio"], capture_output=True, text=True)
            log(f"pip install stdout: {result.stdout[:500]}...")
            log(f"pip install stderr: {result.stderr}")
            if result.returncode != 0:
                return False
            log("minio模块安装完成")
            return True
        except Exception as e:
            log(f"安装minio模块时发生异常: {str(e)}")
            return False

def find_first_conf_file():
    """查找当前目录第一个.conf文件"""
    current_dir = os.getcwd()
    log(f"在当前目录({current_dir})查找.conf文件")
    conf_files = []
    for item in os.listdir(current_dir):
        item_path = os.path.join(current_dir, item)
        if os.path.isfile(item_path) and item.lower().endswith('.conf'):
            conf_files.append(item)
    
    if not conf_files:
        return None
    conf_files.sort()
    return conf_files[0]

def read_config(config_file):
    """读取INI配置文件（增强注释处理和布尔值解析）"""
    log(f"开始读取配置文件: {config_file}")
    if not os.path.exists(config_file):
        raise Exception(f"配置文件不存在: {config_file}")
    
    # 读取并预处理配置文件（移除注释）
    processed_lines = []
    with open(config_file, 'r', encoding='utf-8') as f:
        for line_num, line in enumerate(f, 1):
            original_line = line.strip()
            # 跳过空行和纯注释行
            if not original_line:
                continue
            if original_line.startswith(('#', ';')):
                continue
            # 处理包含注释的行
            processed_line = remove_comments(line)
            if processed_line:  # 只保留非空行
                processed_lines.append(processed_line)
    
    # 检查是否有节头
    if not any('[' in line and ']' in line for line in processed_lines):
        raise Exception(
            f"配置文件格式错误：缺少节头（如[minio]、[file1]）\n"
            f"请使用INI格式，示例：\n"
            f"[minio]\n"
            f"endpoint = \"172.16.0.19:9000\"\n"
            f"access_key = \"your_key\"\n"
        )
    
    # 使用处理后的内容创建配置解析器
    config = configparser.ConfigParser()
    try:
        # 将处理后的行转换为字符串供configparser解析
        config_content = '\n'.join(processed_lines)
        config.read_string(config_content)
    except Exception as e:
        raise Exception(
            f"解析配置文件失败：{str(e)}\n"
            f"请检查是否有以下问题：\n"
            f"1. 所有配置必须放在节头下（如[minio]、[file1]）\n"
            f"2. 节头格式为[节名]（如[minio]）\n"
            f"3. 键值对格式为key = value（等号前后可空格）"
        )
    
    # 检查是否有节
    if not config.sections():
        raise Exception("配置文件中未找到任何节（如[minio]、[file1]），请检查格式")
    
    # 解析MinIO配置
    if not config.has_section('minio'):
        raise Exception("配置文件缺少[minio]节（MinIO连接信息）")
    
    # 解析secure配置（增强布尔值处理）
    try:
        secure = safe_parse_boolean(config.get('minio', 'secure', fallback='false'))
    except ValueError as e:
        raise Exception(f"[minio]节中secure配置错误：{str(e)}")
    
    minio_config = {
        'endpoint': config.get('minio', 'endpoint', fallback='').strip().strip('"'),
        'access_key': config.get('minio', 'access_key', fallback='').strip().strip('"'),
        'secret_key': config.get('minio', 'secret_key', fallback='').strip().strip('"'),
        'secure': secure,
        'bucket_name': config.get('minio', 'bucket_name', fallback='').strip().strip('"')
    }
    
    # 验证MinIO必要配置
    required_minio_keys = ['endpoint', 'access_key', 'secret_key', 'bucket_name']
    for key in required_minio_keys:
        if not minio_config[key]:
            raise Exception(f"[minio]节缺少必要配置: {key}（请检查是否拼写正确）")
    log(f"MinIO配置解析完成: endpoint={minio_config['endpoint']}, bucket={minio_config['bucket_name']}, secure={minio_config['secure']}")
    
    # 解析文件和目录任务
    file_tasks = []
    dir_tasks = []
    for section in config.sections():
        if section == 'minio':
            continue
        
        if section.lower().startswith('file'):
            # 检查文件任务必要配置
            if not config.has_option(section, 'local_file'):
                raise Exception(f"任务[{section}]缺少local_file配置（必须指定要上传的文件路径）")
            task_params = {
                'local_file': config.get(section, 'local_file').strip().strip('"'),
                'object_name': config.get(section, 'object_name', fallback='').strip().strip('"')
            }
            file_tasks.append((section, task_params))
            log(f"解析文件任务 [{section}]: {task_params['local_file']}")
        
        elif section.lower().startswith('dir'):
            # 检查目录任务必要配置
            if not config.has_option(section, 'local_dir'):
                raise Exception(f"任务[{section}]缺少local_dir配置（必须指定要上传的目录路径）")
            task_params = {
                'local_dir': config.get(section, 'local_dir').strip().strip('"'),
                'remote_dir': config.get(section, 'remote_dir', fallback='').strip().strip('"')
            }
            dir_tasks.append((section, task_params))
            log(f"解析目录任务 [{section}]: {task_params['local_dir']}")
    
    if not file_tasks and not dir_tasks:
        raise Exception(
            "未找到任何上传任务，请检查配置文件\n"
            "文件任务节名需以file开头（如[file1]），目录任务需以dir开头（如[dir1]）"
        )
    
    return minio_config, file_tasks, dir_tasks

def get_all_files_in_dir(local_dir):
    """获取目录下所有文件"""
    if not os.path.isdir(local_dir):
        raise Exception(f"目录不存在或不是有效目录: {local_dir}")
    
    file_paths = []
    for root, _, files in os.walk(local_dir):
        for file in files:
            file_paths.append(os.path.join(root, file))
    return file_paths

def run_file_task(task, minio_client, bucket_name):
    """执行文件上传任务"""
    try:
        task.update_status(STATUS_RUNNING)
        local_file = task.params['local_file']
        object_name = task.params['object_name']
        
        if not os.path.isfile(local_file):
            raise Exception(f"文件不存在: {local_file}")
        
        if not object_name:
            object_name = os.path.basename(local_file)
        
        minio_client.fput_object(bucket_name, object_name, local_file)
        task.update_progress(1, 1)
        task.update_status(STATUS_SUCCESS)
        log(f"任务[{task.task_name}]成功: {local_file} -> {object_name}")
    except Exception as e:
        error_msg = str(e)
        task.update_status(STATUS_FAILED, error_msg)
        log(f"任务[{task.task_name}]失败: {error_msg}")

def run_dir_task(task, minio_client, bucket_name):
    """执行目录上传任务"""
    try:
        task.update_status(STATUS_RUNNING)
        local_dir = task.params['local_dir']
        remote_dir = task.params['remote_dir']
        
        all_files = get_all_files_in_dir(local_dir)
        total = len(all_files)
        task.update_progress(0, total)
        if total == 0:
            log(f"任务[{task.task_name}]警告: 目录为空: {local_dir}")
            task.update_status(STATUS_SUCCESS)
            return
        
        for i, local_file in enumerate(all_files, 1):
            relative_path = os.path.relpath(local_file, local_dir)
            if remote_dir:
                object_name = f"{remote_dir}/{relative_path}".replace(os.sep, '/')
            else:
                object_name = relative_path.replace(os.sep, '/')
            
            minio_client.fput_object(bucket_name, object_name, local_file)
            task.update_progress(i, total)
            log(f"任务[{task.task_name}]进度: {i}/{total}")
        
        task.update_status(STATUS_SUCCESS)
        log(f"任务[{task.task_name}]成功: 共上传{total}个文件")
    except Exception as e:
        error_msg = str(e)
        task.update_status(STATUS_FAILED, error_msg)
        log(f"任务[{task.task_name}]失败: {error_msg}")

def main():
    global tasks, task_threads, progress_complete, config_info  # 引用全局配置信息变量
    
    # 解析命令行参数
    parser = argparse.ArgumentParser(description='MinIO多文件/目录上传工具（日志目录优化）')
    parser.add_argument('-f', '--file', help='指定配置文件路径（默认查找当前目录第一个.conf文件）')
    args = parser.parse_args()

    # 确定配置文件
    config_file = args.file
    if not config_file:
        # 自动查找当前目录第一个.conf文件
        config_file = find_first_conf_file()
        if not config_file:
            print("错误: 未指定配置文件，且当前目录未找到任何.conf文件")
            sys.exit(1)
        config_info = f"未指定配置文件，自动使用当前目录第一个.conf文件: {os.path.abspath(config_file)}\n"
    else:
        # 检查指定的配置文件是否存在
        if not os.path.exists(config_file):
            print(f"错误: 指定的配置文件不存在: {config_file}")
            sys.exit(1)
        config_info = f"使用指定的配置文件: {os.path.abspath(config_file)}\n"
    
    # 记录日志信息（不再在这里打印，改为在刷新函数中显示）
    log("===== 脚本开始执行 =====")
    log(f"使用配置文件: {os.path.abspath(config_file)}")
    
    try:
        # 检查系统环境
        log("===== 检查系统环境 =====")
        pkg_manager = get_package_manager()
        if not pkg_manager:
            raise Exception("不支持的操作系统（无法识别包管理器）")
        
        # 检查pip
        log("===== 检查pip =====")
        if not check_and_install_pip():
            raise Exception("pip安装失败")
        
        # 检查minio模块
        log("===== 检查minio模块 =====")
        if not check_and_install_minio():
            raise Exception("minio模块安装失败")
        
        # 读取配置文件
        log("===== 读取配置文件 =====")
        minio_config, file_tasks, dir_tasks = read_config(config_file)
        
        # 连接MinIO
        log("===== 连接MinIO =====")
        minio_client = Minio(
            endpoint=minio_config['endpoint'],
            access_key=minio_config['access_key'],
            secret_key=minio_config['secret_key'],
            secure=minio_config['secure']
        )
        bucket_name = minio_config['bucket_name']
        if not minio_client.bucket_exists(bucket_name):
            minio_client.make_bucket(bucket_name)
            log(f"创建新桶: {bucket_name}")
        else:
            log(f"使用现有桶: {bucket_name}")
        
        # 初始化任务
        log("===== 初始化任务 =====")
        for task_name, params in file_tasks:
            tasks.append(UploadTask(task_name, "file", params))
        for task_name, params in dir_tasks:
            tasks.append(UploadTask(task_name, "dir", params))
        
        # 启动任务线程
        log("===== 启动上传任务 =====")
        for task in tasks:
            if task.task_type == "file":
                thread = threading.Thread(target=run_file_task, args=(task, minio_client, bucket_name))
            else:
                thread = threading.Thread(target=run_dir_task, args=(task, minio_client, bucket_name))
            task_threads.append(thread)
            thread.start()
        
        # 刷新显示
        while not progress_complete:
            update_console_display()
            all_complete = all(task.status in [STATUS_SUCCESS, STATUS_FAILED] for task in tasks)
            if all_complete:
                progress_complete = True
            time.sleep(0.5)
        
        update_console_display()
        print("\n所有任务执行完毕！")
        log("===== 所有任务执行完毕 =====")

    except Exception as e:
        progress_complete = True
        error_msg = f"初始化失败: {str(e)}"
        log(error_msg)
        print(f"\n错误: {error_msg}")
        log("===== 脚本异常退出 =====")
        sys.exit(1)

if __name__ == "__main__":
    main()
