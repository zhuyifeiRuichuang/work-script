#!/bin/bash
# 先web界面或管理员配置一个只读用户reader
# 配置存储别名myminio，后面是web访问的登录信息。
mc alias set myminio http://172.16.0.19:9000 reader Shandong@123
