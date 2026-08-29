---
title: Start with Docker Compose
description: Run a private Jobmatcha instance in a few commands.
---

## Prerequisites

Install Docker with the Docker Compose plugin. Jobmatcha stores its database, sessions, and bootstrap password in a named Docker volume.

## Start Jobmatcha

Clone the source checkout and build and start the service from its root:

```bash
git clone https://github.com/Caslus/jobmatcha.git
cd jobmatcha
docker compose up -d
docker compose exec jobmatcha cat /data/initial-password
```

Open [http://localhost:8181](http://localhost:8181), sign in with the generated password, and complete onboarding. Change the password when prompted; after a successful password change, Jobmatcha removes the `initial-password` file.

> [!CAUTION]
> Treat the initial password like any other credential. Do not paste it into tickets, shell history shared with others, or public logs.

The `jobmatcha-data` volume survives container recreation. Do not remove it unless you intentionally want to erase the application data.

## Next steps

- [Build your first shortlist](/get-started/first-scan/)
- [Understand relevance preferences](/guides/relevance/)
- [Deploy behind HTTPS](/get-started/operating/deployment/)
