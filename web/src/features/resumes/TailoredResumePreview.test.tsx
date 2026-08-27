import { screen } from "@testing-library/react";
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
});
