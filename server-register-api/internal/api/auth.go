package api

import (
	"context"
	"log"
	"net/http"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func BasicAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		if !ok || !authenticateUser(username, password) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}

		handler(w, r)
	}
}

func authenticateUser(username, password string) bool {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error getting in-cluster config: %v", err)
		return false
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
		return false
	}

	secret, err := clientset.CoreV1().Secrets(os.Getenv("NAMESPACE")).Get(context.Background(), "basic-auth-users", metav1.GetOptions{})
	if err != nil {
		log.Printf("Error retrieving secret: %v", err)
		return false
	}

	storedPassword, ok := secret.Data[username]
	if !ok {
		return false
	}

	return password == string(storedPassword)
}
