package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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

func main() {
	apiKey := os.Getenv("CLOUDFLARE_API_KEY")
	email := os.Getenv("CLOUDFLARE_EMAIL")
	zoneName := os.Getenv("CLOUDFLARE_ZONE_NAME")

	if apiKey == "" || email == "" || zoneName == "" {
		log.Fatal("CLOUDFLARE_API_KEY, CLOUDFLARE_EMAIL, and CLOUDFLARE_ZONE_NAME must be set")
	}

	api, err := cloudflare.New(apiKey, email)
	if err != nil {
		log.Fatal(err)
	}

	zoneID, err := api.ZoneIDByName(zoneName)
	if err != nil {
		log.Fatal(err)
	}

	records, err := api.DNSRecords(zoneID, cloudflare.DNSRecord{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Cloudflare DNS Records:")
	for _, record := range records {
		fmt.Printf("Record: %s - Type: %s - Content: %s\n", record.Name, record.Type, record.Content)
	}

	applications, err := getArgoCDApplications()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nArgo CD Applications:")
	for _, app := range applications {
		fmt.Printf("Application: %s\n", app.Metadata.Name)
	}

	trafficRouting := TrafficRouting{
		Name:      "example-rollout",
		Namespace: "default",
		Weight:    50, // Example weight, adjust as needed
	}
	err = manageArgoRolloutsTraffic(trafficRouting)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nArgo Rollouts traffic managed successfully.")
}
