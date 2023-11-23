# minikube

Reminder: minikube config for my linux box.

```bash
minikube start --driver virtualbox --cpus 4 --memory 8192 --disk-size 256g

# one-time cluster configuration
minikube addons enable ingress
minikube addons enable registry-creds
minikube addons enable metallb

ip=$(minikube ip)
minikube addons configure metallb <<EOF
${ip}
${ip}
EOF
```

## Local images

```bash
# build the PIE, uncompressed image on your local docker registry
make image

# preload this image on minikube's host
minikube image load {image}
```

Caveat: with manual load and no registry, need to deploy with `ImagePullPolicy: IfNotPresent`

## Deploy

```bash
make helm-deploy
```

### Undeploy

```bash
make helm-undeploy
```

## TODO
* instructions for private registries (gcr.io, AWS, Azure...)
* instructions for patching a service account with image pull secrets
