import { screen } from "@testing-library/react";
import { delay, HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";
import { DashboardPage } from "../routes/dashboard";
import { fixtures } from "./msw/fixtures";
import { server } from "./msw/server";
import { renderWithProviders } from "./render";

describe("dashboard route", () => {
	it("loads a role and lets the user open and close its detail panel", async () => {
		const { user } = renderWithProviders(<DashboardPage />);

		await user.click(await screen.findByText("Frontend Engineer"));
		expect(
			await screen.findByRole("heading", { name: "Frontend Engineer" }),
		).toBeVisible();
		expect(screen.getByText("Build excellent user experiences.")).toBeVisible();

		await user.click(
			screen.getByRole("button", { name: "Close role details" }),
		);
		expect(
			await screen.findByText("Select a role to view details"),
		).toBeVisible();
	});

	it("shows loading placeholders before roles have loaded", () => {
		server.use(
			http.get("*/roles", async () => {
				await delay(1_000);
				return HttpResponse.json(fixtures.roleList);
			}),
		);

		renderWithProviders(<DashboardPage />);

		expect(
			document.querySelectorAll(".animate-pulse").length,
		).toBeGreaterThanOrEqual(5);
	});

	it("shows the empty dashboard state when a scan has found no roles", async () => {
		server.use(
			http.get("*/roles", () =>
				HttpResponse.json({
					...fixtures.roleList,
					data: [],
					pagination: { total: 0, page: 1, per_page: 25, total_pages: 0 },
				}),
			),
		);

		renderWithProviders(<DashboardPage />);

		expect(await screen.findByText("No roles found")).toBeVisible();
		expect(screen.getByText("Run a scan to find jobs")).toBeVisible();
	});

	it("falls back to the empty dashboard state when role loading fails", async () => {
		server.use(
			http.get("*/roles", () =>
				HttpResponse.json({ message: "service unavailable" }, { status: 503 }),
			),
		);

		renderWithProviders(<DashboardPage />);

		expect(
			await screen.findByText("No roles found", {}, { timeout: 5_000 }),
		).toBeVisible();
	});
});
