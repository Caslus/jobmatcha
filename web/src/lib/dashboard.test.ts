import { describe, expect, it, vi } from "vitest";
import { matchLabel, scoreColor, timeAgo } from "./dashboard";

describe("dashboard helpers", () => {
	it("formats relative times across all display ranges", () => {
		vi.useFakeTimers();
		vi.setSystemTime(new Date("2026-01-31T12:00:00Z"));
		expect(timeAgo(null)).toBe("");
		expect(timeAgo("2026-01-31T11:59:40Z")).toBe("just now");
		expect(timeAgo("2026-01-31T11:45:00Z")).toBe("15m ago");
		expect(timeAgo("2026-01-31T09:00:00Z")).toBe("3h ago");
		expect(timeAgo("2026-01-29T12:00:00Z")).toBe("2d ago");
		expect(timeAgo("2025-12-01T12:00:00Z")).toBe("8w ago");
		vi.useRealTimers();
	});

	it("maps match scores to labels and visual bands", () => {
		expect(matchLabel(0)).toBe("–");
		expect(matchLabel(42)).toBe("42%");
		expect(scoreColor(100)).toContain("purple");
		expect(scoreColor(70)).toContain("green");
		expect(scoreColor(40)).toContain("amber");
		expect(scoreColor(39)).toContain("gray");
	});
});
