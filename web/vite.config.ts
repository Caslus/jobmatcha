import tailwindcss from "@tailwindcss/vite";
import { devtools } from "@tanstack/devtools-vite";

import { tanstackStart } from "@tanstack/react-start/plugin/vite";

import viteReact from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { fileURLToPath, URL } from "node:url";

const config = defineConfig({
	resolve: {
		tsconfigPaths: true,
		alias: {
			"@": fileURLToPath(new URL("./src", import.meta.url)),
			"#": fileURLToPath(new URL("./src", import.meta.url)),
		},
	},
	plugins: [
		devtools(),
		tailwindcss(),
		tanstackStart({
			spa: {
				enabled: true,
				prerender: {
					outputPath: "/index.html",
				},
			},
		}),
		viteReact(),
	],
	server: {
		proxy: {
			"/api": "http://localhost:8181",
		},
		watch: {
			ignored: ["**/.*tmp.*"],
		},
	},
});

export default config;
