
# Incident operator

This operator is for managing incidents in clusters by controllers and custom resources. For now the only implementation is for setting nodes in quarantine with isolating pod from workloads and debugging affected nodes by a deployed pod.

## Installation

For now there is no docker image neither for the operator nor for the planned web backend. So you have to run it either local or you have to build an image and push it to your own account/repository. For the second way only docker is needed. If you want to run it local you need to install [golang](https://golang.org/doc/install) if not already done and [operator-sdk](https://sdk.operatorframework.io/docs/installation/).

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


## Architecture

[Here](docs/COMPONENTS.md) is an explanation how the operator works.

## Usage
[Here](docs/USAGE.md) is an explanation how the operator can be used.

## License
[LICENSE](LICENSE)
