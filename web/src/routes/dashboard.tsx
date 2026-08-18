import { createFileRoute, redirect, useNavigate } from "@tanstack/react-router";
import { LogOut } from "lucide-react";
import { forwardRef, useEffect, useRef, useState } from "react";
import { RoleDetailPanel } from "../features/jobs/RoleDetailPanel";
import { RoleList } from "../features/jobs/RoleList";
import { ScanStatusBar } from "../features/scan/ScanStatusBar";
import { useSettings, useUpdateSettings } from "../hooks/useApi";
import { authApi } from "../lib/api";
import { useAuthStore } from "../stores/auth";

export const Route = createFileRoute("/dashboard")({
	beforeLoad: () => {
		const { authenticated } = useAuthStore.getState();
		if (!authenticated) throw redirect({ to: "/" });
	},
	component: DashboardPage,
});

function DashboardPage() {
	const navigate = useNavigate();
	const { check } = useAuthStore();
	const [selectedId, setSelectedId] = useState<number | null>(null);

	useEffect(() => {
		check().then(() => {
			const { setupComplete } = useAuthStore.getState();
			if (!setupComplete) navigate({ to: "/onboarding" });
		});
	}, []);

	const handleLogout = async () => {
		try {
			await authApi.logout();
		} catch {}
		useAuthStore.getState().logout();
		window.location.href = "/";
	};

	return (
		<div className="flex h-screen flex-col bg-[#080908] text-[#e8e8e8]">
			<header className="flex items-center justify-between border-b border-[#1a2a1a] px-6 py-3">
				<div className="flex items-center gap-3">
					<div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-[#7dba7a] to-[#4a7c4f]">
						<span className="text-sm">🍵</span>
					</div>
					<span className="text-lg font-bold text-[#e8e8e8]">jobmatcha</span>
				</div>
				<div className="flex items-center gap-4">
					<button
						onClick={handleLogout}
						className="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs text-[#6a7a6a] transition hover:bg-[#1a2a1a] hover:text-[#e8e8e8]"
					>
						<LogOut size={14} /> Logout
					</button>
				</div>
			</header>

			<div className="flex flex-1 overflow-hidden">
				<PreferencesPanel />
				<RoleList selectedId={selectedId} onSelect={setSelectedId} />
				<RoleDetailPanel
					selectedId={selectedId}
					onBack={() => setSelectedId(null)}
				/>
			</div>
		</div>
	);
}

// ---- Preferences Panel ----

function PreferencesPanel() {
	const { data: settings, isLoading } = useSettings();
	const updateSettings = useUpdateSettings();
	const [showInputs, setShowInputs] = useState<Set<string>>(new Set());
	const inputRefs = {
		include: useRef<HTMLInputElement>(null),
		exclude: useRef<HTMLInputElement>(null),
		location: useRef<HTMLInputElement>(null),
	};

	const [draft, setDraft] = useState<{
		include: string[];
		exclude: string[];
		location: string[];
		workTypes: string[];
	} | null>(null);
	const [markedForDelete, setMarkedForDelete] = useState<Set<string>>(
		new Set(),
	);
	const [applyFeedback, setApplyFeedback] = useState<
		"idle" | "saving" | "done" | "error"
	>("idle");

	useEffect(() => {
		if (settings) {
			setDraft({
				include: [...settings.include_keywords],
				exclude: [...settings.exclude_keywords],
				location: [...settings.location_keywords],
				workTypes: [...(settings.work_types ?? [])],
			});
		}
	}, [settings]);

	useEffect(() => {
		for (const key of showInputs) {
			const ref = inputRefs[key as keyof typeof inputRefs];
			if (ref.current) {
				ref.current.focus();
				break;
			}
		}
	}, [showInputs]);

	const stageAdd = (
		field: "include" | "exclude" | "location",
		rawValue: string,
	) => {
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

	const handleChipClick = (
		kw: string,
		field: "include" | "exclude" | "location",
	) => {
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

	if (isLoading) {
		return (
			<aside className="w-72 shrink-0 border-r border-[#1a2a1a] p-4">
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

	const display = draft ?? {
		include: [] as string[],
		exclude: [] as string[],
		location: [] as string[],
		workTypes: [] as string[],
	};

	return (
		<aside className="flex w-72 shrink-0 flex-col border-r border-[#1a2a1a]">
			<div className="flex-1 overflow-y-auto p-4">
				<div className="mb-5 flex items-center justify-between">
					<h3 className="text-xs font-semibold uppercase tracking-wider text-[#6a7a6a]">
						Preferences
					</h3>
					<button
						onClick={handleApply}
						disabled={!hasChanges || applyFeedback === "saving"}
						className={`rounded-lg px-3 py-1.5 text-[11px] font-semibold transition ${
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
						<label className="mb-2 block text-[10px] font-medium uppercase tracking-[0.05em] text-[#4a5a4a]">
							Work Type
						</label>
						<div className="flex flex-wrap gap-1.5">
							<button
								type="button"
								onClick={() =>
									setDraft((prev) => (prev ? { ...prev, workTypes: [] } : prev))
								}
								className={`rounded-full border px-2.5 py-1 text-[11px] font-medium transition-all duration-150 ${
									display.workTypes.length === 0
										? "bg-[rgba(125,186,122,0.12)] text-[#7dba7a] border-[rgba(125,186,122,0.2)]"
										: "bg-transparent text-[#6a7a6a] border-[rgba(255,255,255,0.06)] hover:text-[#8a9a8a] hover:border-[rgba(255,255,255,0.12)]"
								}`}
							>
								Any
							</button>
							{["internship", "contract", "part-time", "full-time"].map((t) => {
								const selected = display.workTypes.includes(t);
								return (
									<button
										key={t}
										type="button"
										onClick={() =>
											setDraft((prev) => {
												if (!prev) return prev;
												return {
													...prev,
													workTypes: selected
														? prev.workTypes.filter((x) => x !== t)
														: [...prev.workTypes, t],
												};
											})
										}
										className={`rounded-full border px-2.5 py-1 text-[11px] font-medium transition-all duration-150 ${
											selected
												? "bg-[rgba(125,186,122,0.12)] text-[#7dba7a] border-[rgba(125,186,122,0.2)]"
												: "bg-transparent text-[#6a7a6a] border-[rgba(255,255,255,0.06)] hover:text-[#8a9a8a] hover:border-[rgba(255,255,255,0.12)]"
										}`}
									>
										{t === "full-time"
											? "Full-time"
											: t.charAt(0).toUpperCase() + t.slice(1)}
									</button>
								);
							})}
						</div>
					</div>
				</div>
			</div>
			<ScanStatusBar />
		</aside>
	);
}

// ---- Keyword Section ----

interface KeywordSectionProps {
	label: string;
	field: "include" | "exclude" | "location";
	keywords: string[];
	showInput: boolean;
	onAddInput: () => void;
	onCloseInput: () => void;
	inputRef: React.RefObject<HTMLInputElement | null>;
	highlight: "green" | "red" | "amber";
	onStageAdd: (
		field: "include" | "exclude" | "location",
		value: string,
	) => void;
	onChipClick: (kw: string, field: "include" | "exclude" | "location") => void;
	markedForDelete: Set<string>;
}

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
}: KeywordSectionProps) {
	const c = chipColors[highlight];
	return (
		<div>
			<label className="mb-2 block text-[10px] font-medium uppercase tracking-[0.05em] text-[#4a5a4a]">
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
					<span className="text-[10px] text-[#3a4a3a]">None set</span>
				) : (
					keywords.map((kw) => (
						<button
							key={kw}
							type="button"
							onClick={() => onChipClick(kw, field)}
							className={`rounded-full border px-2.5 py-1 text-[11px] font-medium transition-all duration-150 ${markedForDelete.has(kw) ? `${c.marked} line-through` : `${c.base} ${c.hover}`}`}
						>
							{kw}
						</button>
					))
				)}
				{!showInput && (
					<button
						type="button"
						onClick={onAddInput}
						className="inline-flex items-center gap-0.5 rounded-full border border-dashed border-[rgba(255,255,255,0.08)] px-2.5 py-1 text-[11px] text-[#5a6a5a] transition hover:border-[rgba(255,255,255,0.18)] hover:text-[#8a9a8a]"
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

const ChipInput = forwardRef<
	HTMLInputElement,
	{
		onSubmit: (value: string) => void;
		onCancel: () => void;
		placeholder: string;
	}
>(({ onSubmit, onCancel, placeholder }, ref) => {
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
});
