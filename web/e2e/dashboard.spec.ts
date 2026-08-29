import { expect, test } from "@playwright/test";
import { readFile } from "node:fs/promises";

test("signs in and persists a role preference", async ({ page }) => {
	const password = (await readFile(".e2e-tmp/bootstrap-password", "utf8")).trim();

	await page.goto("/");
	await page.getByPlaceholder("Password").fill(password);
	await page.getByRole("button", { name: "Sign in" }).click();

	await expect(page).toHaveURL(/\/dashboard$/);
	await expect(page.getByText("Software Engineer, Ads Backend").first()).toBeVisible();
	await page.getByRole("button", { name: /Software Engineer, Ads Backend/ }).first().click();
	await page.getByRole("button", { name: "Bookmark", exact: true }).click();

	await expect
		.poll(async () => {
			const response = await page.request.get("/api/roles");
			const roles = (await response.json()) as {
				data: Array<{ is_interested: boolean; title: string }>;
			};
			return roles.data.find((role) => role.title === "Software Engineer, Ads Backend")
				?.is_interested;
		})
		.toBe(true);
});
