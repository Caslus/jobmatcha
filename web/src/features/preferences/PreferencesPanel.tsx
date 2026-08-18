import { forwardRef, useEffect, useRef, useState } from "react";
import { useSettings, useUpdateSettings } from "../../hooks/useApi";
import { ScanStatusBar } from "../scan/ScanStatusBar";

// ── Types ──────────────────────────────────────────

type KeywordField = "include" | "exclude" | "location";
type FeedbackState = "idle" | "saving" | "done" | "error";
type ChipVariant = "green" | "red" | "amber";

interface Draft {
	include: string[];
	exclude: string[];
	location: string[];
	workTypes: string[];
}

interface KWProps {
	label: string;
	field: KeywordField;
	keywords: string[];
	showInput: boolean;
	onAddInput: () => void;
	onCloseInput: () => void;
	inputRef: React.RefObject<HTMLInputElement | null>;
	highlight: ChipVariant;
	onStageAdd: (f: KeywordField, v: string) => void;
	onChipClick: (kw: string, f: KeywordField) => void;
	markedForDelete: Set<string>;
}

// ── Chip Colors ────────────────────────────────────

const chipColors: Record<
	string,
	{ base: string; hover: string; marked: string }
> = {
	green: {
		base: "bg-[rgba(125,186,122,0.08)] text-[#7dba7a]/80 border-[rgba(125,186,122,0.15)]",
		hover:
			"hover:bg-[rgba(125,186,122,0.18)] hover:text-[#7dba7a] hover:border-[rgba(125,186,122,0.35)] hover:line-through hover:decoration-[#7dba7a]/60",
		marked:
			"bg-[rgba(239,68,68,0.08)] text-red-400/60 border-[rgba(239,68,68,0.15)]",
	},
	red: {
		base: "bg-[rgba(239,68,68,0.06)] text-red-400/70 border-[rgba(239,68,68,0.12)]",
		hover:
			"hover:bg-[rgba(239,68,68,0.15)] hover:text-red-400 hover:border-[rgba(239,68,68,0.3)] hover:line-through hover:decoration-red-400/60",
		marked:
			"bg-[rgba(239,68,68,0.12)] text-red-400/50 border-[rgba(239,68,68,0.2)]",
	},
	amber: {
		base: "bg-[rgba(245,158,11,0.06)] text-amber-400/70 border-[rgba(245,158,11,0.12)]",
		hover:
			"hover:bg-[rgba(245,158,11,0.15)] hover:text-amber-400 hover:border-[rgba(245,158,11,0.3)] hover:line-through hover:decoration-amber-400/60",
		marked:
			"bg-[rgba(245,158,11,0.08)] text-amber-400/50 border-[rgba(245,158,11,0.15)]",
	},
};

// ── ChipInput (inline add) ─────────────────────────

interface ChipInputProps {
	onSubmit: (v: string) => void;
	onCancel: () => void;
	placeholder: string;
}

const ChipInput = forwardRef<HTMLInputElement, ChipInputProps>(
	({ onSubmit, onCancel, placeholder }, ref) => {
		const [value, setValue] = useState("");
		return (
			<input
				ref={ref}
				type="text"
				value={value}
				onChange={(e) => setValue(e.target.value)}
				onKeyDown={(e) => {
					if (e.key === "Enter") {
						e.preventDefault();
						if (value.trim()) onSubmit(value.trim());
					} else if (e.key === "Escape") onCancel();
				}}
				onBlur={() =>
					setTimeout(() => {
						if (value.trim()) onSubmit(value.trim());
						else onCancel();
					}, 150)
				}
				placeholder={placeholder}
				className="mb-2 w-full rounded-md border border-[rgba(255,255,255,0.06)] bg-[rgba(0,0,0,0.25)] px-2.5 py-1.5 text-xs text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[rgba(125,186,122,0.3)]"
			/>
		);
	},
);
ChipInput.displayName = "ChipInput";

// ── KeywordSection ─────────────────────────────────

function KeywordSection({
	label,
	field,
	keywords,
	showInput,
	onAddInput,
	onCloseInput,
	inputRef,
	highlight,
	onStageAdd,
	onChipClick,
	markedForDelete,
}: KWProps) {
	const c = chipColors[highlight];
	return (
		<div>
			<label className="mb-2 block text-xs font-medium uppercase tracking-[0.05em] text-[#4a5a4a]">
				{label}
			</label>
			{showInput && (
				<ChipInput
					ref={inputRef}
					onSubmit={(v) => {
						onStageAdd(field, v);
						onCloseInput();
					}}
					onCancel={onCloseInput}
					placeholder={`${label.toLowerCase()}, ...`}
				/>
			)}
			<div className="flex flex-wrap gap-1.5">
				{keywords.length === 0 ? (
					<span className="text-xs text-[#3a4a3a]">None set</span>
				) : (
					keywords.map((kw) => (
						<button
							key={kw}
							type="button"
							onClick={() => onChipClick(kw, field)}
							className={`rounded-full border px-2.5 py-1 text-sm font-medium transition-all duration-150 ${markedForDelete.has(kw) ? `${c.marked} line-through` : `${c.base} ${c.hover}`}`}
						>
							{kw}
						</button>
					))
				)}
				{!showInput && (
					<button
						type="button"
						onClick={onAddInput}
						className="inline-flex items-center gap-0.5 rounded-full border border-dashed border-[rgba(255,255,255,0.08)] px-2.5 py-1 text-sm text-[#5a6a5a] transition hover:border-[rgba(255,255,255,0.18)] hover:text-[#8a9a8a]"
					>
						<svg width="10" height="10" viewBox="0 0 10 10" fill="none">
							<path
								d="M5 1v8M1 5h8"
								stroke="currentColor"
								strokeWidth="1.5"
								strokeLinecap="round"
							/>
						</svg>
						Add
					</button>
				)}
			</div>
		</div>
	);
}

// ── WorkTypeToggle (extracted from inline map) ─────

interface WorkTypeToggleProps {
	label: string;
	selected: boolean;
	onToggle: () => void;
}

function WorkTypeToggle({ label, selected, onToggle }: WorkTypeToggleProps) {
	return (
		<button
			type="button"
			onClick={onToggle}
			className={`rounded-full border px-2.5 py-1 text-sm font-medium transition-all duration-150 ${
				selected
					? "bg-[rgba(125,186,122,0.12)] text-[#7dba7a] border-[rgba(125,186,122,0.2)]"
					: "bg-transparent text-[#6a7a6a] border-[rgba(255,255,255,0.06)] hover:text-[#8a9a8a] hover:border-[rgba(255,255,255,0.12)]"
			}`}
		>
			{label}
		</button>
	);
}

// ── PreferencesPanel (main export) ─────────────────

export function PreferencesPanel() {
	const { data: settings, isLoading } = useSettings();
	const updateSettings = useUpdateSettings();
	const [showInputs, setShowInputs] = useState<Set<string>>(new Set());
	const inputRefs = {
		include: useRef<HTMLInputElement>(null),
		exclude: useRef<HTMLInputElement>(null),
		location: useRef<HTMLInputElement>(null),
	};

	const [draft, setDraft] = useState<Draft | null>(null);
	const [markedForDelete, setMarkedForDelete] = useState<Set<string>>(
		new Set(),
	);
	const [applyFeedback, setApplyFeedback] = useState<FeedbackState>("idle");

	// Sync draft from server settings
	useEffect(() => {
		if (settings)
			setDraft({
				include: [...settings.include_keywords],
				exclude: [...settings.exclude_keywords],
				location: [...settings.location_keywords],
				workTypes: [...(settings.work_types ?? [])],
			});
	}, [settings]);

	// Auto-focus input when showing one
	useEffect(() => {
		for (const key of showInputs) {
			const ref = inputRefs[key as keyof typeof inputRefs];
			if (ref.current) {
				ref.current.focus();
				break;
			}
		}
	}, [showInputs]);

	const stageAdd = (field: KeywordField, rawValue: string) => {
		const keywords = rawValue
			.split(",")
			.map((s) => s.trim().toLowerCase())
			.filter((s) => s.length > 0);
		if (keywords.length === 0) return;
		setDraft((prev) => {
			if (!prev) return prev;
			const arr = [...prev[field]];
			for (const kw of keywords) {
				if (!arr.includes(kw)) arr.push(kw);
			}
			return { ...prev, [field]: arr };
		});
	};

	const toggleDelete = (keyword: string) => {
		setMarkedForDelete((prev) => {
			const next = new Set(prev);
			if (next.has(keyword)) next.delete(keyword);
			else next.add(keyword);
			return next;
		});
	};

	const handleChipClick = (kw: string, field: KeywordField) => {
		const key = `${field}_keywords` as
			| "include_keywords"
			| "exclude_keywords"
			| "location_keywords";
		if (settings && !settings[key].includes(kw)) {
			setDraft((prev) =>
				prev ? { ...prev, [field]: prev[field].filter((k) => k !== kw) } : prev,
			);
		} else {
			toggleDelete(kw);
		}
	};

	const handleApply = () => {
		if (!draft) return;
		setApplyFeedback("saving");
		updateSettings.mutate(
			{
				include_keywords: draft.include.filter((k) => !markedForDelete.has(k)),
				exclude_keywords: draft.exclude.filter((k) => !markedForDelete.has(k)),
				location_keywords: draft.location.filter(
					(k) => !markedForDelete.has(k),
				),
				work_types: draft.workTypes,
			},
			{
				onSuccess: () => {
					setMarkedForDelete(new Set());
					setApplyFeedback("done");
					setTimeout(() => setApplyFeedback("idle"), 2000);
				},
				onError: () => {
					setApplyFeedback("error");
					setTimeout(() => setApplyFeedback("idle"), 3000);
				},
			},
		);
	};

	const hasChanges =
		markedForDelete.size > 0 ||
		(draft &&
			settings &&
			(draft.include.length !== settings.include_keywords.length ||
				draft.exclude.length !== settings.exclude_keywords.length ||
				draft.location.length !== settings.location_keywords.length ||
				draft.workTypes.length !== (settings.work_types ?? []).length ||
				draft.workTypes.some((t, i) => t !== (settings.work_types ?? [])[i])));

	// ── Loading state ──────────────────────────────
	if (isLoading) {
		return (
			<aside className="h-full p-4">
				<div className="space-y-3">
					{[1, 2, 3].map((i) => (
						<div
							key={i}
							className="h-16 animate-pulse rounded-lg bg-[#0d120d]"
						/>
					))}
				</div>
			</aside>
		);
	}

	const display: Draft = draft ?? {
		include: [],
		exclude: [],
		location: [],
		workTypes: [],
	};

	// ── Render ─────────────────────────────────────
	return (
		<aside className="flex h-full flex-col">
			<div className="flex-1 overflow-y-auto p-4">
				<div className="mb-5 flex items-center justify-between">
					<h3 className="text-sm font-semibold uppercase tracking-wider text-[#6a7a6a]">
						Preferences
					</h3>
					<button
						onClick={handleApply}
						disabled={!hasChanges || applyFeedback === "saving"}
						className={`rounded-lg px-3 py-1.5 text-sm font-semibold transition ${
							hasChanges && applyFeedback !== "saving"
								? "bg-gradient-to-r from-[#7dba7a] to-[#5a8f5a] text-[#080908] hover:from-[#8dca8a] hover:to-[#6a9f6a]"
								: "cursor-not-allowed bg-[#1a2a1a] text-[#4a5a4a]"
						}`}
					>
						{applyFeedback === "saving"
							? "Saving..."
							: applyFeedback === "done"
								? "Applied ✓"
								: applyFeedback === "error"
									? "Error!"
									: "Apply"}
					</button>
				</div>

				<div className="space-y-5">
					<KeywordSection
						label="Include"
						field="include"
						keywords={display.include}
						showInput={showInputs.has("include")}
						onAddInput={() =>
							setShowInputs((prev) => {
								const n = new Set(prev);
								n.add("include");
								return n;
							})
						}
						onCloseInput={() =>
							setShowInputs((prev) => {
								const n = new Set(prev);
								n.delete("include");
								return n;
							})
						}
						inputRef={inputRefs.include}
						highlight="green"
						onStageAdd={stageAdd}
						onChipClick={handleChipClick}
						markedForDelete={markedForDelete}
					/>
					<KeywordSection
						label="Exclude"
						field="exclude"
						keywords={display.exclude}
						showInput={showInputs.has("exclude")}
						onAddInput={() =>
							setShowInputs((prev) => {
								const n = new Set(prev);
								n.add("exclude");
								return n;
							})
						}
						onCloseInput={() =>
							setShowInputs((prev) => {
								const n = new Set(prev);
								n.delete("exclude");
								return n;
							})
						}
						inputRef={inputRefs.exclude}
						highlight="red"
						onStageAdd={stageAdd}
						onChipClick={handleChipClick}
						markedForDelete={markedForDelete}
					/>
					<KeywordSection
						label="Location"
						field="location"
						keywords={display.location}
						showInput={showInputs.has("location")}
						onAddInput={() =>
							setShowInputs((prev) => {
								const n = new Set(prev);
								n.add("location");
								return n;
							})
						}
						onCloseInput={() =>
							setShowInputs((prev) => {
								const n = new Set(prev);
								n.delete("location");
								return n;
							})
						}
						inputRef={inputRefs.location}
						highlight="amber"
						onStageAdd={stageAdd}
						onChipClick={handleChipClick}
						markedForDelete={markedForDelete}
					/>

					<div>
						<label className="mb-2 block text-xs font-medium uppercase tracking-[0.05em] text-[#4a5a4a]">
							Work Type
						</label>
						<div className="flex flex-wrap gap-1.5">
							<WorkTypeToggle
								label="Any"
								selected={display.workTypes.length === 0}
								onToggle={() =>
									setDraft((prev) => (prev ? { ...prev, workTypes: [] } : prev))
								}
							/>
							{["internship", "contract", "part-time", "full-time"].map((t) => (
								<WorkTypeToggle
									key={t}
									label={
										t === "full-time"
											? "Full-time"
											: t.charAt(0).toUpperCase() + t.slice(1)
									}
									selected={display.workTypes.includes(t)}
									onToggle={() =>
										setDraft((prev) =>
											prev
												? {
														...prev,
														workTypes: prev.workTypes.includes(t)
															? prev.workTypes.filter((x) => x !== t)
															: [...prev.workTypes, t],
													}
												: prev,
										)
									}
								/>
							))}
						</div>
					</div>
				</div>
			</div>
			<ScanStatusBar />
		</aside>
	);
}
