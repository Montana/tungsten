# tungsten

Tungsten is a Go application that interacts with Cloudflare, Argo CD, and Argo Rollouts. It lists DNS records from Cloudflare, applications from Argo CD, and manages traffic routing in Argo Rollouts.

## Prerequisites

- Go 1.16 or higher
- Cloudflare API credentials
- Argo CD API credentials
- Argo Rollouts API credentials

## Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/yourusername/tungsten.git
   cd tungsten
   ```

## Configuration

Configure the application by setting the following environment variables. You can set them in a Bash script or directly in your shell.

- `CLOUDFLARE_API_KEY`: Your Cloudflare API key.
- `CLOUDFLARE_EMAI`L: Your Cloudflare account email.
- `CLOUDFLARE_ZONE_NAME`: The name of the Cloudflare zone.
- `ARGOCD_URL`: The URL of your Argo CD instance.
- `ARGOCD_TOKEN`: The API token for your Argo CD instance.
- `ARGOROLLOUTS_URL`: The URL of your Argo Rollouts instance.
- `ARGOROLLOUTS_TOKEN`: The API token for your Argo Rollouts instance.

## Copyright

Michael Allen Mendy. (c) 2024.
