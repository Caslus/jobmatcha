import { screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/test/render";
import { CompaniesPage } from "./companies";

const { mutate } = vi.hoisted(() => ({ mutate: vi.fn() }));
vi.mock("../hooks/useApi", () => ({
	authStatusQueryOptions: vi.fn(),
	useCompanies: () => ({
		data: {
			data: [
				{
					id: 1,
					name: "Zulu",
					location: "Tokyo",
					active: true,
					board_count: 1,
					role_count: 5,
					freshness_status: "fresh",
					last_scan_attempt_at: "2026-09-01T12:00:00Z",
					last_new_role_discovery_at: "2026-09-01T12:00:00Z",
					career_boards: [
						{
							id: 10,
							provider: "fake",
							board_identifier: "zulu",
							canonical_url: "https://zulu.test",
							active: true,
							adapter_status: "healthy",
							freshness_status: "fresh",
							last_scan_attempt_at: "2026-09-01T12:00:00Z",
							last_successful_scan_at: "2026-09-01T12:00:00Z",
							last_scan_failure_detail: null,
							last_new_role_discovery_at: "2026-09-01T12:00:00Z",
						},
					],
				},
				{
					id: 2,
					name: "Alpha",
					location: "Remote",
					active: false,
					board_count: 0,
					role_count: 0,
					freshness_status: "not_applicable",
					last_scan_attempt_at: null,
					last_new_role_discovery_at: null,
					career_boards: [],
				},
			],
		},
		isLoading: false,
		error: null,
	}),
	useUpdateCompanyActive: () => ({ mutate, isPending: false }),
	useUpdateCompaniesActiveBulk: () => ({ mutate, isPending: false }),
	useUpdateCareerBoardActive: () => ({ mutate, isPending: false }),
	useDiscoverCareerBoards: () => ({
		data: null,
		isPending: false,
		error: null,
		mutate,
		reset: vi.fn(),
	}),
	useRegisterCareerBoards: () => ({ isPending: false, mutate }),
	useLogout: () => ({ mutate, isPending: false }),
}));

describe("CompaniesPage", () => {
	it("shows a compact company index and board-level operational details", async () => {
		const { user } = renderWithProviders(<CompaniesPage />);
		expect(
			screen.getByLabelText("Freshness is not tracked for this company"),
		).toBeVisible();
		await user.click(screen.getByLabelText("Select Alpha"));
		expect(screen.getByRole("button", { name: /enable 1/i })).toBeEnabled();
		await user.click(screen.getByRole("switch", { name: "Zulu enabled" }));
		expect(mutate).toHaveBeenCalledWith({ id: 1, active: false });
		await user.click(
			screen.getByRole("button", { name: "Manage Zulu career boards" }),
		);
		expect(screen.getByText("https://zulu.test")).toBeVisible();
		expect(screen.getByText("Latest discovery")).toBeVisible();
	});
});
