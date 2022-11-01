
# Dev environment setup

Use [asdf](https://asdf-vm.com/) to install the required tools.

```
asdf plugin add kind
asdf plugin add golang
asdf plugin add kubectl
asdf plugin add skaffold
asdf plugin add golangci-lint
asdf plugin add trivy
asdf plugin add kubebuilder
asdf plugin add k9s
asdf plugin add helm

asdf install
```

# Deploy to a Kind cluster

1. start local Kind cluster

```
make cluster-up
# use make cluster-down to stop the cluster
```

2. install nginx ingress controller, cert-manager and create selfsigned cert

```
make ingress

make cert-manager

# create a selfsigned cert for the "foo.bar" DNS name
make cert
```

3. install NginxOp crd and the controller

```sh
make manifests
make install

# create an NginxOp crd object
# replicas: 1
# host: "foo.bar"
# image: "nginx:latest"
make crd

# run the controller locally for testing
make run

# deploy controller to the Kind cluster
make deploy
```

4. testing

```
curl localhost

# check connection
curl -k -H 'Host: foo.bar' https://localhost

# check certificate
echo | openssl s_client -showcerts -servername foo.bar -connect localhost:443 2>/dev/null | openssl x509 -inform pem -noout -text
```

# Kubebuilder

Kubebuilder was used to initialize the project.

1. init project

```
kubebuilder init --domain my.domain --repo my.domain/nginxop
```

2. create api
```
kubebuilder create api --group nginxop --version v1 --kind NginxOp
```
