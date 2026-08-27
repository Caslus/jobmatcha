import { screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
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
});
