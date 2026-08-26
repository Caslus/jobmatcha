import { defineConfig } from "vitest/config";

export default defineConfig({
	test: {
		environment: "jsdom",
		setupFiles: ["./src/test/setup.ts"],
		include: ["src/**/*.test.{ts,tsx}"],
		coverage: {
			provider: "v8",
			reporter: ["text", "html"],
			exclude: ["src/types/api.gen.ts", "src/routeTree.gen.ts", "src/test/**", "*.config.*"],
		},
	},
});
