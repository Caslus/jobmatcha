# Jobmatcha documentation

This directory is the standalone [Starlight](https://starlight.astro.build/) documentation site. It is not part of the Jobmatcha production container.

```bash
mise install
pnpm --dir docs install --frozen-lockfile
mise run docs:dev
```

Use `mise run docs:build` to validate and generate the production site in
`dist/`. PNPM 10.23.0 is provisioned by Mise; dependency installation remains
an explicit command. Root dotenv files are not loaded automatically by the
application or Mise.
