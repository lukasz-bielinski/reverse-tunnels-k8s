package api

import (
	"chisel-api/internal/k8s"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const maxWaitTime = 20 * time.Second

type CreateRequest struct {
	EdgeClusterName string `json:"edge_cluster_name"`
}

var createRequestsTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "create_requests_total",
		Help: "Total number of create requests.",
	},
)

func init() {
	prometheus.MustRegister(createRequestsTotal)
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	createRequestsTotal.Inc()

	chiselApiServer := os.Getenv("CHISEL_API_SERVER")
	if chiselApiServer == "" {
		chiselApiServer = "http://192.168.1.172" // Use a default value if the environment variable is not set
	}

	chiselTunnelDomain := os.Getenv("CHISEL_TUNNEL_DOMAIN")
	if chiselTunnelDomain == "" {
		chiselTunnelDomain = "chisel-tunnel.lan" // Use a default value if the environment variable is not set
	}
	chiselRegisterDomain := os.Getenv("CHISEL_REGISTER_DOMAIN")
	if chiselRegisterDomain == "" {
		chiselRegisterDomain = "https://chisel-register.lan" // Use a default value if the environment variable is not set
	}
	chiselServerDomain := os.Getenv("CHISEL_SERVER_DOMAIN")
	if chiselServerDomain == "" {
		chiselServerDomain = "https://chisel-server.lan" // Use a default value if the environment variable is not set
	}
	var createRequest CreateRequest
	err := json.NewDecoder(r.Body).Decode(&createRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = k8s.CreateEdgeClusterInfo(createRequest.EdgeClusterName)
	if err != nil {
		http.Error(w, "Internal server error CreateEdgeClusterInfo", http.StatusInternalServerError)
		return
	}

	edgeClusterInfo, err := k8s.GetEdgeClusterInfo(createRequest.EdgeClusterName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//portStr := strconv.Itoa(edgeClusterInfo.Port)

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

	err = k8s.CreateMiddleware(createRequest.EdgeClusterName, edgeClusterInfo.ExposeName, edgeClusterInfo.Namespace)
	if err != nil {
		http.Error(w, "Internal server error CreateMiddleware", http.StatusInternalServerError)
		return
	}

	err = k8s.CreateService(createRequest.EdgeClusterName, edgeClusterInfo.Port)
	if err != nil {
		http.Error(w, "Internal server error CreateService", http.StatusInternalServerError)
		return
	}

	err = k8s.CreateIngress(createRequest.EdgeClusterName, edgeClusterInfo.ExposeName, edgeClusterInfo.Namespace, edgeClusterInfo.Port)
	if err != nil {
		http.Error(w, "Internal server error CreateIngress", http.StatusInternalServerError)
		return
	}

	//outputFile := fmt.Sprintf("%s.yaml", createRequest.EdgeClusterName)
	cmd := exec.Command("bash", "/app/generate-manifests.sh", createRequest.EdgeClusterName, strconv.Itoa(edgeClusterInfo.Port), chiselRegisterDomain, chiselServerDomain)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Printf("Error while executing script: %v", err)
		http.Error(w, "Internal server error bash script", http.StatusInternalServerError)
		return
	}



	response := fmt.Sprintf("\nYour manifest can be downloaded from %s/%s.yaml"+
		"\nYou can access your cluster under curl -k -H \"Authorization:Bearer $TOKEN\" -s %s/%s/api/v1/namespaces/kube-system/pods  | jq '.items[].metadata.name'\n", chiselApiServer, createRequest.EdgeClusterName, chiselTunnelDomain, edgeClusterInfo.ExposeName)

	fmt.Fprint(w, response)

}
