import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "../test/render";
import { Tooltip } from "./Tooltip";

describe("Tooltip", () => {
	it("shows and hides supplementary content on hover", async () => {
		const { user } = renderWithProviders(
			<Tooltip content="Posted on January 1">
				<span>1 day ago</span>
			</Tooltip>,
		);

		await user.hover(screen.getByText("1 day ago"));
		expect(await screen.findByText("Posted on January 1")).toBeVisible();
		await user.unhover(screen.getByText("1 day ago"));
		expect(screen.queryByText("Posted on January 1")).not.toBeInTheDocument();
	});
});
