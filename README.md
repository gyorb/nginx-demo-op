
# Dev environment setup

Use [asdf](https://asdf-vm.com/) to install the required tools.

```sh
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

```sh
make cluster-up
# use make cluster-down to stop the cluster
```

2. install nginx ingress controller, cert-manager and create selfsigned cert

```sh
make ingress

make cert-manager

# Create a selfsigned cert for the "foo.bar" DNS name
# wait until cert-manager starts before creating it.
make cert
```

3. install NginxOp crd and the controller

```sh
make manifests
make install

# create an nginxop custom resource object
# replicas: 1
# host: "foo.bar"
# image: "nginx:latest"
make crd

# run the controller locally for testing
make run

# Deploy controller (image from github) to the Kind cluster
# Use the 'main' tag to test the latest non released image.
IMG=ghcr.io/gyorb/nginx-demo-op:v0.0.1 make deploy

# Load the locally built container image to a Kind cluster
# and deploy the controller.
make load-image deploy
```

4. Check if the nginx pods are running.
```sh
# Get the created nginx pods.
kubectl get pod --selector=nginx=nginx-op-sample
```

5. Edit the example custom resource.
```sh
kubectl edit nginxop -n default nginx-op-sample
```

4. Verify the connection and the certificates for the nginx pods.

```sh
# 404 page not found
curl -k localhost

# nginx should respond wih the welcome page.
curl -k -H 'Host: foo.bar' https://localhost

# check certificate
echo | openssl s_client -showcerts -servername foo.bar -connect localhost:443 2>/dev/null | openssl x509 -inform pem -noout -text
```

# Kubebuilder

Kubebuilder was used to initialize the project.

1. init project

```sh
kubebuilder init --domain my.domain --repo my.domain/nginxop
```

2. create api
```sh
kubebuilder create api --group nginxop --version v1 --kind NginxOp
```
