# docker build -t your-registry/hadoop-operator:0.1.0 -f operator/Dockerfile operator
FROM python:3.12-slim-bookworm

WORKDIR /app
ENV PYTHONUNBUFFERED=1 \
    PIP_NO_CACHE_DIR=1

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY hadoop_operator ./hadoop_operator

USER 65534:65534
ENTRYPOINT ["python", "-m", "hadoop_operator"]
