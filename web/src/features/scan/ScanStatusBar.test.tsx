import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";
import { fixtures } from "@/test/msw/fixtures";
import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { ScanStatusBar } from "./ScanStatusBar";

describe("ScanStatusBar", () => {
	it("shows running progress and starts a manual scan from its settings modal", async () => {
		server.use(
			http.get("*/api/scan/:id", () =>
				HttpResponse.json({
					...fixtures.scanJob,
					status: "running",
					total_companies: 4,
					completed_companies: 2,
				}),
			),
			http.get("*/api/scan/latest", () =>
				HttpResponse.json({
					...fixtures.scanJob,
					status: "running",
					total_companies: 4,
					completed_companies: 2,
				}),
			),
		);
		const { user } = renderWithProviders(<ScanStatusBar />);

		await screen.findByText("Scanning...");
		expect(screen.getByText("2/4 companies")).toBeInTheDocument();
		await user.click(screen.getByTitle("Scan settings"));
		expect(screen.getByRole("button", { name: "Scanning..." })).toBeDisabled();
	});

	it("renders failure details instead of completed scan statistics", async () => {
		server.use(
			http.get("*/api/scan/:id", () =>
				HttpResponse.json({
					...fixtures.scanJob,
					status: "failed",
					error: "ATS timeout",
				}),
			),
			http.get("*/api/scan/latest", () =>
				HttpResponse.json({
					...fixtures.scanJob,
					status: "failed",
					error: "ATS timeout",
				}),
			),
		);
		renderWithProviders(<ScanStatusBar />);
		await screen.findByText("Scan failed");
		expect(screen.getByText("ATS timeout")).toBeInTheDocument();
	});
});
