![Tungsten](https://github.com/Montana/tungsten/assets/20936398/2b3fd18e-3275-48f3-8a63-df576c388315)

Tungsten is a Cloudflare worker that can continuously handle tasks (HTTP requests) to manage traffic in Argo Rollouts.

## Prerequisites

- Go 1.16 or higher
- Cloudflare API credentials
- Argo CD API credentials
- Argo Rollouts API credentials

## Installation

First clone the Tungsten repository: 

   ```bash
   git clone https://github.com/Montana/tungsten.git
   cd tungsten
   ```

## Add tungsten into your Go project

First install it through your CLI:

```go
go get -u github.com/Montana/tungsten
```
Then import it as a dependency in your `main.go`:

```go
import (
    "github.com/Montana/tungsten"
)
```
 Run the application:

   ```bash
   go run main.go
   ```
The server will start on port 8080, and an ngrok tunnel will be created. The tunnel URL will be printed in the console, important to remember to use the HTTP endpoint to manage Argo Rollouts traffic:

 - **Endpoint**: `/manage-traffic`
 - **Method**: `POST`
 - **Request Body**:

   ```json
        {
            "name": "rollout-name",
            "namespace": "rollout-namespace",
            "weight": 50
        }
      ```

## Running the Application with Docker

Build the Docker image:

   ```sh
    docker build -t tungsten .
   ```

 Run the Docker container:

   ```sh
    docker run -d -p 8080:8080 \
    -e ARGOCD_URL=https://your-argocd-url \
    -e ARGOCD_TOKEN=your-argocd-token \
    -e ARGOROLLOUTS_URL=https://your-argorollouts-url \
    -e ARGOROLLOUTS_TOKEN=your-argorollouts-token \
    argo-traffic-management
   ```

The server will start on port 8080 inside the container, and an ngrok tunnel will be created. The tunnel URL will be printed in the container logs. Check the container logs to get the ngrok tunnel URL:

   ```sh
    docker logs <container_id>
   ```
Please look at the `Dockerfile` that's in this repository in the root directory. 

## Flowchart 

This explains the flow of how ngrok interacts with Argo and routes traffic:

![Tungsten](https://github.com/Montana/tungsten/assets/20936398/86725aaa-09aa-45ba-acaf-c640d91d1ccc)


Let's move on to `curl` now.

## Using curl

Example `curl` command:

   ```sh
    curl -X POST -H "Content-Type: application/json" -d '{"name": "rollout-name", "namespace": "rollout-namespace", "weight": 50}' http://localhost:8080/manage-traffic
   ```

## Main Components

- **ArgoApplication**: Struct to hold Argo CD application metadata.
- **TrafficRouting**: Struct to define traffic routing for Argo Rollouts.
- **getArgoCDApplications**: Function to fetch Argo CD applications.
- **manageArgoRolloutsTraffic**: Function to manage Argo Rollouts traffic routing.
- **handleRequest**: HTTP handler for managing traffic.
- **startServer**: Function to start the HTTP server and ngrok tunnel.

## Modify your existing functions

Example modification in `getArgoCDApplications` function:

```go
func getArgoCDApplications() ([]ArgoApplication, error) {
    argoCDURL := os.Getenv("ARGOCD_URL")
    argoCDToken := os.Getenv("ARGOCD_TOKEN")

    if argoCDURL == "" || argoCDToken == "" {
        return nil, fmt.Errorf("ARGOCD_URL and ARGOCD_TOKEN must be set")
    }

    client := tungsten.NewClient() // Using tungsten to create a new client
    req, err := tungsten.NewRequest("GET", fmt.Sprintf("%s/api/v1/applications", argoCDURL), nil) // Using tungsten to create a new request
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
```

## Configuration

Configure the application by setting the following environment variables. You can set them in a Bash script or directly in your shell.

- `CLOUDFLARE_API_KEY`: Your Cloudflare API key.
- `CLOUDFLARE_EMAIL`: Your Cloudflare account email.
- `CLOUDFLARE_ZONE_NAME`: The name of the Cloudflare zone.
- `ARGOCD_URL`: The URL of your Argo CD instance.
- `ARGOCD_TOKEN`: The API token for your Argo CD instance.
- `ARGOROLLOUTS_URL`: The URL of your Argo Rollouts instance.
- `ARGOROLLOUTS_TOKEN`: The API token for your Argo Rollouts instance.

## Copyright

Creator and Maintainer: _Michael Allen Mendy_. (c) 2024.
