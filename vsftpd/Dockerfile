FROM ubuntu:24.04

# 安装依赖
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc make libc-dev libpam0g-dev passwd libpam-pwdfile \
    && rm -rf /var/lib/apt/lists/*

# 复制并编译安装vsftpd
COPY vsftpd-*.tar.gz /tmp/
RUN cd /tmp && \
    tar -zxf vsftpd-*.tar.gz && \
    cd vsftpd-* && \
    mkdir -p /usr/local/man/man8 /usr/local/man/man5 && \
    make && make install && \
    cp vsftpd.conf /etc/ && \
    cd /tmp && rm -rf vsftpd-* && \
    apt-get purge -y gcc make && apt-get autoremove -y && apt-get clean

# ===================== 修复1：创建标准ftp用户 + 强制设置密码 =====================
RUN mkdir -p /usr/share/empty && \
    mkdir -p /data/ && \
    userdel ftp 2>/dev/null || true && \
    useradd -r -d /data -s /usr/sbin/nologin ftp && \
    # 内置账户密码
    echo 'ftp:Ftp@123' | chpasswd && \
	chown ftp:ftp /data/ 

# ===================== Ubuntu专用PAM认证=====================
RUN cat > /etc/pam.d/vsftpd << EOF
auth    required    pam_unix.so
account required    pam_unix.so
EOF
RUN cat > /etc/pam.d/ftp << EOF
auth    required    pam_unix.so
account required    pam_unix.so
EOF

# 日志初始化
RUN touch /var/log/vsftpd.log && chmod 644 /var/log/vsftpd.log

# 启动
CMD ["/bin/sh", "-c", "/usr/local/sbin/vsftpd /etc/vsftpd.conf & tail -f /var/log/vsftpd.log"]
