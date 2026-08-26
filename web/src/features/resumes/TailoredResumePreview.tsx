import { Download, RotateCcw } from "lucide-react";
import { useMemo } from "react";
import { Modal } from "@/components/ui/modal";
import { useLocalStorageState } from "@/hooks/useLocalStorageState";
import type {
	ResumeDocument,
	ResumeEntry,
	ResumeSection,
} from "@/types/api.gen";

interface TailoredResumePreviewProps {
	open: boolean;
	document: ResumeDocument;
	jobTitle: string;
	profileLinks: string[];
	onClose: () => void;
}

type FontChoice = "modern" | "classic" | "clean";
type LinkColor = "accent" | "text";

interface ResumeAppearance {
	font: FontChoice;
	fontSize: number;
	lineHeight: number;
	sectionSpacing: number;
	pageMargin: number;
	accentColor: string;
	linkColor: LinkColor;
	underlineLinks: boolean;
}

interface ResumePage {
	showHeader: boolean;
	showSummary: boolean;
	sections: ResumeSection[];
}

const defaultAppearance: ResumeAppearance = {
	font: "modern",
	fontSize: 10,
	lineHeight: 1.33,
	sectionSpacing: 11,
	pageMargin: 13,
	accentColor: "#245a3c",
	linkColor: "text",
	underlineLinks: true,
};

const APPEARANCE_STORAGE_KEY = "jobmatcha_resume_appearance";

const fontFamilies: Record<FontChoice, string> = {
	modern:
		'Inter, ui-sans-serif, system-ui, -apple-system, "Segoe UI", sans-serif',
	classic: 'Georgia, "Times New Roman", serif',
	clean: "Arial, Helvetica, sans-serif",
};

function safeDocument(document: ResumeDocument): ResumeDocument {
	return {
		...document,
		header: document.header ?? { name: "", contact: [] },
		summary: document.summary ?? "",
		sections: document.sections ?? [],
	};
}

function withProfileLinks(
	document: ResumeDocument,
	profileLinks: string[],
): ResumeDocument {
	const contact = [...document.header.contact];
	for (const link of profileLinks) {
		if (!contact.some((value) => value.trim() === link.trim())) {
			contact.push(link.trim());
		}
	}
	return { ...document, header: { ...document.header, contact } };
}

function escapeHtml(value: string) {
	return value
		.replaceAll("&", "&amp;")
		.replaceAll("<", "&lt;")
		.replaceAll(">", "&gt;")
		.replaceAll('"', "&quot;")
		.replaceAll("'", "&#039;");
}

function contactHref(contact: string) {
	const value = contact.trim();
	if (/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) return `mailto:${value}`;
	if (/^https?:\/\//i.test(value)) return value;
	if (/^(?:www\.)?(?:linkedin\.com|github\.com)\//i.test(value)) {
		return `https://${value}`;
	}
	return null;
}

function contactLabel(contact: string) {
	const hostname = contact
		.trim()
		.replace(/^https?:\/\//i, "")
		.replace(/^www\./i, "")
		.toLowerCase();
	if (hostname.startsWith("linkedin.com/")) return "LinkedIn";
	if (hostname.startsWith("github.com/")) return "GitHub";
	return contact;
}

function contactHTML(contact: string) {
	const href = contactHref(contact);
	const text = escapeHtml(contactLabel(contact));
	return href ? `<a href="${escapeHtml(href)}">${text}</a>` : text;
}

function estimatedLines(value: string, charsPerLine: number) {
	return Math.max(1, Math.ceil(value.length / charsPerLine));
}

function entryHeight(entry: ResumeEntry, charsPerLine: number) {
	return (
		2 +
		entry.highlights.reduce(
			(total, highlight) => total + estimatedLines(highlight, charsPerLine),
			0,
		)
	);
}

function isCompactList(section: ResumeSection) {
	return (
		section.kind === "list" &&
		(section.items.length >= 10 ||
			(section.items.length === 1 && section.items[0].includes(" · ")))
	);
}

// Pages are composed from semantic blocks, so no experience or list row is
// cut in half. The same blocks are used for print output below.
function paginateResume(
	document: ResumeDocument,
	appearance: ResumeAppearance,
): ResumePage[] {
	const { fontSize, lineHeight, pageMargin } = appearance;
	const charsPerLine = Math.max(62, Math.floor(104 - (fontSize - 9) * 9));
	// A4 at CSS's 96dpi is 794 × 1123px. After the 13mm print margins,
	// this leaves about 750px of content height, matching the export template.
	const pageCapacity = Math.max(
		42,
		Math.floor((750 + (13 - pageMargin) * 15) / (fontSize * lineHeight)),
	);
	const pages: ResumePage[] = [];
	let current: ResumePage = {
		showHeader: true,
		showSummary: Boolean(document.summary),
		sections: [],
	};
	let used =
		3 +
		document.header.contact.length * 0.5 +
		(document.summary ? estimatedLines(document.summary, charsPerLine) + 3 : 0);

	const newPage = () => {
		pages.push(current);
		current = { showHeader: false, showSummary: false, sections: [] };
		used = 0;
	};

	const addEntry = (section: ResumeSection, entry: ResumeEntry) => {
		const cost = entryHeight(entry, charsPerLine);
		let target = current.sections.find(
			(candidate) => candidate.heading === section.heading,
		);
		const headingCost = target ? 0 : 2;
		if (used + headingCost + cost > pageCapacity && used > 8) {
			newPage();
			target = undefined;
		}
		if (!target) {
			target = {
				heading: section.heading,
				kind: section.kind,
				entries: [],
				items: [],
			};
			current.sections.push(target);
			used += 2;
		}
		target.entries.push(entry);
		used += cost;
	};

	const addItem = (section: ResumeSection, item: string) => {
		const cost = estimatedLines(item, charsPerLine);
		let target = current.sections.find(
			(candidate) => candidate.heading === section.heading,
		);
		const headingCost = target ? 0 : 2;
		if (used + headingCost + cost > pageCapacity && used > 8) {
			newPage();
			target = undefined;
		}
		if (!target) {
			target = {
				heading: section.heading,
				kind: section.kind,
				entries: [],
				items: [],
			};
			current.sections.push(target);
			used += 2;
		}
		target.items.push(item);
		used += cost;
	};

	for (const section of document.sections) {
		for (const entry of section.entries) addEntry(section, entry);
		if (isCompactList(section)) {
			addItem(section, section.items.join(" · "));
		} else {
			for (const item of section.items) addItem(section, item);
		}
	}
	pages.push(current);
	return pages;
}

function sectionHTML(section: ResumeSection) {
	const entries = section.entries
		.map((entry) => {
			const location = entry.location
				? `<span>${escapeHtml(entry.location)}</span>`
				: "";
			const dates = entry.date_range
				? `<span>${escapeHtml(entry.date_range)}</span>`
				: "";
			const highlights = entry.highlights
				.map((highlight) => `<li>${escapeHtml(highlight)}</li>`)
				.join("");
			return `<section class="entry"><div class="entry-row"><strong>${escapeHtml(entry.title)}</strong>${location}</div><div class="entry-row entry-meta"><em>${escapeHtml(entry.organization)}</em>${dates}</div>${highlights ? `<ul>${highlights}</ul>` : ""}</section>`;
		})
		.join("");
	const items = isCompactList(section)
		? `<p class="compact-items">${section.items.map(escapeHtml).join(" &middot; ")}</p>`
		: section.items.map((item) => `<li>${escapeHtml(item)}</li>`).join("");
	return `<section class="resume-section"><h2>${escapeHtml(section.heading)}</h2>${entries}${items ? (isCompactList(section) ? items : `<ul class="items">${items}</ul>`) : ""}</section>`;
}

function downloadPDF(
	document: ResumeDocument,
	jobTitle: string,
	appearance: ResumeAppearance,
) {
	const structuredDocument = safeDocument(document);
	const pages = paginateResume(structuredDocument, appearance);
	const popup = window.open("", "_blank", "width=900,height=700");
	if (!popup) return;
	const title = jobTitle ? `Tailored resume — ${jobTitle}` : "Tailored resume";
	const contact = structuredDocument.header.contact
		.map(contactHTML)
		.join(" &middot; ");
	const pageMarkup = pages
		.map(
			(page) =>
				`<main class="page">${page.showHeader ? `<h1>${escapeHtml(structuredDocument.header.name)}</h1>${contact ? `<p class="contact">${contact}</p>` : ""}` : ""}${page.showSummary ? `<section class="resume-section"><h2>Summary</h2><p class="summary">${escapeHtml(structuredDocument.summary)}</p></section>` : ""}${page.sections.map(sectionHTML).join("")}</main>`,
		)
		.join("");
	promptPrint(popup, title, pageMarkup, appearance);
}

function promptPrint(
	popup: Window,
	title: string,
	pageMarkup: string,
	appearance: ResumeAppearance,
) {
	popup.document.write(
		`<!doctype html><html><head><title>${escapeHtml(title)}</title><style>@page { size: A4; margin: 0; } body { background: #fff; color: #171717; font: ${appearance.fontSize}pt ${fontFamilies[appearance.font]}; line-height: ${appearance.lineHeight}; margin: 0; } .page { box-sizing: border-box; min-height: 297mm; padding: ${appearance.pageMargin}mm; position: relative; break-after: page; } .page:last-child { break-after: auto; } h1 { font-size: ${appearance.fontSize + 9}pt; margin: 0; text-align: center; } a { color: ${appearance.linkColor === "accent" ? appearance.accentColor : "inherit"}; text-decoration: ${appearance.underlineLinks ? "underline" : "none"}; } .contact { margin: 4px 0 ${appearance.sectionSpacing}px; text-align: center; } h2 { border-bottom: 1px solid ${appearance.accentColor}; color: ${appearance.accentColor}; font-size: ${appearance.fontSize}pt; letter-spacing: .09em; margin: ${appearance.sectionSpacing}px 0 5px; padding-bottom: 2px; text-transform: uppercase; } .summary { margin: 0; } .entry { break-inside: avoid; margin: ${Math.max(3, Math.round(appearance.sectionSpacing * 0.55))}px 0 ${Math.max(4, Math.round(appearance.sectionSpacing * 0.7))}px; } .entry-row { align-items: baseline; display: flex; gap: 12px; justify-content: space-between; } .entry-row > span { color: #667085; text-align: right; white-space: nowrap; } .entry-meta { font-size: ${Math.max(8, appearance.fontSize - 0.5)}pt; } .entry-meta em { font-style: italic; } ul { margin: 2px 0 0 17px; padding: 0; } li { margin: 1px 0; } .items { margin-top: 3px; } .compact-items { margin: 2px 0 0; } footer { bottom: ${Math.max(6, appearance.pageMargin - 5)}mm; color: #667085; font-size: 8pt; position: absolute; right: ${appearance.pageMargin}mm; }</style></head><body>${pageMarkup}<script>window.onload = () => window.print();</script></body></html>`,
	);
	popup.document.close();
}

function ResumeSectionView({
	section,
	appearance,
	first,
}: {
	section: ResumeSection;
	appearance: ResumeAppearance;
	first: boolean;
}) {
	return (
		<section style={{ marginTop: first ? 0 : appearance.sectionSpacing }}>
			<h3
				className="border-b pb-1 font-bold uppercase tracking-widest"
				style={{
					borderColor: appearance.accentColor,
					color: appearance.accentColor,
					fontSize: "1em",
					marginBottom: "5px",
				}}
			>
				{section.heading}
			</h3>
			{section.entries.map((entry) => (
				<section
					key={`${entry.title}-${entry.organization}-${entry.date_range}`}
					className="break-inside-avoid"
					style={{
						marginTop: Math.max(
							3,
							Math.round(appearance.sectionSpacing * 0.55),
						),
						marginBottom: Math.max(
							4,
							Math.round(appearance.sectionSpacing * 0.7),
						),
					}}
				>
					<div className="grid grid-cols-[minmax(0,1fr)_auto] items-baseline gap-x-4">
						<h4 className="font-semibold text-[#171717]">{entry.title}</h4>
						{entry.location && (
							<span className="text-right text-[0.9em] text-[#667085]">
								{entry.location}
							</span>
						)}
					</div>
					<div className="grid grid-cols-[minmax(0,1fr)_auto] items-baseline gap-x-4 text-[0.9em]">
						<p className="italic text-[#2d2d2d]">{entry.organization}</p>
						{entry.date_range && (
							<span className="text-right text-[#667085]">
								{entry.date_range}
							</span>
						)}
					</div>
					{entry.highlights.length > 0 && (
						<ul className="mt-1 list-disc space-y-0.5 pl-5">
							{entry.highlights.map((highlight) => (
								<li key={highlight}>{highlight}</li>
							))}
						</ul>
					)}
				</section>
			))}
			{isCompactList(section) ? (
				<p>{section.items.join(" · ")}</p>
			) : section.items.length > 0 ? (
				<ul className="list-disc space-y-0.5 pl-5">
					{section.items.map((item) => (
						<li key={item}>{item}</li>
					))}
				</ul>
			) : null}
		</section>
	);
}

function ResumePageView({
	page,
	document,
	appearance,
}: {
	page: ResumePage;
	document: ResumeDocument;
	appearance: ResumeAppearance;
}) {
	return (
		<article
			className="relative mx-auto min-h-[1123px] w-[794px] bg-[#fdfdfb] px-10 py-9 text-[#171717] shadow-xl"
			style={{
				width: "794px",
				minHeight: "1123px",
				boxSizing: "border-box",
				backgroundColor: "#fdfdfb",
				padding: `${appearance.pageMargin * 3.7795}px`,
				fontFamily: fontFamilies[appearance.font],
				fontSize: `${appearance.fontSize}pt`,
				lineHeight: appearance.lineHeight,
			}}
		>
			{page.showHeader && (
				<>
					<h1 className="text-center text-[1.9em] font-bold tracking-tight">
						{document.header.name}
					</h1>
					{document.header.contact.length > 0 && (
						<p className="mt-1 text-center text-[0.9em] text-[#3d3d3d]">
							{document.header.contact.map((contact, index) => {
								const href = contactHref(contact);
								return (
									<span key={contact}>
										{index > 0 && " · "}
										{href ? (
											<a
												href={href}
												target="_blank"
												rel="noreferrer"
												className={
													appearance.underlineLinks
														? "underline decoration-[#667085]/60 underline-offset-2"
														: undefined
												}
												style={{
													color:
														appearance.linkColor === "accent"
															? appearance.accentColor
															: "inherit",
													textDecoration: appearance.underlineLinks
														? "underline"
														: "none",
												}}
											>
												{contactLabel(contact)}
											</a>
										) : (
											contact
										)}
									</span>
								);
							})}
						</p>
					)}
				</>
			)}
			{page.showSummary && (
				<section style={{ marginTop: appearance.sectionSpacing }}>
					<h3
						className="mb-2 border-b pb-1 text-[0.82em] font-bold uppercase tracking-widest"
						style={{
							borderColor: appearance.accentColor,
							color: appearance.accentColor,
						}}
					>
						Summary
					</h3>
					<p>{document.summary}</p>
				</section>
			)}
			{page.sections.map((section, index) => (
				<ResumeSectionView
					key={`${section.heading}-${section.entries[0]?.title ?? section.items[0] ?? ""}`}
					section={section}
					appearance={appearance}
					first={index === 0}
				/>
			))}
		</article>
	);
}

function AppearanceSlider({
	label,
	value,
	min,
	max,
	step,
	valueLabel,
	onChange,
}: {
	label: string;
	value: number;
	min: number;
	max: number;
	step: number;
	valueLabel: string;
	onChange: (value: number) => void;
}) {
	const percentage = ((value - min) / (max - min)) * 100;

	return (
		<div className="mb-4">
			<div className="mb-2 flex items-center justify-between text-xs font-medium text-[#9a9a9a]">
				<span>{label}</span>
				<span className="tabular-nums">{valueLabel}</span>
			</div>
			<div className="relative h-1.5">
				<div className="absolute inset-0 rounded-full bg-[#1a2a1a]" />
				<div
					className="absolute inset-y-0 left-0 rounded-full bg-[#7dba7a] transition-[width] duration-100"
					style={{ width: `${percentage}%` }}
				/>
				<input
					type="range"
					aria-label={label}
					min={min}
					max={max}
					step={step}
					value={value}
					onChange={(event) => onChange(Number(event.target.value))}
					className="absolute inset-0 h-full w-full cursor-pointer appearance-none bg-transparent outline-none [&::-moz-range-thumb]:h-3.5 [&::-moz-range-thumb]:w-3.5 [&::-moz-range-thumb]:cursor-pointer [&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:border-0 [&::-moz-range-thumb]:bg-[#7dba7a] [&::-webkit-slider-thumb]:relative [&::-webkit-slider-thumb]:z-10 [&::-webkit-slider-thumb]:h-3.5 [&::-webkit-slider-thumb]:w-3.5 [&::-webkit-slider-thumb]:cursor-pointer [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-[#7dba7a] [&::-webkit-slider-thumb]:shadow-sm"
				/>
			</div>
		</div>
	);
}

function AppearanceControls({
	appearance,
	onChange,
}: {
	appearance: ResumeAppearance;
	onChange: (appearance: ResumeAppearance) => void;
}) {
	return (
		<aside className="border-r border-[#1a2a1a] bg-[#0d120d] p-4">
			<div className="mb-4 flex items-center justify-between">
				<h3 className="text-sm font-semibold text-[#e8e8e8]">Appearance</h3>
				<button
					type="button"
					onClick={() => onChange(defaultAppearance)}
					className="rounded p-1 text-[#6a7a6a] transition hover:bg-[#1a2a1a] hover:text-[#e8e8e8]"
					aria-label="Reset resume appearance"
				>
					<RotateCcw size={14} />
				</button>
			</div>
			<label className="mb-4 block text-xs font-medium text-[#9a9a9a]">
				Typeface
				<select
					value={appearance.font}
					onChange={(event) =>
						onChange({ ...appearance, font: event.target.value as FontChoice })
					}
					className="mt-1.5 w-full rounded-md border border-[#2a3a2a] bg-[#111811] px-2 py-1.5 text-sm text-[#e8e8e8] outline-none focus:border-[#7dba7a]"
				>
					<option value="modern">Modern sans</option>
					<option value="classic">Classic serif</option>
					<option value="clean">Clean Arial</option>
				</select>
			</label>
			<AppearanceSlider
				label="Font size"
				value={appearance.fontSize}
				min={9}
				max={12}
				step={0.5}
				valueLabel={`${appearance.fontSize} pt`}
				onChange={(fontSize) => onChange({ ...appearance, fontSize })}
			/>
			<AppearanceSlider
				label="Line height"
				value={appearance.lineHeight}
				min={1.1}
				max={1.55}
				step={0.05}
				valueLabel={appearance.lineHeight.toFixed(2)}
				onChange={(lineHeight) => onChange({ ...appearance, lineHeight })}
			/>
			<AppearanceSlider
				label="Section spacing"
				value={appearance.sectionSpacing}
				min={6}
				max={18}
				step={1}
				valueLabel={`${appearance.sectionSpacing}px`}
				onChange={(sectionSpacing) =>
					onChange({ ...appearance, sectionSpacing })
				}
			/>
			<label className="mb-4 block text-xs font-medium text-[#9a9a9a]">
				Page margins
				<select
					value={appearance.pageMargin}
					onChange={(event) =>
						onChange({ ...appearance, pageMargin: Number(event.target.value) })
					}
					className="mt-1.5 w-full rounded-md border border-[#2a3a2a] bg-[#111811] px-2 py-1.5 text-sm text-[#e8e8e8] outline-none focus:border-[#7dba7a]"
				>
					<option value="10">Compact · 10 mm</option>
					<option value="13">Standard · 13 mm</option>
					<option value="16">Spacious · 16 mm</option>
				</select>
			</label>
			<label className="block text-xs font-medium text-[#9a9a9a]">
				Accent color
				<span className="mt-1.5 flex items-center gap-2">
					<input
						type="color"
						value={appearance.accentColor}
						onChange={(event) =>
							onChange({ ...appearance, accentColor: event.target.value })
						}
						className="h-8 w-10 cursor-pointer rounded border border-[#2a3a2a] bg-transparent p-0.5"
					/>
					<span className="font-mono text-xs text-[#c8c8c8]">
						{appearance.accentColor.toUpperCase()}
					</span>
				</span>
			</label>
			<fieldset className="mt-4">
				<legend className="text-xs font-medium text-[#9a9a9a]">
					Link color
				</legend>
				<div className="mt-1.5 grid grid-cols-2 overflow-hidden rounded-md border border-[#2a3a2a]">
					{(["text", "accent"] as const).map((linkColor) => (
						<button
							key={linkColor}
							type="button"
							onClick={() => onChange({ ...appearance, linkColor })}
							className={`px-2 py-1.5 text-xs font-medium transition ${
								appearance.linkColor === linkColor
									? "bg-[#245a3c] text-white"
									: "text-[#9a9a9a] hover:bg-[#1a2a1a]"
							}`}
						>
							{linkColor === "text" ? "Text" : "Accent"}
						</button>
					))}
				</div>
			</fieldset>
			<fieldset className="mt-3">
				<legend className="text-xs font-medium text-[#9a9a9a]">
					Underline links
				</legend>
				<div className="mt-1.5 grid grid-cols-2 overflow-hidden rounded-md border border-[#2a3a2a]">
					{([true, false] as const).map((underlineLinks) => (
						<button
							key={String(underlineLinks)}
							type="button"
							aria-pressed={appearance.underlineLinks === underlineLinks}
							onClick={() => onChange({ ...appearance, underlineLinks })}
							className={`px-2 py-1.5 text-xs font-medium transition ${
								appearance.underlineLinks === underlineLinks
									? "bg-[#245a3c] text-white"
									: "text-[#9a9a9a] hover:bg-[#1a2a1a]"
							}`}
						>
							{underlineLinks ? "On" : "Off"}
						</button>
					))}
				</div>
			</fieldset>
		</aside>
	);
}

export function TailoredResumePreview({
	open,
	document: resumeDocument,
	jobTitle,
	profileLinks,
	onClose,
}: TailoredResumePreviewProps) {
	const [savedAppearance, setSavedAppearance] = useLocalStorageState<
		Partial<ResumeAppearance>
	>(APPEARANCE_STORAGE_KEY, {});
	const appearance = useMemo(
		() => ({ ...defaultAppearance, ...savedAppearance }),
		[savedAppearance],
	);
	const updateAppearance = (nextAppearance: ResumeAppearance) =>
		setSavedAppearance(nextAppearance);
	const structuredDocument = withProfileLinks(
		safeDocument(resumeDocument),
		profileLinks,
	);
	const isStructured =
		structuredDocument.header.name || structuredDocument.sections.length > 0;
	const pages = useMemo(
		() => paginateResume(structuredDocument, appearance),
		[structuredDocument, appearance],
	);

	return (
		<Modal
			open={open}
			onClose={onClose}
			title="Tailored resume preview"
			subtitle={jobTitle || "Job-specific wording edits"}
			closeLabel="Close tailored resume preview"
			panelClassName="h-[92vh] max-w-[1120px]"
			headerActions={
				isStructured ? (
					<button
						type="button"
						onClick={() =>
							downloadPDF(structuredDocument, jobTitle, appearance)
						}
						className="inline-flex items-center gap-1.5 rounded-md border border-[#4f7b4d] px-3 py-2 text-xs font-medium text-[#9bd398] transition hover:bg-[#173017]"
					>
						<Download size={14} />
						Save PDF
					</button>
				) : null
			}
		>
			<div
				className="grid min-h-0 flex-1"
				style={{
					gridTemplateColumns: "210px minmax(0, 1fr)",
					minHeight: 0,
					flex: "1 1 auto",
					overflow: "hidden",
				}}
			>
				<AppearanceControls
					appearance={appearance}
					onChange={updateAppearance}
				/>
				<div
					className="overflow-auto bg-[#080b08] p-5 sm:p-8"
					style={{ minWidth: 0, overflow: "auto" }}
				>
					{isStructured ? (
						<div
							className="mx-auto flex w-max flex-col gap-8"
							style={{
								display: "flex",
								width: "794px",
								flexDirection: "column",
								gap: "32px",
							}}
						>
							{pages.map((page) => (
								<ResumePageView
									key={
										page.sections
											.map(
												(section) =>
													`${section.heading}-${section.entries.map((entry) => entry.title).join("-")}-${section.items.join("-")}`,
											)
											.join("|") || "resume-header"
									}
									page={page}
									document={structuredDocument}
									appearance={appearance}
								/>
							))}
						</div>
					) : (
						<p className="mx-auto max-w-xl py-16 text-center text-sm text-[#9a9a9a]">
							This earlier tailored resume uses the legacy format. Tailor it
							again to generate the structured preview.
						</p>
					)}
				</div>
			</div>
		</Modal>
	);
}
