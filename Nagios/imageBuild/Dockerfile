FROM ubuntu:24.04

ENV NAGIOS_HOME=/opt/nagios
ENV NAGIOS_USER=nagios
ENV NAGIOS_GROUP=nagios
ENV NAGIOS_CMDUSER=nagios
ENV NAGIOS_CMDGROUP=nagios
ENV NAGIOS_FQDN=nagios.example.com
ENV NAGIOSADMIN_USER=nagios
ENV NAGIOSADMIN_PASS=nagios
ENV APACHE_RUN_USER=nagios
ENV APACHE_RUN_GROUP=nagios
ENV NAGIOS_TIMEZONE=UTC
ENV DEBIAN_FRONTEND=noninteractive
ENV NG_NAGIOS_CONFIG_FILE=${NAGIOS_HOME}/etc/nagios.cfg
ENV NG_CGI_DIR=${NAGIOS_HOME}/sbin
ENV NG_WWW_DIR=${NAGIOS_HOME}/share/nagiosgraph
ENV NG_CGI_URL=/cgi-bin

# 可指定版本。默认当前最新版
# 版本查询：https://github.com/NagiosEnterprises/nagioscore
ENV NAGIOS_BRANCH=master
# 版本查询：https://github.com/nagios-plugins/nagios-plugins
ENV NAGIOS_PLUGINS_BRANCH=master
# 版本查询：https://github.com/NagiosEnterprises/nrpe
ENV NRPE_BRANCH=master
# 版本查询：https://github.com/NagiosEnterprises/ncpa
ENV NCPA_BRANCH=master
# 版本查询：https://github.com/NagiosEnterprises/nsca
ENV NSCA_BRANCH=master
# 版本查询：https://github.com/chriscareycode/nagiostv-react/
ENV NAGIOSTV_VERSION=0.9.10

RUN echo postfix postfix/main_mailer_type string "'Internet Site'" | debconf-set-selections && \
    echo postfix postfix/mynetworks string "127.0.0/8" | debconf-set-selections && \
    echo postfix postfix/mailname string ${NAGIOS_FQDN} | debconf-set-selections && \
    apt-get update && apt-get install -y --no-install-recommends \
        apache2 apache2-utils autoconf automake bc bsd-mailx build-essential dnsutils fping gettext git \
        gperf iputils-ping jq libapache2-mod-php php-cli php-gd libcache-memcached-perl libcgi-pm-perl \
        libcrypt-des-perl libcrypt-rijndael-perl libcrypt-x509-perl libdbd-mysql-perl libdbd-pg-perl \
        libdbi-dev libdbi-perl libdigest-hmac-perl libfreeradius-dev libgdchart-gd2-xpm-dev libgd-gd2-perl \
        libjson-perl libldap2-dev libmonitoring-plugin-perl libmariadb-dev libnagios-object-perl \
        libnet-snmp-perl libnet-tftp-perl libnet-xmpp-perl libpq-dev libradsec-dev libredis-perl \
        librrds-perl libssl-dev libswitch-perl libtext-glob-perl libwww-perl m4 netcat-traditional parallel \
        postfix python3 python3-pip python3-nagiosplugin rsync runit smbclient snmp snmpd \
        snmp-mibs-downloader unzip && \
    apt-get clean && rm -rf /var/lib/lists/* /tmp/* /var/tmp/*

RUN groupadd -r ${NAGIOS_GROUP} || true && \
    groupadd -r ${NAGIOS_CMDGROUP} || true && \
    useradd -r -d ${NAGIOS_HOME} -g ${NAGIOS_GROUP} ${NAGIOS_USER} || true && \
    useradd -r -d ${NAGIOS_HOME} -g ${NAGIOS_CMDGROUP} ${NAGIOS_CMDUSER} || true

RUN cd /tmp && git clone https://github.com/multiplay/qstat.git && cd qstat && \
    ./autogen.sh && ./configure && make -j$(nproc) && make install && rm -rf /tmp/qstat

# 安装 Nagios Core
RUN cd /tmp && git clone https://github.com/NagiosEnterprises/nagioscore.git -b ${NAGIOS_BRANCH} && cd nagioscore && \
    ./configure \
        --prefix=${NAGIOS_HOME} \
        --exec-prefix=${NAGIOS_HOME} \
        --enable-event-broker \
        --with-command-user=${NAGIOS_CMDUSER} \
        --with-command-group=${NAGIOS_CMDGROUP} \
        --with-nagios-user=${NAGIOS_USER} \
        --with-nagios-group=${NAGIOS_GROUP} && \
    make all -j$(nproc) && make install && make install-config && make install-commandmode && make install-webconf && \
    rm -rf /tmp/nagioscore

# 安装官方插件
RUN cd /tmp && git clone https://github.com/nagios-plugins/nagios-plugins.git -b ${NAGIOS_PLUGINS_BRANCH} && cd nagios-plugins && \
    ./tools/setup && ./configure \
        --prefix=${NAGIOS_HOME} \
        --with-ipv6 \
        --with-ping-command="/usr/bin/ping -n -U -W %d -c %d %s" \
        --with-ping6-command="/usr/bin/ping -6 -n -U -W %d -c %d %s" && \
    make -j$(nproc) && make install && mkdir -p /usr/lib/nagios/plugins && \
    ln -sf ${NAGIOS_HOME}/libexec/utils.pm /usr/lib/nagios/plugins && \
    chmod 0755 ${NAGIOS_HOME}/libexec/check_icmp && \
    rm -rf /tmp/nagios-plugins

# NCPA
RUN wget -O ${NAGIOS_HOME}/libexec/check_ncpa.py https://raw.githubusercontent.com/NagiosEnterprises/ncpa/${NCPA_BRANCH}/client/check_ncpa.py && \
    chmod +x ${NAGIOS_HOME}/libexec/check_ncpa.py

# NRPE
RUN cd /tmp && git clone https://github.com/NagiosEnterprises/nrpe.git -b ${NRPE_BRANCH} && cd nrpe && \
    ./configure --with-ssl=/usr/bin/openssl --with-ssl-lib=/usr/lib/x86_64-linux-gnu && \
    make check_nrpe && cp src/check_nrpe ${NAGIOS_HOME}/libexec/ && rm -rf /tmp/nrpe

# NSCA
RUN cd /tmp && git clone https://github.com/NagiosEnterprises/nsca.git && cd nsca && \
    git checkout ${NSCA_BRANCH} && \
    ./configure \
        --prefix=${NAGIOS_HOME} \
        --with-nsca-user=${NAGIOS_USER} \
        --with-nsca-grp=${NAGIOS_GROUP} && \
    make all && cp src/nsca src/send_nsca ${NAGIOS_HOME}/bin/ && \
    cp sample-config/nsca.cfg sample-config/send_nsca.cfg ${NAGIOS_HOME}/etc/ && \
    sed -i 's/^#server_address.*/server_address=0.0.0.0/' ${NAGIOS_HOME}/etc/nsca.cfg && \
    rm -rf /tmp/nsca

# NagiosGraph
RUN cd /tmp && git clone https://git.code.sf.net/p/nagiosgraph/git nagiosgraph && cd nagiosgraph && \
    ./install.pl --install \
        --prefix /opt/nagiosgraph \
        --nagios-user ${NAGIOS_USER} \
        --www-user ${NAGIOS_USER} \
        --nagios-perfdata-file ${NAGIOS_HOME}/var/perfdata.log \
        --nagios-cgi-url /cgi-bin && \
    cp share/nagiosgraph.ssi ${NAGIOS_HOME}/share/ssi/common-header.ssi && \
    rm -rf /tmp/nagiosgraph

# 安装简单 Python 依赖
RUN pip3 install --break-system-packages --no-cache-dir pymssql paho-mqtt

# NagiosTV
RUN cd /tmp && wget https://github.com/chriscareycode/nagiostv-react/releases/download/v${NAGIOSTV_VERSION}/nagiostv-${NAGIOSTV_VERSION}.tar.gz && \
    tar xf nagiostv-${NAGIOSTV_VERSION}.tar.gz -C /opt/nagios/share/ && rm -rf /tmp/nagiostv-*

# Apache 配置
RUN sed -i.bak 's/.*\=www\-data//g' /etc/apache2/envvars && \
    export DOC_ROOT="DocumentRoot ${NAGIOS_HOME}/share" && \
    sed -i "s,DocumentRoot.*,$DOC_ROOT," /etc/apache2/sites-enabled/000-default.conf && \
    sed -i "s,</VirtualHost>,<IfDefine ENABLE_USR_LIB_CGI_BIN>\nScriptAlias /cgi-bin/ ${NAGIOS_HOME}/sbin/\n</IfDefine>\n</VirtualHost>," /etc/apache2/sites-enabled/000-default.conf && \
    ln -s /etc/apache2/mods-available/cgi.load /etc/apache2/mods-enabled/cgi.load && \
    a2enmod session session_cookie session_crypto auth_form request cgi

# SNMP
RUN mkdir -p -m 0755 /usr/share/snmp/mibs ${NAGIOS_HOME}/etc/conf.d ${NAGIOS_HOME}/etc/monitor && \
    mkdir -m 700 ${NAGIOS_HOME}/.ssh && chown ${NAGIOS_USER}:${NAGIOS_GROUP} ${NAGIOS_HOME}/.ssh && \
    ln -s /usr/share/snmp/mibs ${NAGIOS_HOME}/libexec/mibs && \
    ln -s ${NAGIOS_HOME}/bin/nagios /usr/local/bin/nagios && \
    download-mibs && echo "mibs +ALL" > /etc/snmp/snmp.conf

# 邮件配置
RUN sed -i 's,/bin/mail,/usr/bin/mail,' ${NAGIOS_HOME}/etc/objects/commands.cfg && \
    sed -i 's,/usr/usr,/usr,' ${NAGIOS_HOME}/etc/objects/commands.cfg

RUN cp /etc/services /var/spool/postfix/etc/ && echo "smtp_address_preference = ipv4" >> /etc/postfix/main.cf

# 清理无用服务
RUN rm -rf /etc/rsyslog.d /etc/rsyslog.conf /etc/sv/getty-5

# 添加定制文件
ADD overlay /

# 时区
RUN echo "use_timezone=${NAGIOS_TIMEZONE}" >> ${NAGIOS_HOME}/etc/nagios.cfg

# 备份配置
RUN mkdir -p /orig/var /orig/etc /orig/graph-etc /orig/graph-var && \
    cp -Rp ${NAGIOS_HOME}/var/* /orig/var/ && \
    cp -Rp ${NAGIOS_HOME}/etc/* /orig/etc/ && \
    cp -Rp /opt/nagiosgraph/etc/* /orig/graph-etc/ && \
    cp -Rp /opt/nagiosgraph/var/* /orig/graph-var/

RUN chown -R ${NAGIOS_USER}:${NAGIOS_GROUP} ${NAGIOS_HOME}/etc /orig/etc

# Apache 服务名
RUN echo "ServerName ${NAGIOS_FQDN}" > /etc/apache2/conf-available/servername.conf && \
    echo "PassEnv TZ" > /etc/apache2/conf-available/timezone.conf && \
    ln -s /etc/apache2/conf-available/servername.conf /etc/apache2/conf-enabled/ && \
    ln -s /etc/apache2/conf-available/timezone.conf /etc/apache2/conf-enabled/

# 启动脚本权限
RUN chmod +x /usr/local/bin/start_nagios /etc/sv/*/run /opt/nagiosgraph/etc/fix-nagiosgraph-multiple-selection.sh && \
    cd /opt/nagiosgraph/etc && sh fix-nagiosgraph-multiple-selection.sh && rm -f fix-nagiosgraph-multiple-selection.sh

RUN ln -sf /etc/sv/* /etc/service/

ENV APACHE_LOCK_DIR=/var/run
ENV APACHE_LOG_DIR=/var/log/apache2

EXPOSE 80 5667

VOLUME [ \
    "${NAGIOS_HOME}/var", "${NAGIOS_HOME}/etc", "/var/log/apache2", \
    "/opt/Custom-Nagios-Plugins", "/opt/nagiosgraph/var", "/opt/nagiosgraph/etc" \
]

CMD [ "/usr/local/bin/start_nagios" ]
