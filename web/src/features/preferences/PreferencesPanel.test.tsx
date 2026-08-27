import { fireEvent, screen, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";
import { fixtures } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { PreferencesPanel } from "./PreferencesPanel";

describe("PreferencesPanel", () => {
	it("stages keyword and freshness edits, then saves the generated settings payload", async () => {
		let payload: unknown;
		server.use(
			http.get("*/api/scan/:id", () => HttpResponse.json(fixtures.scanJob)),
			http.put("*/api/settings", async ({ request }) => {
				payload = await request.json();
				return HttpResponse.json({
					...fixtures.settings,
					...(payload as object),
				});
			}),
		);
		const { user } = renderWithProviders(<PreferencesPanel />);

		await screen.findByText("TypeScript");
		await user.click(screen.getAllByRole("button", { name: "Add" })[0]);
		const keywordInput = screen.getByPlaceholderText(", ...");
		await user.type(keywordInput, "React, TypeScript");
		await user.keyboard("{Enter}");
		await user.click(screen.getByRole("button", { name: "Full-time" }));
		fireEvent.change(screen.getByRole("slider"), { target: { value: "1" } });
		await user.click(screen.getByRole("button", { name: "Apply" }));

		await waitFor(() =>
			expect(payload).toMatchObject({
				include_keywords: ["TypeScript", "react", "typescript"],
				work_types: ["remote", "full-time"],
				max_days_old: 2,
			}),
		);
	});
});
