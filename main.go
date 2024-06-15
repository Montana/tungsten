package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

type ArgoApplication struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
}

type TrafficRouting struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Weight    int    `json:"weight"`
}

func getArgoCDApplications() ([]ArgoApplication, error) {
	argoCDURL := os.Getenv("ARGOCD_URL")
	argoCDToken := os.Getenv("ARGOCD_TOKEN")

	if argoCDURL == "" || argoCDToken == "" {
		return nil, fmt.Errorf("ARGOCD_URL and ARGOCD_TOKEN must be set")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/applications", argoCDURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", argoCDToken))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get Argo CD applications, status code: %d", resp.StatusCode)
	}

	var applications struct {
		Items []ArgoApplication `json:"items"`
	}
	err = json.NewDecoder(resp.Body).Decode(&applications)
	if err != nil {
		return nil, err
	}

	return applications.Items, nil
}

func manageArgoRolloutsTraffic(routing TrafficRouting) error {
	argoRolloutsURL := os.Getenv("ARGOROLLOUTS_URL")
	argoRolloutsToken := os.Getenv("ARGOROLLOUTS_TOKEN")

	if argoRolloutsURL == "" || argoRolloutsToken == "" {
		return fmt.Errorf("ARGOROLLOUTS_URL and ARGOROLLOUTS_TOKEN must be set")
	}

	client := &http.Client{}
	routingData, err := json.Marshal(routing)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/namespaces/%s/rollouts/%s/traffic", argoRolloutsURL, routing.Namespace, routing.Name), bytes.NewBuffer(routingData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", argoRolloutsToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to manage Argo Rollouts traffic, status code: %d", resp.StatusCode)
	}

	return nil
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var routing TrafficRouting
		err := json.NewDecoder(r.Body).Decode(&routing)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = manageArgoRolloutsTraffic(routing)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Argo Rollouts traffic managed successfully.\n")

	default:
		http.Error(w, "Unsupported request method.", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/manage-traffic", handleRequest)
	port := ":8080"
	fmt.Printf("Starting server on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
