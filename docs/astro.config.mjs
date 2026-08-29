// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

// https://astro.build/config
export default defineConfig({
	integrations: [
		starlight({
			title: "Jobmatcha",
			description: "Documentation for operating and using Jobmatcha.",
			customCss: ["./src/styles/custom.css"],
			social: [
				{
					icon: "github",
					label: "GitHub",
					href: "https://github.com/Caslus/jobmatcha",
				},
			],
			sidebar: [
				{ label: "Overview", items: [{ label: "Introduction", slug: "index" }] },
				{
					label: "Get started",
					items: [
						{ label: "Docker Compose", slug: "get-started/docker-compose" },
						{ label: "Your first scan", slug: "get-started/first-scan" },
						{
							label: "Keep it running",
							items: [
								{ label: "Deployment", slug: "get-started/operating/deployment" },
								{ label: "Upgrades and backups", slug: "get-started/operating/upgrades-and-backups" },
								{ label: "Runtime configuration", slug: "get-started/operating/runtime-configuration" },
								{ label: "Troubleshooting", slug: "get-started/operating/troubleshooting" },
							],
						},
					],
				},
				{
					label: "Use Jobmatcha",
					items: [
						{ label: "Relevance", slug: "guides/relevance" },
						{ label: "Scanning", slug: "guides/scanning" },
						{ label: "AI and privacy", slug: "guides/ai-and-privacy" },
					],
				},
				{
					label: "Develop",
					items: [{ label: "Development", slug: "developers/development" }],
				},
			],
			editLink: {
				baseUrl: "https://github.com/Caslus/jobmatcha/edit/main/docs/",
			},
		}),
	],
});
