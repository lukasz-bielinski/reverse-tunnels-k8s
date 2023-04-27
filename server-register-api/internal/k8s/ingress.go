package k8s

import (
	"context"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"log"
	"os"
)

func CreateIngress(edgeClusterName, exposeName, namespace string, port int) error {
	log.Printf("Creating Ingress: edgeClusterName=%s, exposeName=%s, namespace=%s\n", edgeClusterName, exposeName, namespace)

	chiselTunnelDomain := os.Getenv("CHISEL_TUNNEL_DOMAIN")
	chiselTunnelHost := os.Getenv("CHISEL_TUNNEL_HOST")

	certManagerClusterIssuer := os.Getenv("CERT_MANAGER_CLUSTER_ISSUER")
	if certManagerClusterIssuer == "" {
		certManagerClusterIssuer = "self-signed-issuer" // Use a default value if the environment variable is not set
	}
	chiselTunnelHost := os.Getenv("CHISEL_TUNNEL_HOST")
	if chiselTunnelHost == "" {
		chiselTunnelHost = "chisel-tunnel" // Use a default value if the environment variable is not set
	}

	ingressObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "networking.k8s.io/v1",
			"kind":       "Ingress",
			"metadata": map[string]interface{}{
				"name":      edgeClusterName,
				"namespace": namespace,
				"annotations": map[string]interface{}{
					"cert-manager.io/cluster-issuer":                   certManagerClusterIssuer,
					"kubernetes.io/ingress.class":                      "traefik",
					"traefik.ingress.kubernetes.io/router.middlewares": "chisel-server-" + edgeClusterName + "@kubernetescrd",
				},
			},
			"spec": map[string]interface{}{
				"rules": []map[string]interface{}{
					{
						"host": chiselTunnelDomain,
						"http": map[string]interface{}{
							"paths": []map[string]interface{}{
								{
									"path":     "/" + exposeName,
									"pathType": "Prefix",
									"backend": map[string]interface{}{
										"service": map[string]interface{}{
											"name": edgeClusterName,
											"port": map[string]interface{}{
												"number": port,
											},
										},
									},
								},
							},
						},
					},
				},
				"tls": []map[string]interface{}{
					{
						"hosts": []string{
							chiselTunnelHost,
						},
						"secretName": chiselTunnelHost,
					},
				},
			},
		},
	}

	// Create the Ingress custom resource in Kubernetes
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	ingressGVR := schema.GroupVersionResource{
		Group:    "networking.k8s.io",
		Version:  "v1",
		Resource: "ingresses",
	}
	_, err = dynamicClient.Resource(ingressGVR).Namespace(namespace).Create(context.Background(), ingressObj, v1.CreateOptions{})
	if err != nil {
		log.Printf("Failed to create Ingress  resource: %v", err)
		return fmt.Errorf("failed to create Ingress  resource: %w", err)
	}

	return nil
}
