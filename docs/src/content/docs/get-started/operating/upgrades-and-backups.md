---
title: Upgrades and backups
description: Protect the Jobmatcha data volume and update predictably.
---

## Upgrade a published image

Back up the data volume, then pull and start the latest published image:

```bash
docker compose pull
docker compose up -d
```

Database migrations run when Jobmatcha starts. Confirm the updated instance works as expected before removing any backup. The published-image Compose example uses `ghcr.io/caslus/jobmatcha:latest`; pin the image to a specific published version if you need a controlled release cadence.

## Upgrade a source checkout

The repository's supplied Compose file builds the checked-out source. Back up the data volume, then update the checkout and rebuild the service:

```bash
git pull --ff-only
docker compose build
docker compose up -d
```

## Back up the data volume

Stop Jobmatcha before making a consistent backup, then archive the volume:

```bash
docker compose stop jobmatcha
docker run --rm --volumes-from "$(docker compose ps -q jobmatcha)" -v "$PWD":/backup alpine tar czf /backup/jobmatcha-data.tgz -C /data .
docker compose start jobmatcha
```

This command obtains the volume from the Compose service, so it does not depend on the Compose project name. Store the resulting archive outside the host and protect it as sensitive data: it contains the SQLite database and potentially API keys and resume information.

To restore, stop Jobmatcha, extract a trusted backup into the intended data volume, then start the service again. Test restore procedures before depending on them in an emergency.
