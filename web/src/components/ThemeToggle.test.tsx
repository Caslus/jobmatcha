import { screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/test/render";
import ThemeToggle from "./ThemeToggle";

describe("ThemeToggle", () => {
	beforeEach(() => {
		document.documentElement.className = "";
		vi.stubGlobal("matchMedia", () => ({
			matches: true,
			addEventListener: vi.fn(),
			removeEventListener: vi.fn(),
		}));
	});

	it("cycles auto, light, dark and persists the explicit choice", async () => {
		const { user } = renderWithProviders(<ThemeToggle />);
		const toggle = screen.getByRole("button", { name: /theme mode: auto/i });
		expect(document.documentElement).toHaveClass("dark");

		await user.click(toggle);
		expect(
			screen.getByRole("button", { name: /theme mode: light/i }),
		).toBeInTheDocument();
		expect(document.documentElement).toHaveAttribute("data-theme", "light");

		await user.click(
			screen.getByRole("button", { name: /theme mode: light/i }),
		);
		expect(document.documentElement).toHaveAttribute("data-theme", "dark");
	});
});
