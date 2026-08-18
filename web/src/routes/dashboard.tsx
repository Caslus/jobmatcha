import { createFileRoute, redirect, useNavigate } from "@tanstack/react-router";
import { Bookmark, Clock, LogOut } from "lucide-react";
import { forwardRef, useEffect, useRef, useState } from "react";
import {
	usePatchRole,
	useRole,
	useRoles,
	useSettings,
	useUpdateSettings,
} from "../hooks/useApi";
import { authApi } from "../lib/api";
import { useAuthStore } from "../stores/auth";

export const Route = createFileRoute("/dashboard")({
	beforeLoad: () => {
		const { authenticated } = useAuthStore.getState();
		if (!authenticated) {
			throw redirect({ to: "/" });
		}
	},
	component: DashboardPage,
});

function timeAgo(dateStr: string | null): string {
	if (!dateStr) return "";
	const diff = Date.now() - new Date(dateStr).getTime();
	const mins = Math.floor(diff / 60000);
	if (mins < 1) return "just now";
	if (mins < 60) return `${mins}m ago`;
	const hrs = Math.floor(mins / 60);
	if (hrs < 24) return `${hrs}h ago`;
	const days = Math.floor(hrs / 24);
	if (days < 30) return `${days}d ago`;
	return `${Math.floor(days / 7)}w ago`;
}

function scoreColor(score: number): string {
	if (score >= 3) return "text-green-400 bg-green-400/10 border-green-400/20";
	if (score >= 1) return "text-amber-400 bg-amber-400/10 border-amber-400/20";
	return "text-gray-500 bg-gray-500/10 border-gray-500/20";
}

function DashboardPage() {
	const navigate = useNavigate();
	const { check } = useAuthStore();
	const [selectedId, setSelectedId] = useState<number | null>(null);
	const [page, setPage] = useState(1);
	const patchRole = usePatchRole();

	const { data: detailData } = useRole(selectedId);
	const selectedRole = detailData ?? null;

	useEffect(() => {
		check().then(() => {
			const { setupComplete } = useAuthStore.getState();
			if (!setupComplete) {
				navigate({ to: "/onboarding" });
			}
		});
	}, []);

	const handleLogout = async () => {
		try {
			await authApi.logout();
		} catch {
			// Ignore errors — the token is still cleared
		}
		useAuthStore.getState().logout();
		window.location.href = "/";
	};

	const { data: rolesData, isLoading: rolesLoading } = useRoles(page);
	const roles = rolesData?.data ?? [];
	const total = rolesData?.pagination?.total ?? 0;
	const totalPages = Math.ceil(total / 25);

	const handleToggleInterested = () => {
		if (!selectedRole) return;
		patchRole.mutate({
			id: selectedRole.id,
			is_interested: !selectedRole.is_interested,
		});
	};

	const handleToggleHidden = () => {
		if (!selectedRole) return;
		patchRole.mutate({
			id: selectedRole.id,
			is_hidden: !selectedRole.is_hidden,
		});
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
					<span className="text-xs text-[#6a7a6a]">{total} openings</span>
					<button
						onClick={handleLogout}
						className="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs text-[#6a7a6a] transition hover:bg-[#1a2a1a] hover:text-[#e8e8e8]"
					>
						<LogOut size={14} />
						Logout
					</button>
				</div>
			</header>

			<div className="flex flex-1 overflow-hidden">
				<PreferencesPanel />

				{/* Center — Role list */}
				<main className="flex flex-1 flex-col overflow-hidden">
					<div className="flex items-center justify-between border-b border-[#1a2a1a] px-6 py-3">
						<div className="flex items-center gap-2">
							<h2 className="text-sm font-semibold">Roles</h2>
							<span className="rounded-full bg-[#1a2a1a] px-2 py-0.5 text-xs text-[#6a7a6a]">
								{total}
							</span>
						</div>
						<div className="flex items-center gap-2 text-xs text-[#6a7a6a]">
							<span>Role</span>
							<span className="w-16 text-right">Match</span>
						</div>
					</div>

					<div className="flex-1 overflow-y-auto">
						{rolesLoading ? (
							<div className="space-y-2 p-4">
								{[1, 2, 3, 4, 5].map((i) => (
									<div
										key={i}
										className="h-16 animate-pulse rounded-lg bg-[#0d120d]"
									/>
								))}
							</div>
						) : roles.length === 0 ? (
							<div className="flex h-full flex-col items-center justify-center text-[#4a5a4a]">
								<p className="text-sm">No roles found</p>
								<p className="mt-1 text-xs">Run a scan to find jobs</p>
							</div>
						) : (
							<div className="space-y-1 p-2">
								{roles.map((role) => (
									<button
										key={role.id}
										onClick={() => setSelectedId(role.id)}
										className={`w-full rounded-xl border px-4 py-3 text-left transition ${
											selectedId === role.id
												? "border-[#7dba7a]/40 bg-[#1a2a1a]"
												: "border-transparent bg-transparent hover:bg-[#0d120d]"
										}`}
									>
										<div className="flex items-start justify-between gap-3">
											<div className="min-w-0 flex-1">
												<p className="truncate text-sm font-medium text-[#e8e8e8]">
													{role.title}
												</p>
												<p className="mt-0.5 text-xs text-[#6a7a6a]">
													{role.company_name}
													{role.location && (
														<>
															<span className="mx-1">·</span>
															{role.location}
														</>
													)}
												</p>
												<p className="mt-1 flex items-center gap-1 text-[10px] text-[#4a5a4a]">
													<Clock size={10} />
													{timeAgo(role.posted_at)}
												</p>
											</div>
											<div className="flex flex-col items-end gap-1">
												<span
													className={`inline-flex min-w-[2rem] items-center justify-center rounded-md border px-2 py-0.5 text-xs font-semibold ${scoreColor(role.relevance_score)}`}
												>
													{role.relevance_score}
												</span>
												{role.is_interested && (
													<Bookmark
														size={12}
														className="text-[#7dba7a]"
														fill="#7dba7a"
													/>
												)}
											</div>
										</div>
									</button>
								))}
							</div>
						)}
					</div>

					{totalPages > 1 && (
						<div className="flex items-center justify-between border-t border-[#1a2a1a] px-6 py-3">
							<button
								onClick={() => setPage((p) => Math.max(1, p - 1))}
								disabled={page <= 1}
								className="rounded-lg px-3 py-1.5 text-xs text-[#6a7a6a] transition hover:bg-[#1a2a1a] disabled:opacity-30"
							>
								Previous
							</button>
							<span className="text-xs text-[#4a5a4a]">
								Page {page} of {totalPages}
							</span>
							<button
								onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
								disabled={page >= totalPages}
								className="rounded-lg px-3 py-1.5 text-xs text-[#6a7a6a] transition hover:bg-[#1a2a1a] disabled:opacity-30"
							>
								Next
							</button>
						</div>
					)}
				</main>

				{/* Right — Role detail panel */}
				<aside className="w-96 shrink-0 border-l border-[#1a2a1a] overflow-y-auto">
					{selectedRole ? (
						<div className="p-6">
							<button
								onClick={() => setSelectedId(null)}
								className="mb-4 text-xs text-[#6a7a6a] transition hover:text-[#e8e8e8]"
							>
								← Back to list
							</button>
							<div className="mb-6">
								<p className="text-xs text-[#6a7a6a]">
									{selectedRole.company_name}
								</p>
								<h3 className="mt-1 text-lg font-bold text-[#e8e8e8]">
									{selectedRole.title}
								</h3>
								<div className="mt-3 flex flex-wrap items-center gap-3 text-xs text-[#6a7a6a]">
									<span className="flex items-center gap-1">
										{selectedRole.location}
									</span>
									<span
										className={`inline-flex items-center gap-1 rounded-md border px-2 py-0.5 font-semibold ${scoreColor(selectedRole.relevance_score)}`}
									>
										✦ Match Score: {selectedRole.relevance_score}
									</span>
								</div>
							</div>

							<div className="space-y-4">
								<div className="flex gap-2">
									<button
										onClick={handleToggleInterested}
										disabled={patchRole.isPending}
										className={`flex-1 rounded-lg border px-4 py-2 text-xs font-medium transition ${
											selectedRole.is_interested
												? "border-[#7dba7a] bg-[#7dba7a]/10 text-[#7dba7a]"
												: "border-[#2a3a2a] text-[#6a7a6a] hover:border-[#7dba7a]/50 hover:text-[#7dba7a]"
										}`}
									>
										{patchRole.isPending
											? "Updating..."
											: selectedRole.is_interested
												? "★ Interested"
												: "☆ Mark Interested"}
									</button>
									<button
										onClick={handleToggleHidden}
										disabled={patchRole.isPending}
										className="flex-1 rounded-lg border border-[#2a3a2a] px-4 py-2 text-xs font-medium text-[#6a7a6a] transition hover:border-red-400/50 hover:text-red-400"
									>
										{selectedRole.is_hidden ? "Unhide" : "Hide"}
									</button>
								</div>

								{selectedRole.url && (
									<div className="rounded-xl border border-[#1a2a1a] bg-[#0d120d] p-4">
										<p className="mb-2 text-xs font-medium text-[#6a7a6a]">
											Job URL
										</p>
										<a
											href={selectedRole.url}
											target="_blank"
											rel="noopener noreferrer"
											className="break-all text-xs text-[#7dba7a] transition hover:underline"
										>
											{selectedRole.url}
										</a>
									</div>
								)}

								<div className="rounded-xl border border-[#1a2a1a] bg-[#0d120d] p-4">
									<p className="mb-2 text-xs font-medium text-[#6a7a6a]">
										Description
									</p>
									<p className="text-xs leading-relaxed text-[#9a9a9a]">
										{selectedRole.description || "No description loaded"}
									</p>
								</div>
							</div>
						</div>
					) : (
						<div className="flex h-full items-center justify-center text-[#4a5a4a]">
							<p className="text-xs">Select a role to view details</p>
						</div>
					)}
				</aside>
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

	// Draft state — local copy of keywords for staging
	const [draft, setDraft] = useState<{
		include: string[];
		exclude: string[];
		location: string[];
		workTypes: string[];
	} | null>(null);

	// Track chips marked for deletion
	const [markedForDelete, setMarkedForDelete] = useState<Set<string>>(
		new Set(),
	);
	const [applyFeedback, setApplyFeedback] = useState<
		"idle" | "saving" | "done" | "error"
	>("idle");

	// Initialize draft from server settings
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

	// Focus input when a new one opens
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
		// Newly added items (staged but not yet applied) — remove immediately, no apply needed
		const key = `${field}_keywords` as
			| "include_keywords"
			| "exclude_keywords"
			| "location_keywords";
		const isStaged = settings && !settings[key].includes(kw);
		if (isStaged) {
			setDraft((prev) => {
				if (!prev) return prev;
				return { ...prev, [field]: prev[field].filter((k) => k !== kw) };
			});
		} else {
			// Existing server item — toggle deletion mark for batch Apply
			toggleDelete(kw);
		}
	};

	const handleApply = () => {
		if (!draft) return;
		setApplyFeedback("saving");
		const finalInclude = draft.include.filter((k) => !markedForDelete.has(k));
		const finalExclude = draft.exclude.filter((k) => !markedForDelete.has(k));
		const finalLocation = draft.location.filter((k) => !markedForDelete.has(k));

		updateSettings.mutate(
			{
				include_keywords: finalInclude,
				exclude_keywords: finalExclude,
				location_keywords: finalLocation,
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

					{/* Apply button — right next to the title */}
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
					{/* Include */}
					<div>
						<label className="mb-2 block text-[10px] font-medium uppercase tracking-[0.05em] text-[#4a5a4a]">
							Include
						</label>
						{showInputs.has("include") ? (
							<ChipInput
								ref={inputRefs.include}
								onSubmit={(v) => {
									stageAdd("include", v);
									setShowInputs((prev) => {
										const n = new Set(prev);
										n.delete("include");
										return n;
									});
								}}
								onCancel={() =>
									setShowInputs((prev) => {
										const n = new Set(prev);
										n.delete("include");
										return n;
									})
								}
								placeholder="engineer, go, backend..."
							/>
						) : null}
						<div className="flex flex-wrap gap-1.5">
							{display.include.length === 0 ? (
								<span className="text-[10px] text-[#3a4a3a]">None set</span>
							) : (
								display.include.map((kw: string) => (
									<Chip
										key={kw}
										label={kw}
										highlight="green"
										marked={markedForDelete.has(kw)}
										onClick={() => handleChipClick(kw, "include")}
									/>
								))
							)}
							{!showInputs.has("include") && (
								<AddChipButton
									onClick={() =>
										setShowInputs((prev) => {
											const n = new Set(prev);
											n.add("include");
											return n;
										})
									}
								/>
							)}
						</div>
					</div>

					{/* Exclude */}
					<div>
						<label className="mb-2 block text-[10px] font-medium uppercase tracking-[0.05em] text-[#4a5a4a]">
							Exclude
						</label>
						{showInputs.has("exclude") ? (
							<ChipInput
								ref={inputRefs.exclude}
								onSubmit={(v) => {
									stageAdd("exclude", v);
									setShowInputs((prev) => {
										const n = new Set(prev);
										n.delete("exclude");
										return n;
									});
								}}
								onCancel={() =>
									setShowInputs((prev) => {
										const n = new Set(prev);
										n.delete("exclude");
										return n;
									})
								}
								placeholder="manager, sales..."
							/>
						) : null}
						<div className="flex flex-wrap gap-1.5">
							{display.exclude.length === 0 ? (
								<span className="text-[10px] text-[#3a4a3a]">None set</span>
							) : (
								display.exclude.map((kw: string) => (
									<Chip
										key={kw}
										label={kw}
										highlight="red"
										marked={markedForDelete.has(kw)}
										onClick={() => handleChipClick(kw, "exclude")}
									/>
								))
							)}
							{!showInputs.has("exclude") && (
								<AddChipButton
									onClick={() =>
										setShowInputs((prev) => {
											const n = new Set(prev);
											n.add("exclude");
											return n;
										})
									}
								/>
							)}
						</div>
					</div>

					{/* Location */}
					<div>
						<label className="mb-2 block text-[10px] font-medium uppercase tracking-[0.05em] text-[#4a5a4a]">
							Location
						</label>
						{showInputs.has("location") ? (
							<ChipInput
								ref={inputRefs.location}
								onSubmit={(v) => {
									stageAdd("location", v);
									setShowInputs((prev) => {
										const n = new Set(prev);
										n.delete("location");
										return n;
									});
								}}
								onCancel={() =>
									setShowInputs((prev) => {
										const n = new Set(prev);
										n.delete("location");
										return n;
									})
								}
								placeholder="Tokyo, Remote..."
							/>
						) : null}
						<div className="flex flex-wrap gap-1.5">
							{display.location.length === 0 ? (
								<span className="text-[10px] text-[#3a4a3a]">None set</span>
							) : (
								display.location.map((kw: string) => (
									<Chip
										key={kw}
										label={kw}
										highlight="amber"
										marked={markedForDelete.has(kw)}
										onClick={() => handleChipClick(kw, "location")}
									/>
								))
							)}
							{!showInputs.has("location") && (
								<AddChipButton
									onClick={() =>
										setShowInputs((prev) => {
											const n = new Set(prev);
											n.add("location");
											return n;
										})
									}
								/>
							)}
						</div>
					</div>

					{/* Work Type */}
					<div>
						<label className="mb-2 block text-[10px] font-medium uppercase tracking-[0.05em] text-[#4a5a4a]">
							Work Type
						</label>
						<div className="flex flex-wrap gap-1.5">
							<button
								type="button"
								onClick={() =>
									setDraft((prev) => {
										if (!prev) return prev;
										return { ...prev, workTypes: [] };
									})
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
												const next = selected
													? prev.workTypes.filter((x) => x !== t)
													: [...prev.workTypes, t];
												return { ...prev, workTypes: next };
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
		</aside>
	);
}

// ---- Chip Components ----

interface ChipProps {
	label: string;
	highlight: "green" | "red" | "amber";
	marked: boolean;
	onClick: () => void;
}

const chipHighlight = {
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

function Chip({ label, highlight, marked, onClick }: ChipProps) {
	const c = chipHighlight[highlight];

	return (
		<button
			type="button"
			onClick={onClick}
			className={`rounded-full border px-2.5 py-1 text-[11px] font-medium transition-all duration-150 ${
				marked ? `${c.marked} line-through` : `${c.base} ${c.hover}`
			}`}
		>
			{label}
		</button>
	);
}

function AddChipButton({ onClick }: { onClick: () => void }) {
	return (
		<button
			type="button"
			onClick={onClick}
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

	const handleKeyDown = (e: React.KeyboardEvent) => {
		if (e.key === "Enter") {
			e.preventDefault();
			if (value.trim()) onSubmit(value.trim());
		} else if (e.key === "Escape") {
			onCancel();
		}
	};

	const handleBlur = () => {
		setTimeout(() => {
			if (value.trim()) onSubmit(value.trim());
			else onCancel();
		}, 150);
	};

	return (
		<input
			ref={ref}
			type="text"
			value={value}
			onChange={(e) => setValue(e.target.value)}
			onKeyDown={handleKeyDown}
			onBlur={handleBlur}
			placeholder={placeholder}
			className="mb-2 w-full rounded-md border border-[rgba(255,255,255,0.06)] bg-[rgba(0,0,0,0.25)] px-2.5 py-1.5 text-xs text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[rgba(125,186,122,0.3)]"
		/>
	);
});
