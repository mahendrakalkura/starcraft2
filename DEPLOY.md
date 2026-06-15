# Deploy

How to install this on a Linux server. This guide assumes the server already has a working nginx and a valid Let's Encrypt certificate for the domain you will serve from. The repo ships two config files for the server: `nginx.conf` (a site you symlink into nginx) and `supervisor.conf` (a supervisor program that runs the Docker stack).

The app has no authentication and is wide open by design. It must never face the internet directly; nginx (with TLS, and any access control you add) should be the only public entry point. The published port (`GO_PORT`) binds on all interfaces, so restrict it with a firewall and let nginx reach the app over the loopback address.

## Prerequisites

- Docker Engine with the Docker Compose plugin.
- supervisor (`apt install supervisor`).
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

Every variable in `.env` is required and there are no defaults; the app prints which are missing and exits if any is blank. The Go process reads `.env` directly (godotenv), so the names in `.env` are the names used in the code. Fill in all of them, and in particular:

- `GO_ENVIRONMENT=production` - builds the binary once at container start and runs it (no air, no file watching).
- `GO_PORT` - port the UI listens on and is published at (on all interfaces, so firewall it and let nginx reach it over loopback).
- `GO_PLAYERS` - tracked player names.
- `GO_REPLAYS` - absolute path to the replay directory on this server (bind-mounted at the same path in the container).
- `OPENROUTER_API_KEY` - your key.
- `OPENROUTER_MODEL` - a tool-calling model, e.g. `deepseek/deepseek-v4-flash`.
- `POSTGRES_DB`, `POSTGRES_HOST`, `POSTGRES_PASSWORD`, `POSTGRES_PORT`, `POSTGRES_USER` - the database settings (`POSTGRES_HOST` is the compose service name, `postgres`). Use a strong password. The app builds its connection string from these.

## 3. Run the stack under supervisor

`supervisor.conf` runs `docker compose up --build` from `/opt/starcraft2` and keeps it running across reboots and crashes. Install it (edit `directory` in the file if you cloned elsewhere):

```bash
sudo cp supervisor.conf /etc/supervisor/conf.d/starcraft2.conf
sudo supervisorctl reread
sudo supervisorctl update
sudo supervisorctl status starcraft2
```

The first start builds the image and initializes the database from `sqlc/schema.sql`. Follow progress with `tail -f /var/log/starcraft2.log`.

Import replays once the stack is up:

```bash
docker compose run --rm cli -action ingest
```

Verify the app is up locally:

```bash
curl -sI http://127.0.0.1:${GO_PORT} | head -1
```

## 4. nginx reverse proxy

`nginx.conf` is preconfigured for `starcraft2.mahendrakalkura.com` and uses certbot's shared `options-ssl-nginx.conf` and `ssl-dhparams.pem` (already on the server). It redirects HTTP to HTTPS and proxies HTTPS to the app on port 8080; its `proxy_read_timeout` is 130s, just above the app's 120s request deadline, so a slow agent answer is not cut off. Adjust the domain, the `proxy_pass` port, or the cert paths only if they differ, then symlink it into the enabled sites and reload:

```bash
sudo ln -s /opt/starcraft2/nginx.conf /etc/nginx/sites-enabled/starcraft2.mahendrakalkura.com.conf
sudo nginx -t && sudo systemctl reload nginx
```

Your domain now serves the UI over the existing TLS certificate.

## Operating

```bash
sudo supervisorctl restart starcraft2   # restart the stack
sudo supervisorctl stop starcraft2      # stop the stack
docker compose ps                       # container status
docker compose logs -f ui               # follow the UI logs
git pull && sudo supervisorctl restart starcraft2   # deploy a new version (rebuilds)
```

Reset the database (drops all games and saved chats):

```bash
docker compose exec -T postgres psql -U postgres -d starcraft2 < sqlc/schema.sql
docker compose run --rm cli -action ingest
```

Postgres data lives in the bind-mounted `./postgres` directory; back that up to preserve games and chats.
