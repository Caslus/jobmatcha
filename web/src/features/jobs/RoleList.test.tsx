import { screen, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";
import { fixtures } from "../../test/msw/fixtures";
import { server } from "../../test/msw/server";
import { renderWithProviders } from "../../test/render";
import { RoleList } from "./RoleList";

describe("RoleList", () => {
	it("shows an actionable empty state when a scan has no matching roles", async () => {
		server.use(
			http.get("*/roles", () =>
				HttpResponse.json({
					...fixtures.roleList,
					data: [],
					pagination: { total: 0, page: 1, per_page: 25, total_pages: 0 },
				}),
			),
		);

		renderWithProviders(
			<RoleList selectedId={null} onSelect={() => undefined} />,
		);
		expect(await screen.findByText("No roles found")).toBeVisible();
		expect(screen.getByText("Run a scan to find jobs")).toBeVisible();
	});

	it("sends a bookmark update for the selected role", async () => {
		let requestBody: unknown;
		server.use(
			http.patch("*/roles/1", async ({ request }) => {
				requestBody = await request.json();
				return HttpResponse.json({});
			}),
		);
		const { user } = renderWithProviders(
			<RoleList selectedId={null} onSelect={() => undefined} />,
		);

		await user.click(
			await screen.findByRole("button", { name: "Bookmark this role" }),
		);
		await waitFor(() => expect(requestBody).toEqual({ is_interested: true }));
	});
});
