import { describe, expect, it } from "vitest";
import { queryKeys } from "../lib/queryKeys";
import { createTestQueryClient } from "../test/render";
import { Route as DashboardRoute } from "./dashboard";
import { Route as LoginRoute } from "./index";
import { Route as OnboardingRoute } from "./onboarding";

const authenticated = { authenticated: true, setup_complete: true };
const incompleteSetup = { authenticated: true, setup_complete: false };
const signedOut = { authenticated: false, setup_complete: false };

function guardContext(status: typeof authenticated) {
	const queryClient = createTestQueryClient();
	queryClient.setQueryData(queryKeys.auth.status, status);
	return { context: { queryClient } } as never;
}

describe("authentication route guards", () => {
	it("keeps signed-out users out of the dashboard and onboarding", async () => {
		await expect(
			DashboardRoute.options.beforeLoad?.(guardContext(signedOut)),
		).rejects.toBeDefined();
		await expect(
			OnboardingRoute.options.beforeLoad?.(guardContext(signedOut)),
		).rejects.toBeDefined();
	});

	it("sends authenticated users to onboarding until setup is complete", async () => {
		await expect(
			DashboardRoute.options.beforeLoad?.(guardContext(incompleteSetup)),
		).rejects.toBeDefined();
	});

	it("allows a completed setup into the dashboard", async () => {
		await expect(
			DashboardRoute.options.beforeLoad?.(guardContext(authenticated)),
		).resolves.toBeUndefined();
	});

	it("keeps authenticated users off the login route", async () => {
		await expect(
			LoginRoute.options.beforeLoad?.(guardContext(authenticated)),
		).rejects.toBeDefined();
	});
});
