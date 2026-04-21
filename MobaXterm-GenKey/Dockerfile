FROM python:3.13-slim

LABEL maintainer="malaohu <tua@live.cn>"

WORKDIR /usr/src/app

COPY requirements.txt ./

RUN pip install --no-cache-dir -r requirements.txt

COPY app.py ./
COPY index.html ./

CMD [ "python", "./app.py" ]
