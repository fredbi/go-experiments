# minikube

Reminder: minikube config for my linux box.

```bash
minikube start --driver virtualbox --cpus 4 --memory 8192 --disk-size 256g

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
make image

minikube image load {image}
```

Caveat: with manual load and no registry, need to deploy with `ImagePullPolicy: IfNotPresent`

## TODO
* instructions for private registries (gcr.io, AWS, Azure...)
* instructions for patching a service account with image pull secrets
