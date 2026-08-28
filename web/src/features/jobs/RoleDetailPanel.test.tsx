import { screen, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it, vi } from "vitest";
import { fixtures } from "../../test/msw/fixtures";
import { server } from "../../test/msw/server";
import { renderWithProviders } from "../../test/render";
import { RoleDetailPanel } from "./RoleDetailPanel";

describe("RoleDetailPanel", () => {
	it("shows an empty-state prompt when no role is selected", () => {
		renderWithProviders(
			<RoleDetailPanel selectedId={null} onBack={() => undefined} />,
		);

		expect(screen.getByText("Select a role to view details")).toBeVisible();
	});

	it("closes the panel and persists a bookmark toggle", async () => {
		const onBack = vi.fn();
		let requestBody: unknown;
		server.use(
			http.patch("*/roles/1", async ({ request }) => {
				requestBody = await request.json();
				return HttpResponse.json({});
			}),
		);
		const { user } = renderWithProviders(
			<RoleDetailPanel selectedId={1} onBack={onBack} />,
		);

		await user.click(await screen.findByRole("button", { name: "Bookmark" }));
		await waitFor(() => expect(requestBody).toEqual({ is_interested: true }));

		await user.click(
			screen.getByRole("button", { name: "Close role details" }),
		);
		expect(onBack).toHaveBeenCalledOnce();
	});

	it("renders job metadata, links, and a detailed match analysis", async () => {
		server.use(
			http.get("*/roles/1", () =>
				HttpResponse.json({
					...fixtures.roleDetail,
					is_interested: true,
					posted_at: "2026-08-20T00:00:00Z",
					match_percent: 76,
					match_reasons: [
						"title:Engineer",
						"location:Remote",
						"work_type:Remote",
						"other:Fallback",
					],
					match_details: {
						matched_keywords: 3,
						total_keywords: 4,
						include_score: 5,
						bonus_score: 1,
						total_score: 6,
						recency_factor: 0.8,
						adjusted_score: 4.8,
						percent: 76,
					},
				}),
			),
		);
		renderWithProviders(
			<RoleDetailPanel selectedId={1} onBack={() => undefined} />,
		);

		expect(await screen.findByText("Match Analysis")).toBeVisible();
		expect(screen.getByRole("button", { name: "Bookmarked" })).toBeVisible();
		expect(screen.getByRole("link", { name: "Visit Website" })).toHaveAttribute(
			"href",
			"https://example.com/jobs/1",
		);
		expect(screen.getAllByText("Engineer")).toHaveLength(1);
		expect(screen.getAllByText("(Title)")).toHaveLength(1);
		expect(screen.getAllByText("(Region)")).toHaveLength(1);
		expect(screen.getAllByText("(Type)")).toHaveLength(1);
		expect(screen.getAllByText("Fallback")).toHaveLength(1);
		expect(screen.getByText("Bonus (work type)")).toBeVisible();
		expect(screen.getByText("×0.80")).toBeVisible();
		expect(screen.getByText("Adjusted score")).toBeVisible();
	});

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
