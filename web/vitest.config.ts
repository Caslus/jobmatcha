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
			reporter: ["text", "html"],
			exclude: ["src/types/api.gen.ts", "src/routeTree.gen.ts", "src/test/**", "*.config.*"],
		},
	},
});
