EDGE_CLUSTER_NAME=client-1
echo > $EDGE_CLUSTER_NAME.yaml ||true
cp template-manifest-for-edge.yaml $EDGE_CLUSTER_NAME.yaml

CHISEL_REGISTER_TOKEN=$(kubectl create token  chisel-register-sa -n chisel-server)

kubectl -n chisel-client create secret generic cp-access-token --from-literal=cp-access-token=$CHISEL_REGISTER_TOKEN --dry-run=client -o yaml >> $EDGE_CLUSTER_NAME.yaml
echo --- >> $EDGE_CLUSTER_NAME.yaml
kubectl -n chisel-client create configmap gen-token-over-socks --from-file=create-secret-on-cp-over-socks.sh --dry-run=client -o yaml  >> $EDGE_CLUSTER_NAME.yaml
#