import { screen } from "@testing-library/react";
import { expect, it } from "vitest";
import { renderWithProviders } from "../../test/render";
import { JobDescription } from "./JobDescription";

it("renders formatted content while excluding dangerous markup", () => {
	renderWithProviders(
		<JobDescription
			description={
				"# Great role\n\n<script>alert(1)</script>Build **products**."
			}
			format="markdown"
		/>,
	);
	expect(screen.getByText("Great role")).toBeInTheDocument();
	expect(screen.getByText("products")).toBeInTheDocument();
	expect(screen.queryByText("alert(1)")).not.toBeInTheDocument();
});

it("shows an empty-state message for missing descriptions", () => {
	renderWithProviders(<JobDescription description="" format="plain" />);
	expect(screen.getByText("No description available")).toBeInTheDocument();
});
