
# download edge access token from chisel server cluster
TOKEN=$(kubectl -n chisel-server get secret edge-access-token -o json | jq -r '.data["edge-access-token"]' | base64 --decode)

curl -k -H "Authorization:Bearer $TOKEN" -s https://chisel-server.lan/client-1/api/v1/namespaces/kube-system/pods | jq '.items[].metadata.name'
