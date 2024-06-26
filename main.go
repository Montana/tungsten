// Author and Maintainer: Michael Allen Mendy (c) 2024 for Travis CI.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/philippgille/gokrok"
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

func startNgrokTunnel() (string, error) {
	opts := gokrok.Options{
		Addr: ":8080",
	}
	tunnel, err := gokrok.Start(opts)
	if err != nil {
		return "", fmt.Errorf("failed to start ngrok tunnel: %v", err)
	}
	defer tunnel.Stop()

	return tunnel.URL(), nil
}

func startNginxReverseProxy() error {
	// example command to start nginx reverse proxy
	cmd := exec.Command("nginx", "-c", "/etc/nginx/nginx.conf")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to start nginx reverse proxy: %v", err)
	}
	return nil
}

func configureSmallstep() error {
	// adjust according to your actual smallstep configuration and environment
	cmd := exec.Command("step", "certificates", "renew", "--daemon")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to configure Smallstep: %v", err)
	}
	return nil
}

func main() {
	// choose between ngrok, nginx reverse proxy, and smallstep based on an environment variable
	proxyOption := os.Getenv("PROXY_OPTION")

	switch proxyOption {
	case "ngrok":
		url, err := startNgrokTunnel()
		if err != nil {
			log.Fatalf("Error starting ngrok tunnel: %v", err)
		}
		fmt.Printf("ngrok tunnel started at %s\n", url)

	case "nginx":
		err := startNginxReverseProxy()
		if err != nil {
			log.Fatalf("Error starting nginx reverse proxy: %v", err)
		}
		fmt.Printf("nginx reverse proxy started\n")

	case "smallstep":
		err := configureSmallstep()
		if err != nil {
			log.Fatalf("Error configuring Smallstep: %v", err)
		}
		fmt.Printf("Smallstep certificate management configured\n")

	default:
		log.Fatalf("Unsupported proxy option: %s", proxyOption)
	}

	http.HandleFunc("/manage-traffic", handleRequest)
	port := ":8080"
	fmt.Printf("Starting server on port %s\n", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
