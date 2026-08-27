import { screen, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";
import { server } from "../../test/msw/server";
import { renderWithProviders } from "../../test/render";
import { RoleDetailPanel } from "./RoleDetailPanel";

describe("RoleDetailPanel", () => {
	it("explains why tailoring is unavailable when the AI provider is disabled", async () => {
		server.use(
			http.get("*/settings/ai", () =>
				HttpResponse.json({ provider: "", enabled: false, has_api_key: false }),
			),
		);
		renderWithProviders(
			<RoleDetailPanel selectedId={1} onBack={() => undefined} />,
		);

		const tailor = await screen.findByRole("button", {
			name: "AI provider disabled",
		});
		expect(tailor).toBeDisabled();
		expect(tailor).toHaveAttribute(
			"title",
			"Enable an AI provider in Settings to tailor resumes",
		);
	});

	it("displays the server error when tailoring a resume fails", async () => {
		server.use(
			http.post("*/roles/1/tailor", () =>
				HttpResponse.json(
					{ error: "Resume service is unavailable" },
					{ status: 503 },
				),
			),
		);
		const { user } = renderWithProviders(
			<RoleDetailPanel selectedId={1} onBack={() => undefined} />,
		);

		await user.click(
			await screen.findByRole("button", { name: "Tailor resume" }),
		);
		await waitFor(() =>
			expect(screen.getByText("Resume service is unavailable")).toBeVisible(),
		);
	});
});
