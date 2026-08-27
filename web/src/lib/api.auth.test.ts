import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";
import { server } from "../test/msw/server";
import { authApi } from "./api";

describe("authentication API client", () => {
	it("sends login, logout, and password-change requests to the API", async () => {
		let loginBody: unknown;
		let passwordBody: unknown;
		let didLogout = false;
		server.use(
			http.post("*/api/auth/login", async ({ request }) => {
				loginBody = await request.json();
				return HttpResponse.json({ authenticated: true });
			}),
			http.post("*/api/auth/change-password", async ({ request }) => {
				passwordBody = await request.json();
				return HttpResponse.json({});
			}),
			http.post("*/api/auth/logout", () => {
				didLogout = true;
				return HttpResponse.json({});
			}),
		);

		await expect(authApi.login("secret")).resolves.toEqual({
			authenticated: true,
		});
		await expect(authApi.changePassword("old", "new-secret")).resolves.toEqual(
			{},
		);
		await expect(authApi.logout()).resolves.toEqual({});

		expect(loginBody).toEqual({ password: "secret" });
		expect(passwordBody).toEqual({
			current_password: "old",
			new_password: "new-secret",
		});
		expect(didLogout).toBe(true);
	});

	it("exposes API error messages from failed login requests", async () => {
		server.use(
			http.post("*/api/auth/login", () =>
				HttpResponse.json({ error: "Invalid password" }, { status: 401 }),
			),
		);

		await expect(authApi.login("wrong")).rejects.toThrow("Invalid password");
	});
});
