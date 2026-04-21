#!/usr/bin/python3
# -*- coding: utf-8 -*-

"""convert environment variables to config"""

import os
import re
import argparse
import sys
import transformation

class Simple(object):
    """Simple conversion"""
    def __init__(self, args):
        parser = argparse.ArgumentParser()
        parser.add_argument("--destination", help="Destination directory", required=True)
        self.args = parser.parse_args(args=args)

        self.known_formats = ['xml', 'properties', 'yaml', 'yml', 'env', "sh", "cfg", 'conf']
        self.output_dir = self.args.destination
        self.excluded_envs = ['HADOOP_CONF_DIR']
        self.configurables = {}

    def destination_file_path(self, name, extension):
        """destination file path"""
        return os.path.join(self.output_dir, "{}.{}".format(name, extension))

    def write_env_var(self, name, extension, key, value):
        """Write environment variables"""
        # 显式指定 utf-8 编码
        file_path = self.destination_file_path(name, extension) + ".raw"
        with open(file_path, "a", encoding='utf-8') as myfile:
            myfile.write("{}: {}\n".format(key, value))

    def process_envs(self):
        """Process environment variables"""
        # 对环境变量进行排序，保证执行的可预测性
        for key in sorted(os.environ.keys()):
            if key in self.excluded_envs:
                continue
            
            pattern = re.compile("[_\\.]")
            parts = pattern.split(key)
            if not parts:
                continue
                
            extension = None
            name = parts[0].lower()
            
            if len(parts) > 1:
                extension = parts[1].lower()
                # 默认配置 key 截取
                config_key = key[len(name) + len(extension) + 2:].strip()
            
            if extension and "!" in extension:
                splitted = extension.split("!")
                extension = splitted[0]
                fmt = splitted[1]
                config_key = key[len(name) + len(extension) + len(fmt) + 3:].strip()
            else:
                fmt = extension

            if extension and extension in self.known_formats:
                if name not in self.configurables:
                    # 初始化文件
                    with open(self.destination_file_path(name, extension) + ".raw", "w", encoding='utf-8') as myfile:
                        myfile.write("")
                self.configurables[name] = (extension, fmt)
                self.write_env_var(name, extension, config_key, os.environ[key])
            else:
                # 修复逻辑：处理不带 format 标记但匹配已定义 configurable 的变量
                for configurable_name in self.configurables:
                    if key.lower().startswith(configurable_name.lower()):
                        ext, _ = self.configurables[configurable_name] # 修正：只提取 extension
                        self.write_env_var(configurable_name,
                                           ext,
                                           key[len(configurable_name) + 1:],
                                           os.environ[key])

    def transform(self):
        """transform"""
        for configurable_name in sorted(self.configurables.keys()):
            name = configurable_name
            extension, fmt = self.configurables[name]

            destination_path = self.destination_file_path(name, extension)

            if not os.path.exists(destination_path + ".raw"):
                continue

            with open(destination_path + ".raw", "r", encoding='utf-8') as myfile:
                content = myfile.read()
                
            # 调用 transformation.py 中的函数
            try:
                transformer_func = getattr(transformation, "to_" + fmt)
                content = transformer_func(content)
                with open(destination_path, "w", encoding='utf-8') as myfile:
                    myfile.write(content)
            except AttributeError:
                print("Error: No transformer found for format '{}'".format(fmt), file=sys.stderr)

    def main(self):
        self.process_envs()
        self.transform()

def main():
    Simple(sys.argv[1:]).main()

if __name__ == '__main__':
    main()