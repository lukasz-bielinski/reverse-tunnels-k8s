// k8s/k8s.go

package k8s

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/rs/xid"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

const (
	minPort       = 10000
	maxPort       = 30000
	exposeNameLen = 10
)

type EdgeClusterInfo struct {
	Name       string `json:"name"`
	Port       int    `json:"port"`
	ExposeName string `json:"expose_name"`
	Namespace  string `json:"namespace"`
}

func CreateEdgeClusterInfo(edgeClusterName string) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	edgeClusterInfoGVR := schema.GroupVersionResource{
		Group:    "xddevelopment.com",
		Version:  "v1",
		Resource: "edgeclusterinfos",
	}

	namespace := "chisel-server" // Replace with the namespace you want to use

	port, err := generateUniquePort(dynamicClient, edgeClusterInfoGVR, namespace)
	if err != nil {
		return err
	}

	exposeName := generateRandomString(exposeNameLen)

	edgeClusterInfo := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "xddevelopment.com/v1",
			"kind":       "EdgeClusterInfo",
			"metadata": map[string]interface{}{
				"name": edgeClusterName,
			},
			"spec": map[string]interface{}{
				"port":        port,
				"expose_name": exposeName,
			},
		},
	}

	_, err = dynamicClient.Resource(edgeClusterInfoGVR).Namespace(namespace).Create(context.Background(), edgeClusterInfo, v1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func GetEdgeClusterInfo(edgeClusterName string) (EdgeClusterInfo, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return EdgeClusterInfo{}, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return EdgeClusterInfo{}, err
	}

	edgeClusterInfoGVR := schema.GroupVersionResource{
		Group:    "xddevelopment.com",
		Version:  "v1",
		Resource: "edgeclusterinfos",
	}

	namespace := "chisel-server" // Replace with the namespace you want to use

	// Get the custom resource with the specified name
	unstructuredEdgeClusterInfo, err := dynamicClient.Resource(edgeClusterInfoGVR).Namespace(namespace).Get(context.Background(), edgeClusterName, v1.GetOptions{})
	if err != nil {
		return EdgeClusterInfo{}, err
	}

	// Extract the information from the unstructured custom resource
	spec := unstructuredEdgeClusterInfo.Object["spec"].(map[string]interface{})
	port := int(spec["port"].(int64))
	exposeName := spec["expose_name"].(string)

	// Create an EdgeClusterInfo instance
	edgeClusterInfo := EdgeClusterInfo{
		Port:       port,
		ExposeName: exposeName,
		Namespace:  namespace,
	}

	return edgeClusterInfo, nil
}

func generateUniquePort(dynamicClient dynamic.Interface, edgeClusterInfoGVR schema.GroupVersionResource, namespace string) (int, error) {
	rand.Seed(time.Now().UnixNano())
	attempts := 0

	for {
		port := rand.Intn(maxPort-minPort+1) + minPort
		existingEdgeClusterInfo, err := dynamicClient.Resource(edgeClusterInfoGVR).Namespace(namespace).List(context.Background(), v1.ListOptions{})

		if err != nil {
			return 0, err
		}

		portExists := false
		for _, item := range existingEdgeClusterInfo.Items {
			existingPort, found, err := unstructured.NestedInt64(item.Object, "spec", "port")

			if err != nil {
				return 0, err
			}

			if found && int(existingPort) == port {
				portExists = true
				break
			}
		}

		if !portExists {
			return port, nil
		}

		attempts++
		if attempts >= 10 {
			return 0, errors.New("failed to generate a unique port after 10 attempts")
		}
	}
}

func generateRandomString(length int) string {
	guid := xid.New()
	return strings.ToLower(guid.String()[:length])
}
