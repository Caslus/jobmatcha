import { useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { expect, it } from "vitest";
import { renderWithProviders } from "./render";

function Counter() {
	const [count, setCount] = useState(0);
	const queryClient = useQueryClient();
	return (
		<button
			onClick={() => {
				queryClient.setQueryData(["count"], count + 1);
				setCount(count + 1);
			}}
			type="button"
		>
			Count: {count}
		</button>
	);
}

it("renders with an isolated query client and user interaction", async () => {
	const { getByRole, queryClient, user } = renderWithProviders(<Counter />);
	await user.click(getByRole("button", { name: "Count: 0" }));
	expect(getByRole("button")).toHaveTextContent("Count: 1");
	expect(queryClient.getQueryData(["count"])).toBe(1);
});
