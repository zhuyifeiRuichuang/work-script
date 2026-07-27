# 说明

Apache Kafka在k8s环境部署的方法依赖第三方，目前没有稳定可用的免费版。

v1 ，Strimzi方法，yaml文件部署方法。参考目录中部署指导。

v2，基于Docker版本改造，集群架构，3节点，combined模式，无SSL。支持集群内访问。

v3, 基于v2新增web UI。取消nodeport的注释即可启用集群外访问。

v4，基于v3，新增对集群外服务的支持，使kafka可同时对集群外提供服务。

v5，基于Docker版本改造，集群架构，3节点，combined模式，配置SSL。支持集群内外访问。

v6，基于v5，支持web ui

# 配置

注意，部署时，先修改yaml里的nodePort端口，确认使用空闲端口。受限kafka功能，必须配置端口。

# 部署

```bash
kubectl create namespace kafka
kubectl apply -f v4.yaml -n kafka
```

# 测试

部署后，有web UI的可访问web界面查看集群状态。



配置SSL的场景进行测试时，

如下所示，要给部署的资源加SSL的配置。调用已部署的kafka的secrets的资源。

```bash
apiVersion: apps/v1
kind: Deployment
metadata:
  name: component-a
  labels: { app: component-a }
spec:
  replicas: 1
  selector: { matchLabels: { app: component-a } }
  template:
    metadata:
      labels: { app: component-a }
    spec:
      containers:
        - name: component-a
          image: your-registry/component-a:latest
          env:
            # 用环境变量把上面那 8 行喂给应用（Spring Boot / 普通 Java 都认）
            - { name: KAFKA_BOOTSTRAP_SERVERS, value: "kafka-0.kafka-headless:19093,kafka-1.kafka-headless:19093,kafka-2.kafka-headless:19093" }
            - { name: KAFKA_SECURITY_PROTOCOL, value: "SSL" }
            - { name: KAFKA_SSL_TRUSTSTORE_LOCATION, value: "/etc/kafka/secrets/kafka.truststore.jks" }
            - { name: KAFKA_SSL_TRUSTSTORE_PASSWORD, value: "abcdefgh" }
            - { name: KAFKA_SSL_KEYSTORE_LOCATION, value: "/etc/kafka/secrets/kafka01.keystore.jks" }
            - { name: KAFKA_SSL_KEYSTORE_PASSWORD, value: "abcdefgh" }
            - { name: KAFKA_SSL_KEY_PASSWORD, value: "abcdefgh" }
            - { name: KAFKA_SSL_ENDPOINT_IDENTIFICATION_ALGORITHM, value: "" }
          volumeMounts:
            - { name: kafka-secrets, mountPath: /etc/kafka/secrets, readOnly: true }
      volumes:
        - name: kafka-secrets
          secret: { secretName: kafka-secrets }   # 与 kafka / kafka-ui 同一个 Secret

```

