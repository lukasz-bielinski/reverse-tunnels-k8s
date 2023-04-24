WEB_STORAGE="/data/"

EDGE_CLUSTER_NAME=$1
EXPOSE_PORT=$2

echo > $EDGE_CLUSTER_NAME.yaml ||true

cp template-manifest-for-edge.yaml $WEB_STORAGE/$EDGE_CLUSTER_NAME.yaml
sed -i "s/<<--EXPOSE-PORT-->>/$EXPOSE_PORT/g"  $WEB_STORAGE/$EDGE_CLUSTER_NAME.yaml
sed -i "s/<<--EDGE-CLUSTER-NAME-->>/$EDGE_CLUSTER_NAME/g"  $WEB_STORAGE/$EDGE_CLUSTER_NAME.yaml

CHISEL_REGISTER_TOKEN=$(kubectl create token  chisel-register-sa --duration=17h -n chisel-server)

kubectl -n chisel-client create secret generic cp-access-token --from-literal=cp-access-token=$CHISEL_REGISTER_TOKEN --dry-run=client -o yaml >> $WEB_STORAGE/$EDGE_CLUSTER_NAME.yaml
echo --- >> $WEB_STORAGE/$EDGE_CLUSTER_NAME.yaml
kubectl -n chisel-client create configmap gen-token-over-socks --from-file=create-secret-on-cp-over-socks.sh --dry-run=client -o yaml  >> $WEB_STORAGE/$EDGE_CLUSTER_NAME.yaml
#