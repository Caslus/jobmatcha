import { screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/test/render";
import { Button } from "./button";
import { Input } from "./input";
import { Label } from "./label";
import { Modal } from "./modal";
import {
	Select,
	SelectContent,
	SelectGroup,
	SelectItem,
	SelectLabel,
	SelectSeparator,
	SelectTrigger,
	SelectValue,
} from "./select";
import { Slider } from "./slider";
import { Switch } from "./switch";
import { ToggleSwitch } from "./ToggleSwitch";
import { Textarea } from "./textarea";

describe("shared UI primitives", () => {
	it("renders button variants, form controls, and their semantic slots", () => {
		renderWithProviders(
			<>
				<Label htmlFor="name">Name</Label>
				<Input id="name" defaultValue="Ada" />
				<Textarea aria-label="Notes" defaultValue="A note" />
				<Button variant="destructive" size="sm">
					Delete
				</Button>
				<Slider aria-label="Score" defaultValue={[25]} />
			</>,
		);

		expect(screen.getByLabelText("Name")).toHaveAttribute("data-slot", "input");
		expect(screen.getByLabelText("Notes")).toHaveAttribute(
			"data-slot",
			"textarea",
		);
		expect(screen.getByRole("button", { name: "Delete" })).toHaveAttribute(
			"data-variant",
			"destructive",
		);
		expect(screen.getByLabelText("Score")).toHaveAttribute(
			"data-slot",
			"slider",
		);
	});

	it("supports the Radix switch and select interactions", async () => {
		const { user } = renderWithProviders(
			<>
				<Switch aria-label="Notifications" size="sm" />
				<Select>
					<SelectTrigger aria-label="Seniority">
						<SelectValue placeholder="Choose" />
					</SelectTrigger>
					<SelectContent>
						<SelectGroup>
							<SelectLabel>Level</SelectLabel>
							<SelectSeparator />
							<SelectItem value="senior">Senior</SelectItem>
						</SelectGroup>
					</SelectContent>
				</Select>
			</>,
		);

		const toggle = screen.getByRole("switch", { name: "Notifications" });
		expect(toggle).toHaveAttribute("data-size", "sm");
		await user.click(toggle);
		expect(toggle).toHaveAttribute("data-state", "checked");

		await user.click(screen.getByRole("combobox", { name: "Seniority" }));
		await user.click(await screen.findByRole("option", { name: "Senior" }));
		expect(
			screen.getByRole("combobox", { name: "Seniority" }),
		).toHaveTextContent("Senior");
	});

	it("delegates the custom toggle and modal dismissal paths", async () => {
		const onToggle = vi.fn();
		const onClose = vi.fn();
		const { user } = renderWithProviders(
			<>
				<ToggleSwitch checked={false} onChange={onToggle} />
				<Modal
					open
					onClose={onClose}
					title="Preferences"
					subtitle="Optional settings"
				>
					<p>Content</p>
				</Modal>
			</>,
		);

		await user.click(screen.getByRole("switch"));
		expect(onToggle).toHaveBeenCalledOnce();
		await user.click(screen.getByRole("button", { name: "Close dialog" }));
		expect(onClose).toHaveBeenCalledOnce();
	});
});
