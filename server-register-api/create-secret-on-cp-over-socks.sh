#!/bin/bash

apt update
apt install -y jq curl apt-transport-https ca-certificates curl sudo gnupg
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee -a /etc/apt/sources.list.d/kubernetes.list

# Update package list and install kubectl
sudo apt-get update
sudo apt-get install -y kubectl


secret_name="cp-access-token"
namespace="chisel-client"

while true; do
  secret=$(kubectl -n "$namespace" get secret "$secret_name" --ignore-not-found)
  if [[ -n "$secret" ]]; then
    echo "Secret '$secret_name' found in namespace '$namespace'."
    break
  else
    echo "Secret '$secret_name' not found in namespace '$namespace'. Retrying in 5 seconds..."
    sleep 5
  fi
done

TOKEN=$(kubectl get secret cp-access-token -o json | jq -r '.data["cp-access-token"]' | base64 --decode)


# create edge access token
EDGE_ACCESS_TOKEN="$(kubectl create token edge-access-token -n chisel-client)"

# Set your Kubernetes API server address and bearer token
API_SERVER="https://kubernetes.default.svc.cluster.local"

# Set the desired secret metadata
NAMESPACE="chisel-server"
SECRET_NAME="$EDGE_CLUSTER_NAME"-token
KEY="edge-access-token"
VALUE=$EDGE_ACCESS_TOKEN

# Base64 encode the value
ENCODED_VALUE=$(echo -n "$EDGE_ACCESS_TOKEN" | base64 -w 0)

# Create a JSON payload for the secret
generate_post_data() {
  cat <<EOF
  {
    "apiVersion": "v1",
    "data": {
      "edge-access-token": "${ENCODED_VALUE}"
    },
    "kind": "Secret",
    "metadata": {
      "name": "${SECRET_NAME}",
      "namespace":  "${NAMESPACE}"
    },
    "type": "Opaque"
  }
EOF
}

echo "show json"
generate_post_data

# delete old one
curl -x socks5h://chisel-register:1080 -k -X DELETE "$API_SERVER/api/v1/namespaces/$NAMESPACE/secrets/$EDGE_CLUSTER_NAME" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN"

# create access token to edge in secret on control plane
curl -x socks5h://chisel-register:1080 -k -X POST "$API_SERVER/api/v1/namespaces/$NAMESPACE/secrets" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  --data "$(generate_post_data)"

echo "secret created"
sleep 300