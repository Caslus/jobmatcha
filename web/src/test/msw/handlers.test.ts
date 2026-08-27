import { expect, it, vi } from "vitest";
import { fixtures } from "./fixtures";
import { apiHandlers } from "./handlers";
import { useHandlers } from "./server";

it("serves typed fixture responses and accepts handler overrides", async () => {
	vi.stubEnv("VITE_API_BASE_URL", "http://localhost/api");
	useHandlers(
		apiHandlers.authStatus({ ...fixtures.authStatus, authenticated: false }),
	);
	const { authApi } = await import("../../lib/api");

	await expect(authApi.status()).resolves.toEqual({
		...fixtures.authStatus,
		authenticated: false,
	});
});
