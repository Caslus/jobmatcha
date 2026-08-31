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
					ats_type: "fake",
					active: true,
					role_count: 5,
					adapter_status: "healthy",
					freshness_status: "fresh",
				},
				{
					id: 2,
					name: "Alpha",
					location: "Remote",
					ats_type: "other",
					active: false,
					role_count: 0,
					adapter_status: "unsupported",
					freshness_status: "not_applicable",
				},
				{
					id: 3,
					name: "Broken",
					location: "Tokyo",
					ats_type: "fake",
					active: true,
					role_count: 2,
					adapter_status: "failing",
					freshness_status: "stale",
					last_scan_failure_detail: "Timeout",
				},
				{
					id: 4,
					name: "New",
					location: "Tokyo",
					ats_type: "fake",
					active: true,
					role_count: 1,
					adapter_status: "unknown",
					freshness_status: "no_activity_yet",
				},
			],
		},
		isLoading: false,
		error: null,
	}),
	useUpdateCompanyActive: () => ({ mutate, isPending: false }),
	useUpdateCompaniesActiveBulk: () => ({ mutate, isPending: false }),
	useLogout: () => ({ mutate, isPending: false }),
}));

describe("CompaniesPage", () => {
	it("sorts by role count by default and exposes status explanations", async () => {
		const { user } = renderWithProviders(<CompaniesPage />);
		expect(
			screen.getByLabelText("No new roles found in 30 days"),
		).toBeVisible();
		expect(screen.getByLabelText("Latest scan failed: Timeout")).toBeVisible();
		expect(
			screen.getByLabelText("Freshness is not tracked for this company"),
		).toBeVisible();
		expect(screen.getAllByRole("row")[1]).toHaveTextContent("Zulu");
		await user.click(
			screen.getByRole("button", { name: /^Company$/ }),
		);
		const names = screen
			.getAllByRole("row")
			.slice(1)
			.map((row) => row.textContent);
		expect(names[0]).toContain("Alpha");
		await user.click(screen.getByLabelText("Select Alpha"));
		expect(screen.getByRole("button", { name: /enable 1/i })).toBeEnabled();
		await user.click(screen.getByRole("switch", { name: "Zulu enabled" }));
		expect(mutate).toHaveBeenCalledWith({ id: 1, active: false });
	});
});
