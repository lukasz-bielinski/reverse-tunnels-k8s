package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func CreateService(createRequestName string, edgeClusterInfoPort int) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	namespace := "chisel-server"

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      createRequestName,
			Namespace: namespace,
			Annotations: map[string]string{
				"traefik.ingress.kubernetes.io/service.serversscheme":    "https",
				"traefik.ingress.kubernetes.io/service.serverstransport": "chisel-server-tunnel-transport@kubernetescrd",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "chisel-server",
			},
			Ports: []corev1.ServicePort{
				{
					Name:       fmt.Sprintf("chisel-client-%s", createRequestName),
					Protocol:   corev1.ProtocolTCP,
					Port:       int32(edgeClusterInfoPort),
					TargetPort: intstr.FromInt(edgeClusterInfoPort),
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	_, err = clientset.CoreV1().Services(namespace).Create(context.Background(), service, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}
