#!/usr/bin/python3
# -*- coding: utf-8 -*-

"""This module transform properties into different format"""

def render_yaml(yaml_root, prefix=""):
    """render yaml"""
    result = ""
    if isinstance(yaml_root, dict):
        if prefix:
            result += "\n"
        # 兼容 Py3: 字典遍历建议排序以保证生成文件的一致性
        for key in sorted(yaml_root.keys()):
            result += "{}{}: {}".format(prefix, key, render_yaml(
                yaml_root[key], prefix + "   "))
    elif isinstance(yaml_root, list):
        result += "\n"
        for item in yaml_root:
            result += prefix + " - " + render_yaml(item, prefix + " ")
    else:
        result += "{}\n".format(yaml_root)
    return result


def to_yaml(content):
    """transform to yaml"""
    props = process_properties(content)
    keys = sorted(props.keys()) # 排序保证输出稳定
    yaml_props = {}
    for key in keys:
        parts = key.split(".")
        node = yaml_props
        prev_part = None
        parent_node = {}
        for part in parts[:-1]:
            if part.isdigit():
                idx = int(part)
                if isinstance(node, dict):
                    parent_node[prev_part] = []
                    node = parent_node[prev_part]
                while len(node) <= idx:
                    node.append({})
                parent_node = node
                node = node[idx] # 修正了原代码的 int(node) 错误
            else:
                if part not in node:
                    node[part] = {}
                parent_node = node
                node = node[part]
            prev_part = part
        
        last_part = parts[-1]
        if last_part.isdigit():
            idx = int(last_part)
            if isinstance(node, dict):
                parent_node[prev_part] = []
                node = parent_node[prev_part]
            node.append(props[key])
        else:
            node[last_part] = props[key]

    return render_yaml(yaml_props)

def to_yml(content):
    return to_yaml(content)

def to_properties(content):
    result = ""
    props = process_properties(content)
    for key, val in sorted(props.items()): # 增加 items() 并排序
        result += "{}: {}\n".format(key, val)
    return result

def to_env(content):
    result = ""
    props = process_properties(content)
    for key, val in sorted(props.items()): # 修正：增加 .items()
        result += "{}={}\n".format(key, val)
    return result

def to_sh(content):
    result = ""
    props = process_properties(content)
    for key, val in sorted(props.items()): # 修正：增加 .items()
        result += "export {}=\"{}\"\n".format(key, val)
    return result

def to_cfg(content):
    result = ""
    props = process_properties(content)
    for key, val in sorted(props.items()): # 修正：增加 .items()
        result += "{}={}\n".format(key, val)
    return result

def to_conf(content):
    result = ""
    props = process_properties(content)
    for key, val in sorted(props.items()): # 修正：增加 .items()
        result += "export {}={}\n".format(key, val)
    return result

def to_xml(content):
    result = "<configuration>\n"
    props = process_properties(content)
    for key in sorted(props.keys()):
        result += "<property><name>{0}</name><value>{1}</value></property>\n". \
          format(key, props[key])
    result += "</configuration>"
    return result

def process_properties(content, sep=': ', comment_char='#'):
    props = {}
    if not content:
        return props
    for line in content.split("\n"):
        sline = line.strip()
        if sline and not sline.startswith(comment_char):
            if sep in sline:
                key_value = sline.split(sep)
                key = key_value[0].strip()
                value = sep.join(key_value[1:]).strip().strip('"')
                props[key] = value
    return props