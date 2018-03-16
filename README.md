# Kubernetes NodePort-Exposer [![CircleCI](https://circleci.com/gh/kubermatic/nodeport-exposer.svg?style=svg)](https://circleci.com/gh/kubermatic/nodeport-exposer)

Controller which exposes NodePorts via a LoadBalancer service.

## Overview
The NodePort-Exposer watches Services with the annotation `nodeport-exposer.k8s.io/expose="true"` and exposes them via a Service of type `LoadBalancer`.

Routing of traffic will happen via service-to-service forwarding in Kubernetes: https://kubernetes.io/docs/concepts/services-networking/service/#services-without-selectors

## Deployment

### Automatic

`nodeport-exposer` is wired to CircleCI thus if you want to publish a new Docker image all you have to do is to add a tag
that starts with "v" prefix to the repository. The CI will pick it up and will start the build process. At the end the CI will push the new image to Docker Hub.
May the force be with you!

### With RBAC enabled
```
kubectl create -f example/deployment-with-rbac.yaml
# Or without rbac kubectl create -f example/deployment.yaml

# expose an example nodeport service
kubectl create -f example/service.yaml
```
