replacement for reverse tunnel connection to edge cluster    

encrypted, authorized connections    
multi edge severs connected to central endpoint on control plane    



TODO:    
user password rotation    
initial token creation    
reverse tunnel token refresh    
verify multiple replicas of server


Challenges:    
performance tests with hundreds open connections/chisel server
performance tests with  hundreds    open connections/traefik and ingress 


flow:
1. generate `DEPLOYMENT REGISTER` manifest over `api` - to be ready to be installed on edge cluster
   2. this manifest needs to open cocks proxy
   3. this manifest needs to include token generated on control plane which will allow to create secret on control plane
   4. deployment register needs to create token on edge
   5. deployment register needs to create secret on control plane with token from edge
2. control plane needs to create refreshed token for accessing edge


DEMO:

create edge cluster  


1st cluster

curl -X POST -u 'user:password' "http://192.168.1.172:8080/create" -d '{"edge_cluster_name": "edge-12"}' -H "Content-Type: application/json"

minikube start    --driver=kvm2 --memory 6192 --cpus 8 --kubernetes-version v1.24.4 -p edge-12
k apply -f http://192.168.1.172/edge-12.yaml

k apply -f link

change to control plane/default ctx

TOKEN12=$(kubectl -n chisel-server get secret edge-12-token -o json | jq -r '.data["edge-access-token"]' | base64 --decode)
CURL now

curl -X POST -u 'user:password' "http://192.168.1.172:8080/create" -d '{"edge_cluster_name": "edge-13"}' -H "Content-Type: application/json"
minikube start    --driver=kvm2 --memory 6192 --cpus 8 --kubernetes-version v1.24.4 -p edge-13
k apply -f http://192.168.1.172/edge-13.yaml

change context to default 

TOKEN13=$(kubectl -n chisel-server get secret edge-13-token -o json | jq -r '.data["edge-access-token"]' | base64 --decode)
