import type { RequestHandler } from "msw";
import { setupServer } from "msw/node";
import { handlers } from "./handlers";

export const server = setupServer(...handlers);

export function useHandlers(...overrides: RequestHandler[]) {
	server.use(...overrides);
}
