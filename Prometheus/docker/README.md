# 说明

用于docker环境部署Prometheus

# 部署

部署前，按照业务需求调整配置文件。执行部署脚本即可。

```bash
bash deploy.sh
```



# 脚本说明

```bash
# 创建专用数据卷
docker volume create prometheus-data
# 启动容器
docker run -d \
  --name prometheus \
  --restart unless-stopped \
  -p 9090:9090 \
  -e TZ=Asia/Shanghai \
  # 配置文件应真实存在且配置有效数据。
  -v /opt/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml \
  -v prometheus-data:/prometheus \
  --health-cmd="wget --no-verbose --tries=1 --spider http://localhost:9090/-/healthy || exit 1" \
  --health-interval=15s \
  --health-timeout=5s \
  --health-retries=3 \
  # 推荐使用最新版本。
  prom/prometheus
```

# 配置文件说明

```bash
# my global config
global:
  scrape_interval: 15s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
  evaluation_interval: 15s # Evaluate rules every 15 seconds. The default is every 1 minute.
  # scrape_timeout is set to the global default (10s).

# Alertmanager configuration
alerting:
  alertmanagers:
    - static_configs:
        - targets:
          # - alertmanager:9093

# Load rules once and periodically evaluate them according to the global 'evaluation_interval'.
rule_files:
# 若有新增的配置文件，务必在此处写文件名，且挂载到容器内/etc/prometheus/下。
  # - "first_rules.yml"
  # - "second_rules.yml"

# A scrape configuration containing exactly one endpoint to scrape:
# Here it's Prometheus itself.
scrape_configs:
  # The job name is added as a label `job=<job_name>` to any timeseries scraped from this config.
  - job_name: "prometheus"

    # metrics_path defaults to '/metrics'
    # scheme defaults to 'http'.

    static_configs:
    # 监控自己可以写localhost
      - targets: ["localhost:9090"]
       # The label name is added as a label `label_name=<label_value>` to any timeseries scraped from this config.
        labels:
          app: "prometheus"
  - job_name: node
    static_configs:
    # 监控其他资源务必写容器可访问到的IP和端口。
    - targets: ['localhost:9100']

```

