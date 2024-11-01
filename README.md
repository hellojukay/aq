# aq
A key list sever, save a key and qury return a list.
```bash
go install github.com/hellojukay/aq@latest
```
# Usage
```bash
aq (main) $ ./aq -h
  -dir string
    	server data dir, default ./data (default "./data")
  -port int
    	server port, default 9090 (default 9090)
  -prefix string
    	server api prefix, default api (default "api")
```
Post an pair of key and value to the server:
```bash
curl -X POST http://localhost:9090/api/{name}:{tag}
```
Query key`s value list:
```bash
curl http://localhost:9090/api/{{name}
```

# Download
[https://github.com/hellojukay/aq/releases](https://github.com/hellojukay/aq/releases)
