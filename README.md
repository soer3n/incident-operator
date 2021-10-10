
# Incident operator

This operator can be used for managing incidents in clusters by controllers and custom resources. For now the only implementation is for setting nodes in quarantine with isolating pod from workloads and debugging affected nodes by a deployed pod.

## Installation

```

helm repo add charts https://soer3n.github.io/charts/charts
helm upgrade --install incident-operator charts/incident-operator

```

## Architecture

[Here](docs/COMPONENTS.md) is an explanation how the operator works.

## Usage
[Here](docs/USAGE.md) is an explanation how the operator can be used.

## License
[LICENSE](LICENSE)
