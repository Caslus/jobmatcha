import { describe, expect, it } from "vitest";
import { authStatusQueryOptions } from "./useApi";

describe("auth status query", () => {
	it("uses the shared auth status cache key", () => {
		expect(authStatusQueryOptions().queryKey).toEqual(["auth", "status"]);
	});
});
