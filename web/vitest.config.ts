import { defineConfig } from "vitest/config";
import { fileURLToPath } from "node:url";

export default defineConfig({
	resolve: {
		alias: {
			"@": fileURLToPath(new URL("./src", import.meta.url)),
			"#": fileURLToPath(new URL("./src", import.meta.url)),
		},
	},
	test: {
		environment: "jsdom",
		env: {
			VITE_API_BASE_URL: "http://localhost/api",
		},
		setupFiles: ["./src/test/setup.ts"],
		include: ["src/**/*.test.{ts,tsx}"],
		coverage: {
			provider: "v8",
			reporter: ["text", "html", "lcov", "json-summary"],
			include: ["src/**/*.{ts,tsx}"],
			exclude: [
				"src/**/*.test.{ts,tsx}",
				"src/test/**",
				"src/types/api.gen.ts",
				"src/routeTree.gen.ts",
				"src/router.tsx",
				"src/integrations/tanstack-query/devtools.tsx",
				"*.config.*",
			],
		},
		pool: "threads",
		maxWorkers: 1,
	},
});
