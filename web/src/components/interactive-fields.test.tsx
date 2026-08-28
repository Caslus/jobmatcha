import { screen } from "@testing-library/react";
import { createRef, useState } from "react";
import { describe, expect, it, vi } from "vitest";
import {
	KeywordSection,
	WorkTypeToggle,
} from "@/components/keywords/KeywordSection";
import { ScanScheduleFields } from "@/components/scan-settings/ScanScheduleFields";
import { renderWithProviders } from "@/test/render";

describe("interactive settings fields", () => {
	it("edits scan schedule fields and exposes the disabled guidance", async () => {
		const onChange = vi.fn();
		function ScheduleHarness() {
			const [value, setValue] = useState({
				scan_enabled: false,
				scan_cron_expr: "0 6 * * *",
				scan_timezone: "UTC",
			});
			return (
				<ScanScheduleFields
					value={value}
					onChange={(nextValue) => {
						onChange(nextValue);
						setValue(nextValue);
					}}
				/>
			);
		}
		const { user } = renderWithProviders(<ScheduleHarness />);

		expect(screen.getByText(/scanner will run automatically/i)).toBeVisible();
		await user.click(screen.getByRole("switch"));
		expect(onChange).toHaveBeenCalledWith(
			expect.objectContaining({ scan_enabled: true }),
		);
		await user.clear(screen.getByLabelText("Cron expression"));
		await user.type(screen.getByLabelText("Cron expression"), "0 9 * * 1");
		expect(onChange).toHaveBeenLastCalledWith(
			expect.objectContaining({ scan_cron_expr: "0 9 * * 1" }),
		);
	});

	it("stages keyword additions and marks selected chips", async () => {
		const onStageAdd = vi.fn();
		const onChipClick = vi.fn();
		const onAddInput = vi.fn();
		const { user, rerender } = renderWithProviders(
			<KeywordSection
				label="Skills"
				field="include"
				keywords={["React"]}
				showInput={false}
				onAddInput={onAddInput}
				onCloseInput={vi.fn()}
				inputRef={createRef()}
				highlight="green"
				onStageAdd={onStageAdd}
				onChipClick={onChipClick}
				markedForDelete={new Set()}
			/>,
		);

		await user.click(screen.getByRole("button", { name: "React" }));
		expect(onChipClick).toHaveBeenCalledWith("React", "include");
		await user.click(screen.getByRole("button", { name: "Add" }));
		expect(onAddInput).toHaveBeenCalledOnce();

		rerender(
			<KeywordSection
				label="Skills"
				field="include"
				keywords={[]}
				showInput
				onAddInput={onAddInput}
				onCloseInput={vi.fn()}
				inputRef={createRef()}
				highlight="green"
				onStageAdd={onStageAdd}
				onChipClick={onChipClick}
				markedForDelete={new Set()}
			/>,
		);
		await user.type(
			screen.getByPlaceholderText("skills, ..."),
			"TypeScript{enter}",
		);
		expect(onStageAdd).toHaveBeenCalledWith("include", "TypeScript");
	});

	it("renders work-type selection state and invokes its callback", async () => {
		const onToggle = vi.fn();
		const { user } = renderWithProviders(
			<WorkTypeToggle label="Remote" selected onToggle={onToggle} />,
		);
		await user.click(screen.getByRole("button", { name: "Remote" }));
		expect(onToggle).toHaveBeenCalledOnce();
		expect(screen.getByRole("button", { name: "Remote" })).toHaveClass(
			"text-[#7dba7a]",
		);
	});
});
