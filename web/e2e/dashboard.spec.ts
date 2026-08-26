import { expect, test } from "@playwright/test";
import { readFile } from "node:fs/promises";

test("signs in and persists a role preference", async ({ page }) => {
	const password = (await readFile(".e2e-tmp/bootstrap-password", "utf8")).trim();

	await page.goto("/");
	await page.getByPlaceholder("Password").fill(password);
	await page.getByRole("button", { name: "Sign in" }).click();

	await expect(page).toHaveURL(/\/dashboard$/);
	await expect(page.getByText("Software Engineer").first()).toBeVisible();
	await page.getByRole("button", { name: "Bookmark this role" }).click();

	await expect
		.poll(async () => {
			const response = await page.request.get("/api/roles");
			const roles = (await response.json()) as { data: Array<{ is_interested: boolean }> };
			return roles.data[0]?.is_interested;
		})
		.toBe(true);
});
