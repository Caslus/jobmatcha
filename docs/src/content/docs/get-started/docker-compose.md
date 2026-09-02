---
title: Start with Docker Compose
description: Run a private Jobmatcha instance in a few commands.
---

## Prerequisites

Install Docker with the Docker Compose plugin. Jobmatcha stores its database, sessions, and bootstrap password in a named Docker volume.

## Start Jobmatcha

Create a `compose.yaml` with the published Jobmatcha image:

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

Start it and read the generated password:

```bash
docker compose up -d
docker compose exec jobmatcha cat /data/initial-password
```

The `pull_policy` fetches [`ghcr.io/caslus/jobmatcha:latest`](https://github.com/Caslus/jobmatcha/pkgs/container/jobmatcha), so Docker does not build Jobmatcha from source. Open [http://localhost:8181](http://localhost:8181), sign in with the generated password, and complete onboarding. Change the password when prompted; after a successful password change, Jobmatcha removes the `initial-password` file.

To stop the application without deleting data, run `docker compose down`. Start it again with `docker compose up -d`.

> [!CAUTION]
> Treat the initial password like any other credential. Do not paste it into tickets, shell history shared with others, or public logs.

The `jobmatcha-data` volume survives container recreation. Do not remove it unless you intentionally want to erase the application data.

## Build from source instead

For local development, clone the repository and use its supplied `compose.yaml`. That configuration builds the image from the checkout:

```bash
git clone https://github.com/Caslus/jobmatcha.git
cd jobmatcha
docker compose up -d
```

## Next steps

- [Build your first shortlist](/get-started/first-scan/)
- [Understand relevance preferences](/guides/relevance/)
- [Deploy behind HTTPS](/get-started/operating/deployment/)
