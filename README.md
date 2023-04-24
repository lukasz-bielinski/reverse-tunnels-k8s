# reverse-tunnels-k8s
This project is a prototype to allow to communicate to Kubernetes clusters hidden behind firewalls, nat etc.

This project uses:
1. [Chisel](https://github.com/jpillora/chisel)
2. [Traefik](https://doc.traefik.io/traefik/)


General flow:
1. manifest generated via `poor's man api' is applied on the edge cluster
2. edge cluster creates secret with access token to the edge cluster on control plane
2. Reverse tunnel to edge cluster is exposed on control plane