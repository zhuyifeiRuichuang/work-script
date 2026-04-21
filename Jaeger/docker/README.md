# 说明

部署前按业务需求修改脚本。



# 脚本说明

```bash
#!/bin/bash

# /opt/jaeger/config.yaml可参考https://github.com/jaegertracing/jaeger/tree/main/cmd/jaeger
# 部署
docker run -d \
  --name jaeger \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  -p 5778:5778 \
  -p 9411:9411 \
  -v /opt/jaeger/config.yaml:/jaeger/config.yaml \
  jaegertracing/jaeger:2.16.0 \
  --config /jaeger/config.yaml
```

若仅快捷体验，采用以下脚本

```bash
#!/bin/bash

# 部署
docker run -d \
  --name jaeger \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  -p 5778:5778 \
  -p 9411:9411 \
  jaegertracing/all-in-one
```

