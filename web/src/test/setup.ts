import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterAll, afterEach, beforeAll } from "vitest";
import { server } from "./msw/server";

// jsdom does not implement the native dialog methods used by the shared modal.
Object.defineProperties(HTMLDialogElement.prototype, {
	showModal: {
		configurable: true,
		value() {
			this.open = true;
		},
	},
	close: {
		configurable: true,
		value() {
			this.open = false;
		},
	},
});

beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterEach(() => {
	cleanup();
	server.resetHandlers();
});
afterAll(() => server.close());
