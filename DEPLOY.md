# Deploy

How to install this on a Linux server. This guide assumes the server already has a working nginx and a valid Let's Encrypt certificate for the domain you will serve from. We only add a reverse proxy block that forwards to the app container.

The app has no authentication and is wide open by design. It must never be published directly; nginx (with TLS, and any access control you add) is the only thing that should face the internet. The container binds to `127.0.0.1` so it is unreachable except through the proxy.

## Prerequisites

- Docker Engine with the Docker Compose plugin.
- nginx already serving your domain over HTTPS (Let's Encrypt certificate in place).
- An OpenRouter API key.
- Replays already on the server (or a directory you will copy them into).

## 1. Get the code

```bash
sudo git clone <repo-url> /opt/starcraft2
cd /opt/starcraft2
```

## 2. Configure

```bash
cp .env.sample .env
```

Edit `.env`:

- `ENVIRONMENT=production` - builds the binary once at container start and runs it (no air, no file watching).
- `OPENROUTER_API_KEY` - your key.
- `POSTGRES_PASSWORD` - a strong password.
- `PLAYERS` - tracked player names.
- `REPLAYS_HOST` - absolute path to the replay directory on this server.
- `PORT` - host port for the UI (stays bound to `127.0.0.1`). Default `8080`.

## 3. Start the stack

```bash
docker compose up -d --build
```

This starts `postgres` and `ui`. On first boot the database initializes from `sqlc/schema.sql`. The `ui` container builds the Go binary at startup and serves on `127.0.0.1:${PORT}`.

Import replays:

```bash
docker compose run --rm cli -action ingest
```

Verify the app is up locally:

```bash
curl -sI http://127.0.0.1:8080 | head -1
```

## 4. nginx reverse proxy

Add a `location` block to the existing HTTPS server block for your domain (adjust the port if you changed `PORT`):

```nginx
location / {
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    # The agent loop can take a while; allow long responses.
    proxy_read_timeout 130s;
}
```

The app's overall request deadline is 120s, so a `proxy_read_timeout` slightly above that avoids nginx cutting off a slow answer.

Reload nginx:

```bash
sudo nginx -t && sudo systemctl reload nginx
```

Your domain now serves the UI over the existing TLS certificate.

## Operating

```bash
docker compose ps                 # status
docker compose logs -f ui         # follow the UI logs
docker compose pull && docker compose up -d --build   # not needed; rebuild after pulling code
git pull && docker compose up -d --build              # deploy a new version
docker compose run --rm cli -action statistics        # row counts per table
```

Reset the database (drops all games and saved chats):

```bash
docker compose exec -T postgres psql -U postgres -d starcraft2 < sqlc/schema.sql
docker compose run --rm cli -action ingest
```

Postgres data lives in the bind-mounted `./postgres` directory; back that up to preserve games and chats.
