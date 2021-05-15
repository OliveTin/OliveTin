# OliveTin

OliveTin is a web based quick access control panel for running jobs.

For example, it can be used to turn home automation lights on or off, or start
workflows in n8n.  

## `config.yaml`

```
listenAddressRestActions: :1337 # Listen on all addresses available, port 1337
listenAddressWebUi: :1339
logLevel: "INFO"
```

## Building the container 

### Podman/Docker

```
podman create --name olivetin -p 1337 -p 1338 -p 1339 -v /etc/olivetin/:/config:ro olivetin

```

### Buildah/Docker

```
buildah bud -t olivetin
```

