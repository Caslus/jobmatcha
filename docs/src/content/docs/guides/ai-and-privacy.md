---
title: AI features and privacy
description: Use optional OpenRouter-powered resume parsing and tailoring responsibly.
---

AI is optional. Scanning, filtering, ranking, bookmarking, and scheduled scans work without an AI provider.

## What OpenRouter receives

When you enable OpenRouter and ask Jobmatcha to parse a resume, Jobmatcha sends the extracted resume text to the configured provider. When you tailor a resume to a role, it sends the stored resume information together with that role's title, company, location, and description.

Resume uploads accept PDF, Markdown, and plain-text files up to 10 MB. Extracted resume text is saved in your SQLite database before parsing, so an upload is retained if the provider request fails.

## What Jobmatcha stores

Your API key, resume data, profile, preferences, roles, scan history, and sessions are stored in the SQLite database in the Jobmatcha data volume. Jobmatcha does not apply application-level encryption to that data. Keep the volume private, include it in backups, and use encrypted storage when your threat model requires it.

Review OpenRouter's own terms and privacy practices before enabling the integration. You can replace the stored API key in Settings; deleting it through the UI is not currently supported.
