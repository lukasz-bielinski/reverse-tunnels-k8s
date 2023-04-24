# possible scenario how to refresh access token to the edge cluster
# on the edge cluster
kubectl create serviceaccount api-explorer

kubectl create clusterrolebinding api-explorer:cluster-admin --clusterrole cluster-admin --serviceaccount default:api-explorer

# on the central cluster
## create new token for access edge cluster
# old token needs to be still valid, as new will be created over reverse tunnel

# json request
export TOKEN="eyJhbGciOiJSUzI1NiIsImtpZCI6InN4U2diRmx1VXZvckFDTjItWnRmM3RZZEZBMUFLdDJLb1p2dldzVVVwZHcifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiXSwiZXhwIjoxNjgxOTc1Mjk5LCJpYXQiOjE2ODE5MTQwOTksImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJkZWZhdWx0Iiwic2VydmljZWFjY291bnQiOnsibmFtZSI6ImFwaS1leHBsb3JlciIsInVpZCI6ImNhNzI2M2MxLWU0MzMtNGJhZS04M2RhLWU1NTFhYmNkZWFiZCJ9fSwibmJmIjoxNjgxOTE0MDk5LCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6ZGVmYXVsdDphcGktZXhwbG9yZXIifQ.S9fHasR4JIxAzX_DTGxP3TfkHyVVKsnXlzaJb9ASxdPRbwYuiBMYFTwABDK38OI2BJzwNXVk-YzMWyItHQEicUebmKDIaHJXxqtXZ37ta7oGWn0aw37d3WsTa89eVl0QHke7HylS5eGGIZoiPVnWFazrAliaQ9-2t5avDs8DMGb7dRHBAqDXuRX6_V_pR1SI7IhjcfIDfScLkl6D61B_9eGCwB8p7uvhqKTSoXttNjQb6sAZwYtgbZW94F6tHBi3MPbrOTkuoMo2j-eacaMEU7WYpc4F27SITIwNCpMynaIfLFfifTdJPNNUZF3ExFE4dcyhGHE6YXTcxhhsmn9Q3w"


generate_post_data()
{
  cat <<EOF
{
  "kind": "TokenRequest",
  "apiVersion": "authentication.k8s.io/v1",
  "metadata": {
    "creationTimestamp": null
  },
  "spec": {
    "audiences": null,
    "expirationSeconds": 61200,
    "boundObjectRef": null
  },
  "status": {
    "token": "",
    "expirationTimestamp": null
  }
}
EOF
}

NEW_ACCESS_TOKEN=$(curl   \
      -H "Accept: application/json" \
      -H "Content-Type:application/json" \
      -X POST --data "$(generate_post_data)" \
      -k -H "Authorization:Bearer $TOKEN" -s https://192.168.1.83:1111/api/v1/namespaces/default/serviceaccounts/api-explorer/token |jq -r .status.token    )

echo "test new token"
echo $NEW_ACCESS_TOKEN
curl -k -H "Authorization:Bearer $NEW_ACCESS_TOKEN" -s https://192.168.1.83:1111/api/v1/namespaces/kube-system/pods | jq '.items[].metadata.name'


kubectl create secret generic client-1-access-token --from-literal=new-access-token=$NEW_ACCESS_TOKEN