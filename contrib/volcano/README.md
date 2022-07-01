## 部署命令：
1. docker save -o sfe-volcano-go-v1.x.tar sfe-volcano-go:v1.x

2. scp sfe-volcano-go-v1.x.tar 10.28.124.131:/tmp 
    ssh 10.28.124.131 
    docker load < /tmp/sfe-volcano-go:v1.x.tar
    scp sfe-volcano-go-v1.x.tar 10.28.124.132:/tmp
    ssh 10.28.124.132
    docker load < /tmp/sfe-volcano-go:v1.x.tar
    
3. kustomize4 build volcano/overlays/dev | kubectl apply -f -