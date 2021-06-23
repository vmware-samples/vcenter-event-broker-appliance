# kn-ps-mm-vrops
This example is a function which when a host enters Maintenance Mode in vCenter makes call to vRealize Operations Manager to mark host Maintenance Mode state.

# Step 1 - Build container image

Create the container images and optionally push to an external registry like GitHub Container Registry.

```bash
docker build -t <your-container-repository>/kn-ps-mm-sync-image:1.0 .
# docker buildx build --platform linux/amd64 -t ghcr.io/darrylcauldwell/kn-ps-mm-sync-image:1.0 .
docker push <your-container-repository>/kn-ps-mm-sync-image:1.0
# docker push ghcr.io/darrylcauldwell/kn-ps-mm-sync-image:1.0
```

# Step 2 - Define environment variables

In order to make successful call to vRealize Operations Manager environmentally specific FQDN and credentials are required. The function depends on environment variables passed as a Kubernetes secret.

```bash
# Define environment variables as secret

kubectl -n vmware-functions create secret generic kn-ps-mm-sync-secret \
  --from-literal=vropsFqdn=vrops.rainpole.local \
  --from-literal=vropsUser=admin \
  --from-literal=vropsPassword='DontUseThisPassword'
```

# Step 3 - Deploy

```bash
# Deploy function

kubectl -n vmware-functions apply -f kn-ps-mm-sync.yml
```

# Step 4 - Undeploy

```bash
# Undeploy function

kubectl -n vmware-functions delete -f kn-ps-mm-sync.yaml
```