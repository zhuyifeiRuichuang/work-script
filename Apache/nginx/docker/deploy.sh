docker run -d \
  --name nginx \
  --restart unless-stopped \
  -p 80:80 \
  -e TZ=Asia/Shanghai \
  -v /data/nginx/conf.d:/etc/nginx/conf.d:ro \
  -v /data/nginx/html:/usr/share/nginx/html:rw \
  --health-cmd="curl -f http://localhost/ || exit 1" \
  --health-interval=30s \
  --health-timeout=10s \
  --health-retries=3 \
  --health-start-period=10s \
  nginx:1.29.5