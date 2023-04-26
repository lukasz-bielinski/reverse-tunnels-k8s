package api

import (
	"chisel-api/internal/k8s"
	"encoding/json"
	"net/http"
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

	// You can add more code here to perform other tasks related to creating the remote kubeconfig.

	// Send a success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Remote kubeconfig created successfully"))
}
