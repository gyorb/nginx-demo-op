
1. use [asdf](https://asdf-vm.com/) to install the required tools

```
asdf plugin add kind
asdf plugin add golang
asdf plugin add kubectl
asdf plugin add skaffold
asdf plugin add golangci-lint
asdf plugin add trivy
asdf plugin add kubebuilder

asdf install
```

2. init project

```
kubebuilder init --domain my.domain --repo my.domain/nginxop
```

3. create api
```
kubebuilder create api --group nginxop --version v1 --kind NginxOp
```

4. start local kind cluster
```
make cluster-up
# use make cluster-down to stop the cluster
```

5. install nginx ingress controller, cert-manager and create selfsigned cert

```
make ingress

make cert-manager
make cert
```

6. install crd and controller

```sh
# install CRD

make manifests
make install

# create crd object
make crd

# run the controller locally for testing
make run

# deploy controller docker image to kind
make load-image deploy

```

6. testing
```
curl localhost

# check connection
curl -k -H 'Host: foo.bar' https://localhost

# check certificate
echo | openssl s_client -showcerts -servername foo.bar -connect localhost:443 2>/dev/null | openssl x509 -inform pem -noout -text
```
