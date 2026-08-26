import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
	testDir: "./e2e",
	fullyParallel: true,
	forbidOnly: !!process.env.CI,
	reporter: process.env.CI ? "github" : "list",
	use: {
		baseURL: "http://127.0.0.1:8182",
		trace: "retain-on-failure",
	},
	projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
	webServer: {
		command:
			"cd ../server && go run ./cmd/e2e -port 8182 -static-dir ../web/dist/client -work-dir ../web/.e2e-tmp",
		url: "http://127.0.0.1:8182/api/health",
		reuseExistingServer: !process.env.CI,
		timeout: 120_000,
	},
});
