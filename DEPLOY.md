# Deploy

Install on a Linux server that already has nginx and a Let's Encrypt certificate for the domain. The app has no authentication; nginx is the only public entry point and the `GO_PORT` should be firewalled.

## Steps

1. Install prerequisites:

   ```bash
   sudo apt install -y supervisor
   ```

   (Docker Engine + the Compose plugin must also be present.)

2. Clone the repo:

   ```bash
   sudo git clone <repo-url> /opt/starcraft2
   cd /opt/starcraft2
   ```

3. Configure (every variable is required, no defaults):

   ```bash
   cp .env.sample .env
   ```

   Set `GO_ENVIRONMENT=production`, `OPENROUTER_API_KEY`, a strong `POSTGRES_PASSWORD`, `GO_REPLAYS` (absolute path to replays on this server), and the rest.

4. Start the stack under supervisor:

   ```bash
   sudo cp supervisor.conf /etc/supervisor/conf.d/starcraft2.conf
   sudo supervisorctl reread
   sudo supervisorctl update
   ```

5. Import replays:

   ```bash
   docker compose run --rm cli -action ingest
   ```

6. Enable the nginx site (the file is preconfigured for `starcraft2.mahendrakalkura.com`):

   ```bash
   sudo ln -s /opt/starcraft2/nginx.conf /etc/nginx/sites-enabled/starcraft2.mahendrakalkura.com.conf
   sudo nginx -t
   sudo systemctl reload nginx
   ```

## Operating

```bash
sudo supervisorctl restart starcraft2               # restart the stack
git pull && sudo supervisorctl restart starcraft2   # deploy a new version
docker compose logs -f ui                           # follow logs
```

Reset the database (drops all games and saved chats):

```bash
docker compose exec -T postgres psql -U postgres -d starcraft2 < sqlc/schema.sql
docker compose run --rm cli -action ingest
```

Postgres data lives in `./postgres`; back it up to preserve games and chats.
