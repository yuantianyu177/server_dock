# ServerDock

ServerDock is a web console for managing remote servers, Docker containers, images, volumes, SSH terminals, and container application workflows.

## One-command deployment

Requirements: Docker with the Compose plugin, and OpenSSL.

```bash
git clone <repo-url>
cd ServerDock
./scripts/deploy.sh <hostname-or-ip>
```

For a server behind NAT or a campus firewall, create a remotely managed Cloudflare Tunnel, set its public hostname to `http://frontend:80`, and run:

```bash
./scripts/deploy.sh --cloudflare-tunnel serverdock.example.com
```

The script asks for the tunnel token without displaying it. It then automatically:

- creates `.env` with random JWT and SSH encryption keys;
- creates a SAN-enabled self-signed HTTPS certificate;
- builds and starts the Docker Compose services;
- reuses existing secrets, certificates, and data when run again.

The initial administrator account is `admin` / `admin123`. Change its password immediately after signing in. The production frontend exposes HTTPS only, on port 443 by default.

With Cloudflare Tunnel, Cloudflare manages the browser-facing trusted certificate and no inbound port is required. Without Tunnel, the generated certificate is self-signed and must be trusted on each client device or replaced with a CA-issued certificate.

## Configuration

Deployment settings are stored in the generated `.env`. Keep this file private and do not delete it: `SSH_CREDENTIAL_KEY` is required to decrypt stored server credentials.

To use non-standard ports on the first deployment:

```bash
HTTPS_PORT=8443 ./scripts/deploy.sh serverdock.example.com
```

To use your own certificate without Cloudflare Tunnel, update these values in `.env` and run `./scripts/deploy.sh` again:

```dotenv
PUBLIC_URL=https://serverdock.example.com
TLS_CERT_PATH=/absolute/path/to/fullchain.pem
TLS_KEY_PATH=/absolute/path/to/privkey.pem
```

## Development

After running `./scripts/deploy.sh localhost` once, start the hot-reload stack with:

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

- Frontend: `https://localhost:3000`
- Backend API: `http://localhost:8000`

Backend tests and frontend builds:

```bash
cd backend && mise exec -- go test ./...
cd frontend && mise exec -- npm run build
```

## Main features

- Remote server, Docker container, image, and volume management
- Browser-based SSH and container terminals over WSS
- Public container applications with administrator approval and email notifications
- Configurable ports, mounts, SMTP, and administrator credentials

## License

MIT
