# Kubernetes Validating Admission Webhook for etcd member remove


## Prerequisites

Kubernetes 1.9.0 or above with the `admissionregistration.k8s.io/v1beta1` API enabled. Verify that by the following command:
```
kubectl api-versions | grep admissionregistration.k8s.io/v1beta1
```
The result should be:
```
admissionregistration.k8s.io/v1beta1
```

In addition, the `MutatingAdmissionWebhook` and `ValidatingAdmissionWebhook` admission controllers should be added and listed in the correct order in the admission-control flag of kube-apiserver.

## Build

1. Setup dep

   The repo uses [dep](https://github.com/golang/dep) as the dependency management tool for its Go codebase. Install `dep` by the following command:
```
go get -u github.com/golang/dep/cmd/dep
```

2. Build and push docker image

```
./build
```

## Deploy

1. Create a signed cert/key pair and store it in a Kubernetes `secret` that will be consumed by sidecar deployment
```
./deployment/webhook-create-signed-cert.sh \
    --service etcd-webhook-webhook-svc \
    --secret etcd-webhook-webhook-certs \
    --namespace default
```

2. Patch the `MutatingWebhookConfiguration` by set `caBundle` with correct value from Kubernetes cluster
```
cat deployment/mutatingwebhook.yaml | \
    deployment/webhook-patch-ca-bundle.sh > \
    deployment/mutatingwebhook-ca-bundle.yaml
```

3. Deploy resources
```
kubectl create -f deployment/deployment.yaml
kubectl create -f deployment/service.yaml
kubectl create -f deployment/mutatingwebhook-ca-bundle.yaml
```

## Verify

1. The sidecar inject webhook should be running
```
[root@mstnode ~]# kubectl get pods
NAME                                                  READY     STATUS    RESTARTS   AGE
etcd-webhook-deployment-bbb689d69-882dd   1/1       Running   0          5m
[root@mstnode ~]# kubectl get deployment
NAME                                  DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
etcd-webhook-deployment   1         1         1            1           5m
```

2. Label the default namespace with `etcd-webhook=enabled`
```
kubectl label namespace default etcd-webhook=enabled
[root@mstnode ~]# kubectl get namespace -L etcd-webhook
NAME          STATUS    AGE       etcd-webhook
default       Active    18h       enabled
kube-public   Active    18h
kube-system   Active    18h
```

3. Deploy an app in Kubernetes cluster, take `sleep` app as an example
```
[root@mstnode ~]# cat <<EOF | kubectl create -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: etcd-test-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: test-server
  template:
    metadata:
      annotations:
        etcd.web-hook.me/remove: "yes"
      labels:
        app: test-server
    spec:
      containers:
      - name: test-server
        image: k8s.gcr.io/etcd-statefulset-e2e-test:0.0
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
        readinessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 2
          periodSeconds: 2

EOF
```


