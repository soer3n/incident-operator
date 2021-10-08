# Development

There repo contains a okteto.yml which can be used to play with the components of this project in a kubernetes cluster. For running the webhook server you need to generate a ca and client cert by yourself at the moment. 

# running local

You have to run it either local or you have to build an image and push it to your own account/repository. For the second way only docker is needed. If you want to run it local you need to install [golang](https://golang.org/doc/install) if not already done and [operator-sdk](https://sdk.operatorframework.io/docs/installation/).

```

# Install the CRDs
make install


# Building and pushing as an image to private registry
export IMG="image_name:image_tag"
make docker-build docker-push

# create image pull secret if needed (if private registry is used)
kubectl create secret generic harbor-registry-secret -n helm --from-file=.dockerconfigjson=harbor.json --type=kubernetes.io/dockerconfigjson

# Deploy the built operator
kubectl apply -f deploy/rbac.yaml
cat deploy/operator.yaml | envsubst | kubectl apply -f -

########
## OR ##
########

# Run it local
make run

```

# tests

Currently there is only a rudimentary integration test. In general tests can be run from root dir with this command:

```

ACK_GINKGO_DEPRECATIONS=1.16.4 go test -v ./...

```
