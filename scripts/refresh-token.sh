# possible scenario how to refresh access token to the edge cluster
# on the edge cluster
#kubectl create serviceaccount api-explorer
#
#kubectl create clusterrolebinding api-explorer:cluster-admin --clusterrole cluster-admin --serviceaccount default:api-explorer

# on the central cluster
## create new token for access edge cluster
# old token needs to be still valid, as new will be created over reverse tunnel

# json request
export TOKEN="eyJhbGciOiJSUzI1NiIsImtpZCI6IjI5dFJMczBaa1FSOHpHMHdRUHpmZUZKLUxKNEpQczEyaHp6TGNxYVR3OFEifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiXSwiZXhwIjoxNjgyNTI5NTc1LCJpYXQiOjE2ODI1MjU5NzUsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJkZWZhdWx0Iiwic2VydmljZWFjY291bnQiOnsibmFtZSI6ImVkZ2UtMzUtY3AtYWNjZXNzIiwidWlkIjoiODE0ZTI3Y2ItOTQyYS00NDEzLWI1ZjctMDQxNjJkNDhjOTUxIn19LCJuYmYiOjE2ODI1MjU5NzUsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmVkZ2UtMzUtY3AtYWNjZXNzIn0.ckBnabJqkN1DXzpptD1SSdJUyorkF6Hc1KVTVwDLRwyFnxyZne3jWVpu6868CVODz2T8LSgJFlomG36kmE6xPcm9_1R5Vwi9mCpsUzpbjiiHzMxO4cwol0dGifBnF5nmR-bUlxC1TA9BneKP8qHsfjPJRxiAYeJ9mb1u6hC5j6LOCHtCzl0g-MjmZixlv4Oco9PLw77YeqYp7SXHm9Xch6-XMsNZD8haDpOyKdJV9OVa2vHm9BS54eOFezOBp6tfGOZruYfKdAyY_v2s3qFUFyzRFjKGLFBznbQ6ELccNRnvRHFHAIWywP8vK8oN_83-wJtRP_x88u5XPlm6FmlvCw%"
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
      -k -H "Authorization:Bearer $TOKEN" -s https://chisel-tunnel.lan/ch4ki5jh1k/api/v1/namespaces/default/serviceaccounts/edge-35-cp-access/token |jq -r .status.token    )

echo "test new token"
echo $NEW_ACCESS_TOKEN
curl -k -H "Authorization:Bearer $NEW_ACCESS_TOKEN" -s https://chisel-tunnel.lan/ch4ki5jh1k/api/v1/namespaces/kube-system/pods | jq '.items[].metadata.name'


#kubectl create secret generic client-1-access-token --from-literal=new-access-token=$NEW_ACCESS_TOKEN