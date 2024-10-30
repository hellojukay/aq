# aq
A key list sever, save a key and qury return a list.
# Usage
```bash
aq (main) $ ./aq -h
Usage of ./aq:
  -dir string
        server data dir, default ./data (default "./data")
  -port int
        server port, default 9090 (default 9090)
```
Post an image to the server:
```bash
curl -X POST http://localhost:8081/image/{name}:{tag}
```
Query image version list:
```bash
curl http://localhost:8081/image/{{name}
```

# Download
[https://github.com/hellojukay/aq/releases](https://github.com/hellojukay/aq/releases)
