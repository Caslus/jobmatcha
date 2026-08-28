import { screen, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it, vi } from "vitest";
import { fixtures } from "@/test/msw/fixtures";
import { apiHandlers } from "@/test/msw/handlers";
import { server } from "@/test/msw/server";
import { renderWithRouter } from "@/test/render";
import { SettingsPanel } from "./SettingsPanel";

describe("SettingsPanel", () => {
	it("loads AI configuration and switches to profile editing", async () => {
		const onClose = vi.fn();
		const { user } = renderWithRouter(<SettingsPanel onClose={onClose} />);

		await screen.findByText(/a key is already saved/i);
		expect(
			screen.getByRole("option", { name: "OpenRouter" }),
		).toBeInTheDocument();
		await user.click(screen.getByRole("button", { name: "Profile" }));
		expect(screen.getByDisplayValue("test@example.com")).toBeInTheDocument();
		await user.click(screen.getByRole("button", { name: "Close dialog" }));
		expect(onClose).toHaveBeenCalledOnce();
	});

	it("validates a replacement key and saves the enabled AI provider settings", async () => {
		let validation: unknown;
		let update: unknown;
		server.use(
			apiHandlers.aiSettings({
				...fixtures.aiSettings,
				enabled: false,
				has_api_key: false,
			}),
			http.post("*/api/ai/validate-key", async ({ request }) => {
				validation = await request.json();
				return HttpResponse.json({ valid: true });
			}),
			http.put("*/api/settings/ai", async ({ request }) => {
				update = await request.json();
				return HttpResponse.json({});
			}),
		);
		const { user } = renderWithRouter(<SettingsPanel onClose={vi.fn()} />);

		const enabled = await screen.findByRole("switch", { name: "Enabled" });
		await user.click(enabled);
		await user.type(screen.getByLabelText("API Key"), "replacement-key");
		await user.click(screen.getByRole("button", { name: "Validate" }));

		await screen.findByText("Key is valid");
		expect(validation).toEqual({
			provider: "openai",
			api_key: "replacement-key",
		});
		await user.click(screen.getByRole("button", { name: "Save" }));
		await waitFor(() =>
			expect(update).toEqual({
				enabled: true,
				provider: "openai",
				api_key: "replacement-key",
			}),
		);
		await screen.findByRole("button", { name: "Saved ✓" });
	});

	it("saves profile edits and displays the summarized keyword preferences", async () => {
		let update: unknown;
		server.use(
			http.put("*/api/settings/ai", async ({ request }) => {
				update = await request.json();
				return HttpResponse.json({});
			}),
			apiHandlers.settings({
				...fixtures.settings,
				include_keywords: ["TypeScript", "React"],
				exclude_keywords: ["agency"],
				max_days_old: 0,
			}),
		);
		const { user } = renderWithRouter(<SettingsPanel onClose={vi.fn()} />);

		await screen.findByRole("switch", { name: "Enabled" });
		await user.click(screen.getByRole("button", { name: "Profile" }));
		const name = await screen.findByLabelText("Name");
		await user.clear(name);
		await user.type(name, "Ada Lovelace");
		await user.clear(screen.getByLabelText("GitHub URL"));
		await user.type(
			screen.getByLabelText("GitHub URL"),
			"https://github.com/ada",
		);
		await user.click(screen.getByRole("button", { name: "Save" }));
		await waitFor(() =>
			expect(update).toMatchObject({
				user_name: "Ada Lovelace",
				user_email: "test@example.com",
				user_github: "https://github.com/ada",
			}),
		);

		await user.click(screen.getByRole("button", { name: "Keywords" }));
		await screen.findByText("TypeScript, React");
		expect(screen.getByText("agency")).toBeInTheDocument();
		expect(screen.getByText("Any date")).toBeInTheDocument();
	});

	it("validates password input before submitting and surfaces a server error", async () => {
		server.use(
			http.post("*/api/auth/change-password", () =>
				HttpResponse.json(
					{ error: "Current password is incorrect" },
					{ status: 401 },
				),
			),
		);
		const { user } = renderWithRouter(<SettingsPanel onClose={vi.fn()} />);

		await screen.findByRole("switch", { name: "Enabled" });
		await user.click(screen.getByRole("button", { name: "Password" }));
		const current = screen.getByLabelText("Current Password");
		const next = screen.getByLabelText("New Password");
		const confirm = screen.getByLabelText("Confirm New Password");
		await user.type(current, "old-password");
		await user.type(next, "short");
		await user.type(confirm, "short");
		await user.click(screen.getByRole("button", { name: "Change Password" }));
		expect(
			screen.getByText("Password must be at least 6 characters"),
		).toBeInTheDocument();

		await user.clear(next);
		await user.clear(confirm);
		await user.type(next, "new-password");
		await user.type(confirm, "different-password");
		await user.click(screen.getByRole("button", { name: "Change Password" }));
		expect(screen.getByText("Passwords do not match")).toBeInTheDocument();

		await user.clear(confirm);
		await user.type(confirm, "new-password");
		await user.click(screen.getByRole("button", { name: "Change Password" }));
		await screen.findByText("Current password is incorrect");
	});
});
