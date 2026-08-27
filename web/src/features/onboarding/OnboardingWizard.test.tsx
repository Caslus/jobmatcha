import { screen, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";
import { apiHandlers } from "../../test/msw/handlers";
import { server } from "../../test/msw/server";
import { renderWithRouter } from "../../test/render";
import { OnboardingWizard } from "./OnboardingWizard";

const noKeyAiSettings = {
	provider: "openrouter",
	enabled: false,
	has_api_key: false,
	user_name: "",
	user_email: "",
	user_location: "",
	user_linkedin: "",
	user_github: "",
};

async function skipAiAndEnterManualProfile() {
	await screen.findByRole("button", {
		name: /skip ai setup/i,
	});
	await screen.getByRole("button", { name: /skip ai setup/i }).click();
	await screen.findByRole("heading", { name: /review & edit/i });
	const user = (await import("@testing-library/user-event")).default.setup();
	await user.type(screen.getByLabelText("Name"), "Ada Lovelace");
	await user.type(screen.getByLabelText("Email"), "ada@example.com");
	await user.type(screen.getByLabelText("Location"), "London");
	await user.click(screen.getAllByRole("button", { name: "Add" })[0]);
	await user.type(screen.getByPlaceholderText("keywords, ..."), "typescript");
	await user.keyboard("{Enter}");
	await user.click(screen.getAllByRole("button", { name: "Add" })[1]);
	await user.type(screen.getByPlaceholderText("exclude, ..."), "sales");
	await user.keyboard("{Enter}");
	return user;
}

describe("OnboardingWizard", () => {
	it("validates the first-run password and advances through AI skip to manual profile review", async () => {
		server.use(
			apiHandlers.authStatus({
				authenticated: true,
				setup_complete: false,
				oidc_enabled: false,
			}),
			apiHandlers.aiSettings(noKeyAiSettings),
		);
		const { user } = renderWithRouter(<OnboardingWizard />);

		await screen.findByRole("heading", { name: /set your password/i });
		expect(screen.getByRole("button", { name: "Continue" })).toBeDisabled();
		await user.type(screen.getByLabelText("Bootstrap Password"), "bootstrap");
		await user.type(screen.getByLabelText("New Password"), "new-pass");
		await user.type(screen.getByLabelText("Confirm New Password"), "new-pass");
		await user.click(screen.getByRole("button", { name: "Continue" }));

		await screen.findByRole("heading", { name: /ai provider/i });
		await skipAiAndEnterManualProfile();
		expect(screen.getByText("typescript")).toBeInTheDocument();
		expect(screen.getByText("sales")).toBeInTheDocument();
	});

	it("uses the repeat-onboarding flow and submits chosen scan settings", async () => {
		let completion: unknown;
		server.use(
			apiHandlers.authStatus({
				authenticated: true,
				setup_complete: true,
				oidc_enabled: false,
			}),
			apiHandlers.aiSettings(noKeyAiSettings),
			http.post("*/api/onboarding/complete", async ({ request }) => {
				completion = await request.json();
				return HttpResponse.json({});
			}),
		);
		const { user } = renderWithRouter(<OnboardingWizard />);

		await screen.findByRole("heading", { name: /ai provider/i });
		expect(screen.queryByText(/set your password/i)).not.toBeInTheDocument();
		await skipAiAndEnterManualProfile();
		await user.click(screen.getByRole("button", { name: "Continue" }));
		await screen.findByRole("heading", { name: /scan schedule/i });
		await user.click(screen.getByRole("button", { name: "Launch" }));

		await waitFor(() =>
			expect(completion).toMatchObject({
				user_name: "Ada Lovelace",
				include_keywords: ["typescript"],
				exclude_keywords: ["sales"],
				scan_enabled: true,
			}),
		);
	});
});
