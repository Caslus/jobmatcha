import { fireEvent, screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it, vi } from "vitest";
import { server } from "@/test/msw/server";
import { renderWithProviders } from "@/test/render";
import { ResumeUploadStep } from "./ResumeUploadStep";

describe("ResumeUploadStep", () => {
	it("rejects unsupported files, parses supported resumes, and keeps manual entry available", async () => {
		const onData = vi.fn();
		const onSkip = vi.fn();
		server.use(
			http.post("*/api/ai/parse-resume", () =>
				HttpResponse.json({ user_name: "Ada" }),
			),
		);
		const { user } = renderWithProviders(
			<ResumeUploadStep onData={onData} onSkip={onSkip} />,
		);
		const input = document.querySelector(
			'input[type="file"]',
		) as HTMLInputElement;

		fireEvent.change(input, {
			target: { files: [new File(["x"], "resume.docx")] },
		});
		expect(onData).not.toHaveBeenCalled();
		fireEvent.change(input, {
			target: { files: [new File(["# Ada"], "resume.md")] },
		});
		await screen.findByText("resume.md");
		expect(onData).toHaveBeenCalledWith({ user_name: "Ada" });
		await user.click(
			screen.getByRole("button", { name: /enter details manually/i }),
		);
		expect(onSkip).toHaveBeenCalledOnce();
	});
});
