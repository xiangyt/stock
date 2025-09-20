## work
自动抓取服务

### 编译
```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./cmd/worker ./cmd/worker/main.go
```

### 打包成docker镜像
```
docker build -f ./cmd/worker/Dockerfile -t stock .
```

### 启动镜像
```
docker run --rm -d -v /Users/xiangyt/go/src/stock/work/data-service:/data --network=host data-service:0.0.1 
```
需要挂载目录输出日志