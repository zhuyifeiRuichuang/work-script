# 说明



参考`https://github.com/apache/kafka/blob/trunk/docker/examples/README.md`

`https://hub.docker.com/r/apache/kafka`



区分无ssl和有ssl



# 阿里提供的参考

Kafka安全证书准备
## 生成SSL证书

```bash
keytool -genkey -keystore server.keystore.jks -alias kafka-1 -validity 365 -keyalg RSA
```



## 配置SCRAM用户

```bash
kafka-configs.sh --bootstrap-server localhost:9092 \
  --alter --add-config 'SCRAM-SHA-256=[password=your-secure-password]' \
  --entity-type users --entity-name admin
```



# 豆包提供的参考

SSL配置

生成密码凭证

```bash
echo "abcdefgh" > kafka_keystore_creds   # 密钥库密码文件
echo "abcdefgh" > kafka_ssl_key_creds    # 私钥密码文件
echo "abcdefgh" > kafka_truststore_creds # 信任库密码文件
```



配置CA密钥库

```bash
# 生成CA的密钥对（一次性创建，用于签名服务端/客户端证书）
keytool -genkeypair \
  -alias ca \
  -keyalg RSA \
  -keysize 2048 \
  -keystore ca.keystore.jks \
  -storepass abcdefgh \
  -storetype JKS \
  -dname "CN=kafka-ca"  # 自定义CA的标识
```



配置服务端客户端密钥对

```bash
# 跨主机部署时，生成服务端keystore（替换为实际IP/域名）
keytool -genkeypair \
  -alias kafka \
  -keyalg RSA \
  -keysize 2048 \
  -keystore kafka01.keystore.jks \
  -storepass abcdefgh \
  -storetype JKS \
  -keypass abcdefgh \
  -dname "CN=192.168.1.20" \ # 服务端实际IP
  -ext SAN=DNS:kafka-server,DNS:192.168.1.20,IP:192.168.1.20 # 包含所有跨主机访问的标识
```



配置证书签名请求CSR

```bash
keytool -certreq \
  -keystore kafka01.keystore.jks \
  -storepass abcdefgh \
  -storetype JKS \
  -keypass abcdefgh \
  -alias kafka \
  -file kafka.csr  # 生成的CSR文件
```



用CA签名CSR

```bash
keytool -gencert \
  -keystore ca.keystore.jks \  # 步骤1创建的CA密钥库
  -storepass abcdefgh \
  -storetype JKS \
  -alias ca \
  -infile kafka.csr \          # 步骤3生成的CSR
  -outfile kafka.crt \         # 签名后的证书
  -dname "CN=systemtest" \
  -ext SAN=DNS:broker,DNS:kafka-1,DNS:localhost
```



导入CA证书和签名证书到密钥库

```bash
# 1. 导出CA证书（从CA密钥库中提取）
keytool -exportcert \
  -keystore ca.keystore.jks \
  -storepass abcdefgh \
  -alias ca \
  -file ca.crt \
  -noprompt

# 2. 导入CA证书到服务端keystore（信任CA）
keytool -importcert \
  -keystore kafka01.keystore.jks \
  -storepass abcdefgh \
  -storetype JKS \
  -alias ca \
  -file ca.crt \
  -noprompt

# 3. 导入签名后的证书到keystore（完成私钥+证书绑定）
keytool -importcert \
  -keystore kafka01.keystore.jks \
  -storepass abcdefgh \
  -storetype JKS \
  -keypass abcdefgh \
  -alias kafka \
  -file kafka.crt \
  -noprompt
```



信任库

```bash
# 导入CA证书到truststore（复用步骤5导出的ca.crt）
keytool -importcert \
  -keystore kafka.truststore.jks \  # 生成的信任库文件名
  -storepass abcdefgh \             # 信任库密码（对应kafka_truststore_creds）
  -storetype JKS \
  -alias ca \
  -file ca.crt \
  -noprompt
```



