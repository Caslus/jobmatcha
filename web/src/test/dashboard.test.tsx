import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { DashboardPage } from "../routes/dashboard";
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
});
