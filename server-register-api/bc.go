package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/exec"
    "math/rand"
    "time"
    "github.com/gorilla/mux"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/client-go/dynamic"
    "k8s.io/client-go/rest"
)

type CreateRequest struct {
    EdgeClusterName string `json:"edge_cluster_name"`
    PortName        string `json:"port_name"`
}


func main() {
    router := mux.NewRouter()
    router.HandleFunc("/create", BasicAuth(CreateHandler)).Methods("POST")

    httpPort := "8080"
    log.Printf("Starting server on port %s...", httpPort)
    log.Fatal(http.ListenAndServe(":"+httpPort, router))
}

func BasicAuth(handler http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        username, password, ok := r.BasicAuth()

        if !ok || username != os.Getenv("USERNAME") || password != os.Getenv("PASSWORD") {
            w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
            http.Error(w, "Unauthorized.", http.StatusUnauthorized)
            return
        }

        handler(w, r)
    }
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
    var createRequest CreateRequest
    err := json.NewDecoder(r.Body).Decode(&createRequest)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    outputFile := fmt.Sprintf("%s.yaml", createRequest.EdgeClusterName)
    cmd := exec.Command("bash", "/app/generate-manifests.sh", createRequest.EdgeClusterName)
    err = cmd.Run()
    if err != nil {
        log.Printf("Error while executing script: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    fmt.Fprintln(w, "Script executed successfully.")
    fmt.Fprintf(w, "Your manifest can be downloaded from http://192.168.1.172/%s\n", outputFile)
}

func generateUniquePort(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource) (int, error) {
    rand.Seed(time.Now().UnixNano())
    for i := 0; i < 100; i++ {
        port := rand.Intn(20001) + 10000
        edgeClusterInfos, err := dynamicClient.Resource(gvr).List(context.Background(), metav1.ListOptions{})
        if err != nil {
            return 0, err
        }

        portExists := false
        for _, item := range edgeClusterInfos.Items {
            spec := item.Object["spec"].(map[string]interface{})
            if int(spec["port"].(int64)) == port {
                portExists = true
                break
            }
        }

        if !portExists {
            return port, nil
        }
    }
    return 0, fmt.Errorf("failed to generate a unique port number")
}

func createEdgeClusterInfo(edgeClusterName string) error {
    config, err := rest.InClusterConfig()
    if err != nil {
        return err
    }

    dynamicClient, err := dynamic.NewForConfig(config)
    if err != nil {
        return err
    }

    gvr := schema.GroupVersionResource{
        Group:    "xddevelopment.com",
        Version:  "v1",
        Resource: "edgeclusterinfos",
    }

    port, err := generateUniquePort(dynamicClient, gvr)
    if err != nil {
        return err
    }

    edgeClusterInfo := &unstructured.Unstructured{
        Object: map[string]interface{}{
            "apiVersion": "xddevelopment.com/v1",
            "kind":       "EdgeClusterInfo",
            "metadata": map[string]interface{}{
                "name": edgeClusterName,
            },
            "spec": map[string]interface{}{
                "port": port,
            },
        },
    }

    _, err = dynamicClient.Resource(gvr).Create(context.Background(), edgeClusterInfo, metav1.CreateOptions{})
    return err
}
