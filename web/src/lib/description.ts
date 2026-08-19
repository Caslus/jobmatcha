/**
 * Description rendering utilities.
 *
 * The server now stores descriptions in their original format and provides
 * a `description_format` field: "markdown", "html", or "plain".
 * This module provides a converter for each format.
 */

/* ───────── HTML entity unescaping ───────── */

/** Convert HTML entities back to their characters. */
function unescapeHtml(text: string): string {
	return text
		.replace(/&lt;/g, "<")
		.replace(/&gt;/g, ">")
		.replace(/&amp;/g, "&")
		.replace(/&quot;/g, '"')
		.replace(/&#39;/g, "'")
		.replace(/&nbsp;/g, " ")
		.replace(/&#(\d+);/g, (_m, code) => String.fromCharCode(Number(code)));
}

/* ───────── HTML sanitisation ───────── */

/** Tags we allow through the sanitizer. Everything else is stripped. */
const ALLOWED_TAGS = new Set([
	"p",
	"div",
	"span",
	"h1",
	"h2",
	"h3",
	"h4",
	"h5",
	"h6",
	"ul",
	"ol",
	"li",
	"a",
	"strong",
	"em",
	"b",
	"i",
	"u",
	"br",
	"hr",
	"table",
	"thead",
	"tbody",
	"tfoot",
	"tr",
	"th",
	"td",
	"pre",
	"code",
	"blockquote",
	"dl",
	"dt",
	"dd",
	"sub",
	"sup",
	"del",
	"ins",
	"abbr",
	"caption",
	"col",
	"colgroup",
	"figure",
	"figcaption",
]);

/** Attributes allowed on their respective tags (lowercased tag → set of attr names). */
const ALLOWED_ATTRS: Record<string, Set<string>> = {
	a: new Set(["href", "title", "rel", "target"]),
	td: new Set(["colspan", "rowspan", "headers"]),
	th: new Set(["colspan", "rowspan", "scope", "headers"]),
	abbr: new Set(["title"]),
};

/** Default attributes allowed on every tag. */
const GLOBAL_ATTRS = new Set(["class", "id", "lang", "dir"]);

/** Strip dangerous tags and their content. */
function stripDangerousTags(html: string): string {
	return html.replace(
		/<(script|style|iframe|object|embed|form|input|select|textarea|button|noscript|meta|link)\b[^>]*>[\s\S]*?<\/\1>/gi,
		"",
	);
}

/** Strip all HTML comments. */
function stripComments(html: string): string {
	return html.replace(/<!--[\s\S]*?-->/g, "");
}

/** Strip event-handler attributes. */
function stripEventHandlers(html: string): string {
	return html.replace(/\s+on\w+\s*=\s*(?:"[^"]*"|'[^']*'|[^\s>]+)/gi, "");
}

/** Remove javascript:/data: from href/src to prevent XSS. */
function sanitiseUrls(html: string): string {
	return html.replace(
		/(href|src)\s*=\s*(?:"|')(?:javascript|data|vbscript):/gi,
		"$1=''",
	);
}

/** Strip tags/attrs not in our allow-list. Returns safe HTML. */
export function sanitiseHtml(raw: string): string {
	let safe = raw;

	// 0. Unescape HTML entities so real tags can be filtered properly
	safe = unescapeHtml(safe);

	// 1. Strip dangerous block elements and their contents
	safe = stripDangerousTags(safe);
	// 2. Strip comments
	safe = stripComments(safe);
	// 3. Strip event handlers
	safe = stripEventHandlers(safe);
	// 4. Sanitise URLs
	safe = sanitiseUrls(safe);

	// 5. Remove disallowed attributes per tag while keeping allowed ones
	safe = safe.replace(
		/<(\/?)([a-z][a-z0-9]*)\b([^>]*?)>/gi,
		(full, slash, tagName, attrs) => {
			const tag = tagName.toLowerCase();

			if (!ALLOWED_TAGS.has(tag)) {
				return "";
			}

			if (!attrs.trim()) return full;

			const allowed = ALLOWED_ATTRS[tag] ?? new Set();
			const kept: string[] = [];
			const attrRe =
				/\s*([a-z][a-z0-9-]*)\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+))/gi;
			let m = attrRe.exec(attrs);
			while (m !== null) {
				const name = m[1].toLowerCase();
				const val = m[2] ?? m[3] ?? m[4] ?? "";
				if (GLOBAL_ATTRS.has(name) || allowed.has(name)) {
					kept.push(` ${name}="${val.replace(/"/g, "&quot;")}"`);
				}
				m = attrRe.exec(attrs);
			}

			return `<${slash}${tagName}${kept.join("")}>`;
		},
	);

	return safe;
}

/* ───────── Lightweight Markdown → HTML ───────── */

/** Convert a single line of inline markdown to HTML. */
function inlineMarkdown(line: string): string {
	return line
		.replace(/`([^`]+)`/g, "<code>$1</code>")
		.replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>")
		.replace(/\*([^*]+)\*/g, "<em>$1</em>")
		.replace(
			/\[([^\]]+)\]\(([^)]+)\)/g,
			'<a href="$2" rel="noopener noreferrer" target="_blank">$1</a>',
		)
		.replace(/!\[([^\]]*)\]\(([^)]+)\)/g, "")
		.replace(/~~([^~]+)~~/g, "<del>$1</del>");
}

/** Convert markdown text to HTML. Handles block-level elements. */
export function markdownToHtml(md: string): string {
	const lines = md.split("\n");
	const out: string[] = [];
	let inList: "ul" | "ol" | null = null;

	for (let i = 0; i < lines.length; i++) {
		const trimmed = lines[i].trim();

		// Close list if we left it
		if (
			inList &&
			!/^\s*[-*+]\s/.test(lines[i]) &&
			!/^\s*\d+\.\s/.test(lines[i])
		) {
			out.push(`</${inList}>`);
			inList = null;
		}

		// Empty line — paragraph break
		if (!trimmed) {
			out.push("</p><p>");
			continue;
		}

		// Heading
		const hMatch = trimmed.match(/^(#{1,6})\s+(.+)$/);
		if (hMatch) {
			const level = hMatch[1].length;
			out.push(`<h${level}>${inlineMarkdown(hMatch[2])}</h${level}>`);
			continue;
		}

		// Horizontal rule
		if (/^-{3,}$/.test(trimmed)) {
			out.push("<hr />");
			continue;
		}

		// Blockquote
		if (trimmed.startsWith("> ")) {
			out.push(`<blockquote>${inlineMarkdown(trimmed.slice(2))}</blockquote>`);
			continue;
		}

		// Unordered list
		const ulMatch = trimmed.match(/^\s*[-*+]\s+(.+)$/);
		if (ulMatch) {
			if (!inList) {
				out.push("<ul>");
				inList = "ul";
			}
			out.push(`<li>${inlineMarkdown(ulMatch[1])}</li>`);
			continue;
		}

		// Ordered list
		const olMatch = trimmed.match(/^\s*\d+\.\s+(.+)$/);
		if (olMatch) {
			if (!inList) {
				out.push("<ol>");
				inList = "ol";
			} else if (inList !== "ol") {
				out.push(`</${inList}><ol>`);
				inList = "ol";
			}
			out.push(`<li>${inlineMarkdown(olMatch[1])}</li>`);
			continue;
		}

		// Regular paragraph line
		out.push(inlineMarkdown(lines[i]));
	}

	if (inList) {
		out.push(`</${inList}>`);
	}

	let result = out.join("\n");

	if (
		!result.startsWith("<h") &&
		!result.startsWith("<ul") &&
		!result.startsWith("<ol") &&
		!result.startsWith("<blockquote") &&
		!result.startsWith("<hr")
	) {
		result = `<p>${result}</p>`;
	}

	result = result.replace(/<\/p>\s*<p>/g, "</p>\n<p>");
	result = result.replace(/<p><\/p>/g, "");
	result = result.replace(/<\/p>\n<(h[1-6]|ul|ol|blockquote|hr)/g, "\n<$1");

	return result;
}

/* ───────── Plain text → HTML ───────── */

/** Escape HTML entities and wrap in paragraphs. */
export function plainTextToHtml(text: string): string {
	const escaped = text
		.replace(/&/g, "&amp;")
		.replace(/</g, "&lt;")
		.replace(/>/g, "&gt;");
	return `<p>${escaped.replace(/\n\n/g, "</p><p>").replace(/\n/g, "<br />")}</p>`;
}

/* ───────── Format-specific dispatch ───────── */

export type DescriptionFormat = "markdown" | "html" | "plain";

/** Convert a description to safe HTML based on its format tag. */
export function formatToHtml(format: DescriptionFormat, raw: string): string {
	switch (format) {
		case "html":
			return sanitiseHtml(raw);
		case "markdown":
			return markdownToHtml(raw);
		default:
			return plainTextToHtml(raw);
	}
}
