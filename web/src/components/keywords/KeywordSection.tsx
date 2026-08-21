import { forwardRef, useState } from "react";

// ── Types ────────────────────────────────────────

export type KeywordField = "include" | "exclude" | "location";

type ChipVariant = "green" | "red" | "amber";

// ── Chip Colors ──────────────────────────────────

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

// ── ChipInput (inline add) ───────────────────────

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

// ── KeywordSection ───────────────────────────────

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

export function KeywordSection({
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
			<span className="mb-2 block text-xs font-medium uppercase tracking-wider text-[#4a5a4a]">
				{label}
			</span>
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
							className={`rounded-full border px-2.5 py-1 text-sm font-medium transition-all duration-150 ${
								markedForDelete.has(kw)
									? `${c.marked} line-through`
									: `${c.base} ${c.hover}`
							}`}
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
						<svg
							width="10"
							height="10"
							viewBox="0 0 10 10"
							fill="none"
							aria-hidden="true"
						>
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

// ── WorkTypeToggle ───────────────────────────────

interface WorkTypeToggleProps {
	label: string;
	selected: boolean;
	onToggle: () => void;
}

export function WorkTypeToggle({
	label,
	selected,
	onToggle,
}: WorkTypeToggleProps) {
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
