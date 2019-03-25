FROM alpine:latest

ADD etcd-webhook /etcd-webhook
ENTRYPOINT ["./etcd-webhook"]
