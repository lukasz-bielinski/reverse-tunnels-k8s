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
)

func CreateMiddleware(edgeClusterName, exposeName, namespace string) error {
	log.Printf("Creating Middleware: edgeClusterName=%s, exposeName=%s, namespace=%s\n", edgeClusterName, exposeName, namespace)

	middlewareObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "traefik.containo.us/v1alpha1",
			"kind":       "Middleware",
			"metadata": map[string]interface{}{
				"name":      edgeClusterName,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"stripPrefix": map[string]interface{}{
					"forceSlash": false,
					"prefixes":   []string{"/" + exposeName},
				},
			},
		},
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	middlewareGVR := schema.GroupVersionResource{
		Group:    "traefik.containo.us",
		Version:  "v1alpha1",
		Resource: "middlewares",
	}
	_, err = dynamicClient.Resource(middlewareGVR).Namespace(namespace).Create(context.Background(), middlewareObj, v1.CreateOptions{})
	if err != nil {
		log.Printf("Failed to create Middleware custom resource: %v", err)
		return fmt.Errorf("failed to create Middleware custom resource: %w", err)
	}

	log.Println("Middleware created successfully")
	return nil
}
