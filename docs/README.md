# Jobmatcha documentation

This directory is the standalone [Starlight](https://starlight.astro.build/) documentation site. It is not part of the Jobmatcha production container.

```bash
corepack enable
pnpm install
pnpm dev
```

Use `pnpm build` to validate and generate the production site in `dist/`. From the repository root, the equivalent commands are `mise run "[Docs] Run dev"` and `mise run "[Docs] Build"`.
