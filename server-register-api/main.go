package main

import (
    "log"
    "net/http"

    "chisel-api/internal/api"
    "chisel-api/internal/k8s"
    "github.com/gorilla/mux"
)

func main() {
    err := k8s.EnsureEdgeClusterInfoCRD()
    if err != nil {
        log.Fatalf("Failed to ensure EdgeClusterInfo CRD: %v", err)
    }

    router := mux.NewRouter()
    router.HandleFunc("/create", api.BasicAuth(api.CreateHandler)).Methods("POST")

    httpPort := "8080"
    log.Printf("Starting server on port %s...", httpPort)
    log.Fatal(http.ListenAndServe(":"+httpPort, router))
}
