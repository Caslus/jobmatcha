import { execFile, spawn } from "node:child_process";
import { promisify } from "node:util";
import { readFile, rm, mkdir } from "node:fs/promises";
import { resolve } from "node:path";
import { setTimeout as delay } from "node:timers/promises";
import { chromium } from "@playwright/test";

const webDir = resolve(import.meta.dirname, "..");
const repoDir = resolve(webDir, "..");
const fixtureDir = resolve(webDir, ".e2e-tmp-readme");
const imageDir = resolve(repoDir, "docs", "public", "images");
const baseURL = "http://127.0.0.1:8183";
const execFileAsync = promisify(execFile);

async function waitForHealthyServer() {
	for (let attempt = 0; attempt < 60; attempt += 1) {
		try {
			const response = await fetch(`${baseURL}/api/health`);
			if (response.ok) return;
		} catch {
			// The Go fixture is still starting.
		}
		await delay(500);
	}
	throw new Error("Timed out waiting for the documentation fixture.");
}

async function stopServer(server) {
	if (server.exitCode !== null || server.signalCode !== null) return;
	const closed = new Promise((resolveClose) => server.once("close", resolveClose));
	server.kill("SIGTERM");
	await Promise.race([closed, delay(5_000)]);
	if (server.exitCode === null && server.signalCode === null) {
		server.kill("SIGKILL");
		await closed;
	}
}

await rm(fixtureDir, { recursive: true, force: true });
await mkdir(imageDir, { recursive: true });
const fixtureBinary = resolve(fixtureDir, process.platform === "win32" ? "jobmatcha-e2e.exe" : "jobmatcha-e2e");

await execFileAsync("go", ["build", "-o", fixtureBinary, "./cmd/e2e"], {
	cwd: resolve(repoDir, "server"),
});

const server = spawn(
	fixtureBinary,
	[
		"-port",
		"8183",
		"-static-dir",
		"../web/dist/client",
		"-work-dir",
		"../web/.e2e-tmp-readme",
	],
	{ cwd: resolve(repoDir, "server"), stdio: "inherit" },
);

try {
	await waitForHealthyServer();
	const password = (await readFile(resolve(fixtureDir, "bootstrap-password"), "utf8")).trim();
	const browser = await chromium.launch();
	const page = await browser.newPage({ viewport: { width: 1440, height: 960 }, deviceScaleFactor: 1 });
	await page.addInitScript(() => {
		localStorage.setItem("jobmatcha_panel_widths", JSON.stringify({ left: 23, right: 39 }));
	});

	await page.goto(baseURL, { waitUntil: "networkidle" });
	await page.getByPlaceholder("Password").fill(password);
	await page.getByRole("button", { name: "Sign in" }).click();
	await page.waitForURL(`${baseURL}/dashboard`);
	await page.getByRole("heading", { name: "Roles" }).waitFor();
	await page.getByRole("button", { name: /Software Engineer, Ads Backend/ }).first().click();
	await page.getByRole("heading", { name: "Software Engineer, Ads Backend" }).waitFor();
	await page.screenshot({ path: resolve(imageDir, "dashboard.png") });

	await browser.close();
} finally {
	await stopServer(server);
	await rm(fixtureDir, { recursive: true, force: true });
}
