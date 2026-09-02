<p align="center">
  <img src="web/src/assets/wordmark.svg" width="360" alt="jobmatcha" />
</p>

<p align="center">
  <strong>Turn a noisy job search into a shortlist worth opening.</strong><br />
A private, self-hosted job tracker that scans company career boards and ranks roles around what you want next.
</p>

<p align="center">
  <a href="https://github.com/Caslus/jobmatcha/actions/workflows/test.yml"><img src="https://github.com/Caslus/jobmatcha/actions/workflows/test.yml/badge.svg?branch=main" alt="Tests" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-a8d5a2" alt="MIT License" /></a>
</p>

![Jobmatcha dashboard showing ranked jobs and search preferences](docs/public/images/dashboard.png)

Jobmatcha starts with a seeded set of company career sites and turns them into an explainable, preference-based shortlist. You can discover supported boards from a careers page or manage companies and boards manually. The implemented scanners support Greenhouse and Workable. Your data lives in your own SQLite volume—there is no hosted account, analytics script, or required AI subscription.

## Start in two minutes

Create a `compose.yaml` containing:

```yaml
services:
  jobmatcha:
    image: ghcr.io/caslus/jobmatcha:latest
    pull_policy: always
    ports:
      - "8181:8181"
    environment:
      DB_PATH: /data/app.db
      COOKIE_SECURE: "false"
    volumes:
      - jobmatcha-data:/data
    restart: unless-stopped

volumes:
  jobmatcha-data:
```

Start Jobmatcha and read the generated password:

```bash
docker compose up -d
docker compose exec jobmatcha cat /data/initial-password
```

Docker Compose pulls the published image instead of building from source. Open [http://localhost:8181](http://localhost:8181), sign in with the generated password, and complete onboarding. Your data stays in the `jobmatcha-data` Docker volume.

## Documentation

The full documentation is a Starlight project in [`docs/`](docs/README.md). Until it has a public URL, GitHub renders the source guides directly:

- [Docker Compose and first scan](docs/src/content/docs/get-started/docker-compose.md)
- [Relevance, schedules, and AI privacy](docs/src/content/docs/guides/relevance.md)
- [Deployment, HTTPS, backups, and configuration](docs/src/content/docs/get-started/operating/deployment.md)
- [Local development and documentation assets](docs/src/content/docs/developers/development.md)

## Under the hood

- **Frontend:** React 19, TanStack Start/Router/Query, Tailwind CSS
- **Backend:** Go, Gin, GORM, SQLite
- **Delivery:** local multi-stage container build; the release workflow is prepared to publish multi-architecture images when releases are available.

## Community

Jobmatcha is MIT licensed and built in the open. Contributions, provider integrations, and documentation improvements are welcome.

- [Contributing guide](CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Security policy](SECURITY.md)
- [MIT License](LICENSE)
