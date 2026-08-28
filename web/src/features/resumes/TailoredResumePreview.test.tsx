import { fireEvent, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/test/render";
import type { ResumeDocument } from "@/types/api.gen";
import { TailoredResumePreview } from "./TailoredResumePreview";

describe("TailoredResumePreview", () => {
	it("explains legacy tailored resumes and closes from the shared modal", async () => {
		const onClose = vi.fn();
		const legacyDocument: ResumeDocument = {
			header: { name: "", contact: [] },
			summary: "",
			sections: [],
			content: "legacy markdown",
		};
		const { user } = renderWithProviders(
			<TailoredResumePreview
				open
				document={legacyDocument}
				jobTitle="Frontend Engineer"
				profileLinks={["github.com/ada"]}
				onClose={onClose}
			/>,
		);

		expect(screen.getByText(/uses the legacy format/i)).toBeInTheDocument();
		expect(
			screen.queryByRole("button", { name: /save pdf/i }),
		).not.toBeInTheDocument();
		await user.click(
			screen.getByRole("button", { name: /close tailored resume preview/i }),
		);
		expect(onClose).toHaveBeenCalledOnce();
	});

	it("renders a structured document, saves appearance choices, and exports it", async () => {
		const onClose = vi.fn();
		const write = vi.fn();
		const close = vi.fn();
		const printWindow = {
			document: { write, close },
		} as unknown as Window;
		vi.spyOn(window, "open").mockReturnValue(printWindow);
		const document: ResumeDocument = {
			header: { name: "Ada Lovelace", contact: ["ada@example.com"] },
			summary: "Engineer who makes software understandable.",
			sections: [
				{
					heading: "Experience",
					kind: "experience",
					entries: [
						{
							title: "Staff Engineer",
							organization: "Analytical Engines",
							location: "London",
							date_range: "1843 – present",
							highlights: ["Built reliable systems."],
						},
					],
					items: [],
				},
				{
					heading: "Skills",
					kind: "list",
					entries: [],
					items: ["TypeScript", "React"],
				},
			],
			content: "",
		};
		const { user } = renderWithProviders(
			<TailoredResumePreview
				open
				document={document}
				jobTitle="Frontend Engineer"
				profileLinks={["github.com/ada", "ada@example.com"]}
				onClose={onClose}
			/>,
		);

		expect(screen.getByText("Ada Lovelace")).toBeInTheDocument();
		expect(screen.getByRole("link", { name: "GitHub" })).toHaveAttribute(
			"href",
			"https://github.com/ada",
		);
		expect(screen.getByText("Built reliable systems.")).toBeInTheDocument();

		await user.selectOptions(screen.getByLabelText(/typeface/i), "classic");
		await user.selectOptions(screen.getByLabelText(/page margins/i), "16");
		await user.click(screen.getByRole("button", { name: "Accent" }));
		await user.click(screen.getByRole("button", { name: "Off" }));
		fireEvent.change(screen.getByLabelText(/font size/i), {
			target: { value: "12" },
		});

		expect(localStorage.getItem("jobmatcha_resume_appearance")).toContain(
			'"font":"classic"',
		);
		await user.click(screen.getByRole("button", { name: /save pdf/i }));
		expect(write).toHaveBeenCalledWith(
			expect.stringContaining("Tailored resume — Frontend Engineer"),
		);
		expect(write).toHaveBeenCalledWith(
			expect.stringContaining("mailto:ada@example.com"),
		);
		expect(close).toHaveBeenCalledOnce();

		await user.click(
			screen.getByRole("button", { name: /reset resume appearance/i }),
		);
		expect(localStorage.getItem("jobmatcha_resume_appearance")).toContain(
			'"font":"modern"',
		);
	});
});
