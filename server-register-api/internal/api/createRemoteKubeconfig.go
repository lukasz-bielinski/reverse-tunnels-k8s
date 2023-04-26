package api

import (
	"chisel-api/internal/k8s"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func CreateRemoteKubeconfig(w http.ResponseWriter, r *http.Request) {
	var createRequest CreateRequest
	err := json.NewDecoder(r.Body).Decode(&createRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	for {
		_, err := k8s.GetEdgeClusterInfo(createRequest.EdgeClusterName)
		if err == nil {
			break
		}
		if time.Since(startTime) > maxWaitTime {
			http.Error(w, "Timed out waiting for custom resource to be created", http.StatusInternalServerError)
			return
		}
		time.Sleep(1 * time.Second)
	}
	edgeClusterInfo, err := k8s.GetEdgeClusterInfo(createRequest.EdgeClusterName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = k8s.CreateRemoteServiceAccount(createRequest.EdgeClusterName, edgeClusterInfo.Port, createRequest.EdgeClusterName+"-token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = k8s.CreateRemoteClusterRoleBinding(createRequest.EdgeClusterName, edgeClusterInfo.Port, createRequest.EdgeClusterName+"-token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := k8s.CreateRemoteTokenForUser(createRequest.EdgeClusterName, edgeClusterInfo.Port, createRequest.EdgeClusterName+"-token")
	if err != nil {
		fmt.Printf("Error creating remote token for user: %v\n", err)
		return
	}

	chiselTunnelDomain := os.Getenv("CHISEL_TUNNEL_DOMAIN")
	if chiselTunnelDomain == "" {
		chiselTunnelDomain = "chisel-tunnel.lan" // Use a default value if the environment variable is not set
	}
	// Invoke GenerateKubeConfig
	kubeConfig := k8s.GenerateKubeConfig(createRequest.EdgeClusterName, chiselTunnelDomain, edgeClusterInfo.ExposeName, token)

	// Write the generated kubeconfig to a file
	err = k8s.WriteKubeConfigToFile(kubeConfig, createRequest.EdgeClusterName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	chiselApiServer := os.Getenv("CHISEL_API_SERVER")
	if chiselApiServer == "" {
		chiselApiServer = "http://192.168.1.172" // Use a default value if the environment variable is not set
	}
	// Send a success response
	w.WriteHeader(http.StatusOK)

	response := fmt.Sprintf("\nRemote kubeconfig created successfully\nYour can download your kubeconfig here: %s/%s-kubeconfig.yaml", chiselApiServer, createRequest.EdgeClusterName)


	fmt.Fprint(w, response)
}
