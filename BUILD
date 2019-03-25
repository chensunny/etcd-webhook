dep ensure
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o etcd-webhook .
docker build --no-cache -t registry-internal.cn-hangzhou.aliyuncs.com/sunny_chen/sunny_chen/etcd-webhook:v1 .
rm -rf etcd-webhook

docker push registry-internal.cn-hangzhou.aliyuncs.com/sunny_chen/sunny_chen/etcd-webhook:v1
