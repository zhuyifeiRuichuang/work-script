# Upstream Project

Code forked from `https://github.com/malaohu/MobaXterm-GenKey`

# Description

Tool Name: MobaXterm-GenKey

Purpose: Generate MobaXterm activation file, supporting configuration of authorized username, authorized release version number, authorized user count.

# Software Download

`https://mobaxterm.mobatek.net/`

## Local Startup

Requires Python 3 environment

```
pip install --no-cache-dir -r requirements.txt
python app.py
```

# Build Container Image

```bash
docker build -t mobaxterm-genkey:dev1 .
```

# Run Software with Docker

```
# Original image malaohu/mobaxterm-genkey
# Backup image zhuyifeiruichuang/mobaxterm-genkey:dev2
docker run -d \
--name mobaxterm-genkey \
-p 5000:5000 \
mobaxterm-genkey:dev1
```

## Usage

Access via browser: IP:5000, fill in all information, obtain the authentication file.

### Activation Method

Place the authentication file in the MobaXterm software root directory, run the software.