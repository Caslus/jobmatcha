import type { DescriptionFormat } from "../../lib/description";
import { formatToHtml } from "../../lib/description";

interface JobDescriptionProps {
	description: string;
	format: DescriptionFormat;
	className?: string;
}

/**
 * Renders a job description in the format specified by the server's
 * `description_format` field.
 *
 * Accepts `"markdown"`, `"html"`, or `"plain"` and applies the appropriate
 * conversion + Tailwind Typography prose styling for the matcha dark theme.
 */
export function JobDescription({
	description,
	format,
	className = "",
}: JobDescriptionProps) {
	if (!description) {
		return (
			<p className="text-sm italic text-[#6a7a6a]">No description available</p>
		);
	}

	const safeHtml = formatToHtml(format, description);

	return (
		<div className={className}>
			<div
				className="prose prose-invert prose-sm max-w-none
          prose-headings:text-[#e8e8e8] prose-headings:font-bold prose-headings:tracking-tight
          prose-h1:text-lg prose-h1:mt-4 prose-h1:mb-2
          prose-h2:text-base prose-h2:mt-4 prose-h2:mb-2
          prose-h3:text-sm prose-h3:mt-3 prose-h3:mb-1
          prose-p:text-[#9a9a9a] prose-p:leading-relaxed prose-p:my-2
          prose-a:text-[#7dba7a] prose-a:no-underline hover:prose-a:underline
          prose-strong:text-[#c8c8c8] prose-strong:font-semibold
          prose-code:text-[#7dba7a] prose-code:bg-[#1a2a1a] prose-code:px-1 prose-code:py-0.5 prose-code:rounded prose-code:text-xs
          prose-pre:bg-[#0a0f0a] prose-pre:border prose-pre:border-[#1a2a1a] prose-pre:rounded-lg
          prose-blockquote:border-l-[#7dba7a] prose-blockquote:text-[#9a9a9a] prose-blockquote:pl-4 prose-blockquote:italic
          prose-ul:text-[#9a9a9a] prose-ol:text-[#9a9a9a]
          prose-li:my-0.5
          prose-hr:border-[#1a2a1a] prose-hr:my-4
          [&_table]:w-full [&_table]:border-collapse [&_table]:text-sm
          [&_th]:border [&_th]:border-[#1a2a1a] [&_th]:p-2 [&_th]:text-left [&_th]:text-[#c8c8c8] [&_th]:bg-[#0a0f0a] [&_th]:font-semibold
          [&_td]:border [&_td]:border-[#1a2a1a] [&_td]:p-2 [&_td]:text-[#9a9a9a]
          [&_del]:text-[#6a7a6a] [&_del]:line-through
          [&_ins]:text-[#7dba7a] [&_ins]:no-underline
          [&_abbr[title]]:border-b [&_abbr[title]]:border-dotted [&_abbr[title]]:border-[#4a5a4a] [&_abbr[title]]:cursor-help"
				// biome-ignore lint/security/noDangerouslySetInnerHtml: sanitised by formatToHtml()
				dangerouslySetInnerHTML={{ __html: safeHtml }}
			/>
		</div>
	);
}
