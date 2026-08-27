import { describe, expect, it } from "vitest";
import { formatToHtml, markdownToHtml, sanitiseHtml } from "./description";

describe("description formatting", () => {
	it("removes executable markup and unsafe URLs", () => {
		const html = sanitiseHtml(
			'<script>alert(1)</script><a href="javascript:alert(1)" onclick="x()">safe</a>',
		);
		expect(html).not.toContain("script");
		expect(html).not.toContain("onclick");
		expect(html).not.toContain("javascript:");
	});

	it("renders markdown and plain text using safe HTML", () => {
		expect(markdownToHtml("# Role\n\n**TypeScript**")).toContain("<h1>");
		expect(formatToHtml("plain", "first\nsecond")).toContain("<br />");
		expect(formatToHtml("markdown", "[link](https://example.com)")).toContain(
			"https://example.com",
		);
	});
});
