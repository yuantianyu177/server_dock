# ServerDock

Web-based server management system for managing remote servers via SSH, Docker containers, images, volumes, and container application/approval workflows with email notifications.

## Tech Stack

- **Backend:** Go + Gin + GORM + SQLite
- **Frontend:** Vue 3 + Vite
- **Deployment:** Docker Compose + Nginx reverse proxy

## Quick Start

### 1. Clone

```bash
git clone <repo-url>
cd ServerDock
```

### 2. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` and change `SECRET_KEY` and `SSH_CREDENTIAL_KEY`. Generate random keys:

```bash
# Generate SSH_CREDENTIAL_KEY (exactly 32 characters)
openssl rand -base64 24

# Generate SECRET_KEY
openssl rand -base64 32
```

### 3. Deploy with Docker Compose

```bash
docker compose up -d --build
```

Access the application at `http://<your-ip>:8080`.

### 4. Login

Use the admin credentials configured above (default: `admin` / `admin123`).

## Development

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

- Frontend (hot-reload): `http://localhost:3000`
- Backend API: `http://localhost:8000`

## Architecture

```
                    ┌──────────┐
   Browser ──────── │  Nginx   │ :8080
                    │ (frontend)│
                    └────┬─────┘
                         │ /api/*
                    ┌────▼─────┐
                    │  Go API  │ :8000
                    │ (backend)│
                    └────┬─────┘
                         │ SSH
                    ┌────▼─────┐
                    │  Remote  │
                    │ Servers  │
                    └──────────┘
```

## Features

- **Server Management** — Add/edit/delete remote servers, test SSH connections
- **Container Management** — Create/start/stop/restart/delete Docker containers with automatic port allocation
- **Image Management** — List/pull/delete remote Docker images, mark images as available for users
- **Volume Management** — Create/list/delete Docker volumes
- **Web Terminal** — Interactive SSH terminal to servers and containers via WebSocket
- **Application Workflow** — Public users submit container requests, admins approve/reject with email notifications
- **System Config** — Port range, extra ports, volume mount path, SMTP settings, all configurable via UI

## API Overview

| Group | Prefix | Auth |
|-------|--------|------|
| Auth | `/api/auth/*` | No (login) / Yes |
| Servers | `/api/servers/*` | Yes |
| Images (DB) | `/api/images/*` | Yes |
| Images (Remote) | `/api/servers/:id/images/*` | Yes |
| Containers | `/api/servers/:id/containers/*` | Yes |
| Volumes | `/api/servers/:id/volumes/*` | Yes |
| Applications (Public) | `/api/applications/public/*` | No |
| Applications (Admin) | `/api/applications/*` | Yes |
| Config | `/api/config/*` | Yes |
| Terminal (WS) | `/api/terminal/*` | Yes (query param) |

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `SECRET_KEY` | Yes | — | JWT signing key |
| `SSH_CREDENTIAL_KEY` | Yes | — | AES-256-GCM encryption key (32 chars) |
| `DATABASE_URL` | No | `data/serverdock.db` | SQLite database path |
| `DEFAULT_ADMIN_USERNAME` | No | `admin` | Initial admin username |
| `DEFAULT_ADMIN_PASSWORD` | No | `admin123` | Initial admin password |
| `DEBUG` | No | `false` | Enable debug mode |
| `PORT` | No | `8080` | Host port (compose) |

## License

MIT
