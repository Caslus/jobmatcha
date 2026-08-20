import { Play, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { formatDate } from "#/lib/date.ts";
import { useSettings, useUpdateSettings } from "../../hooks/useApi";

interface Props {
	onClose: () => void;
	scanning: boolean;
	onStartScan: () => void;
}

function ResetBtn({ title, onClick }: { title: string; onClick: () => void }) {
	return (
		<button
			type="button"
			onClick={onClick}
			className="absolute right-2 top-1/2 -translate-y-1/2 text-[#4a5a4a] transition hover:text-[#6a7a6a]"
			title={title}
		>
			<X size={12} />
		</button>
	);
}

function ToggleSwitch({
	checked,
	onChange,
}: {
	checked: boolean;
	onChange: () => void;
}) {
	return (
		<button
			type="button"
			role="switch"
			aria-checked={checked}
			onClick={onChange}
			className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full transition-colors ${
				checked ? "bg-[#7dba7a]" : "bg-[#2a3a2a]"
			}`}
		>
			<span
				className={`inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform ${
					checked ? "translate-x-[18px]" : "translate-x-[3px]"
				}`}
			/>
		</button>
	);
}

export function ScanSettingsModal({ onClose, scanning, onStartScan }: Props) {
	const { data: settings } = useSettings();
	const updateSettings = useUpdateSettings();

	const browserTz = useRef(Intl.DateTimeFormat().resolvedOptions().timeZone);

	const [draftEnabled, setDraftEnabled] = useState(
		settings?.scan_enabled ?? false,
	);
	const [draftCron, setDraftCron] = useState(
		settings?.scan_cron_expr ?? "0 */6 * * *",
	);
	const [draftTimezone, setDraftTimezone] = useState(
		settings?.scan_timezone || browserTz.current,
	);
	const [saveError, setSaveError] = useState<string | null>(null);

	const savedCron = settings?.scan_cron_expr ?? "0 */6 * * *";
	const savedTimezone = settings?.scan_timezone || "";
	const savedEnabled = settings?.scan_enabled ?? false;

	const cronDirty = draftCron !== savedCron || draftTimezone !== savedTimezone;

	const canSave =
		draftCron.trim().length > 0 && (draftEnabled !== savedEnabled || cronDirty);

	const handleSave = () => {
		const payload: Record<string, unknown> = {
			scan_enabled: draftEnabled,
			scan_cron_expr: draftCron,
		};
		if (draftTimezone !== savedTimezone) {
			payload.scan_timezone = draftTimezone;
		}
		updateSettings.mutate(
			payload as Parameters<typeof updateSettings.mutate>[0],
			{
				onSuccess: () => onClose(),
				onError: (err) => {
					setSaveError(
						err instanceof Error ? err.message : "Failed to save settings",
					);
				},
			},
		);
	};

	const set =
		(fn: (v: string) => void) => (e: React.ChangeEvent<HTMLInputElement>) => {
			fn(e.target.value);
			setSaveError(null);
		};

	const dialogRef = useRef<HTMLDialogElement>(null);

	useEffect(() => {
		const el = dialogRef.current;
		if (!el) return;
		el.showModal();
		return () => el.close();
	}, []);

	const handleBackdropClick = (e: React.MouseEvent<HTMLDialogElement>) => {
		if (e.target === dialogRef.current) onClose();
	};

	const handleCancel = (e: React.SyntheticEvent<HTMLDialogElement>) => {
		e.preventDefault();
		onClose();
	};

	return (
		<dialog
			ref={dialogRef}
			onClick={handleBackdropClick}
			onCancel={handleCancel}
			onKeyDown={(e) => {
				if (e.key === "Escape") handleCancel(e);
			}}
			className="fixed inset-0 z-50 m-0 flex h-full w-full items-center justify-center bg-transparent open:flex"
		>
			<style>{`dialog::backdrop { background: rgba(0,0,0,0.5); }`}</style>

			<div className="w-full max-w-sm rounded-xl border border-[#1a2a1a] bg-[#0d120d] p-6 shadow-2xl">
				{/* Header */}
				<div className="mb-4 flex items-center justify-between">
					<h3 className="text-sm font-semibold text-[#e8e8e8]">
						Scan Settings
					</h3>
					<button
						type="button"
						onClick={onClose}
						className="text-[#4a5a4a] transition hover:text-[#6a7a6a]"
					>
						<X size={14} />
					</button>
				</div>

				<div className="space-y-4">
					{/* Run Manual Scan */}
					<button
						type="button"
						onClick={() => {
							onStartScan();
							onClose();
						}}
						disabled={scanning}
						className="flex w-full items-center justify-center gap-2 rounded-lg bg-gradient-to-r from-[#7dba7a] to-[#5a8f5a] px-3 py-2 text-sm font-semibold text-[#080908] transition hover:from-[#8dca8a] hover:to-[#6a9f6a] disabled:cursor-not-allowed disabled:opacity-50"
					>
						<Play size={14} className={scanning ? "hidden" : ""} />
						{scanning ? "Scanning..." : "Run Manual Scan"}
					</button>

					<hr className="border-[#1a2a1a]" />

					{/* Toggle */}
					<label
						htmlFor="toggle"
						className="flex items-center justify-between cursor-pointer"
					>
						<span className="text-xs font-medium text-[#e8e8e8]">
							Enable scheduled scanning
						</span>
						<ToggleSwitch
							checked={draftEnabled}
							onChange={() => {
								setSaveError(null);
								setDraftEnabled(!draftEnabled);
							}}
						/>
					</label>

					{/* Cron expression */}
					<div>
						<label
							htmlFor="cron-expr"
							className="mb-1.5 block text-xs font-medium text-[#6a7a6a]"
						>
							Cron expression
						</label>
						<div className="relative">
							<input
								id="cron-expr"
								type="text"
								value={draftCron}
								onChange={set(setDraftCron)}
								placeholder="0 */6 * * *"
								className="w-full rounded-lg border border-[#2a3a2a] bg-[#080b08] px-3 py-2 pr-8 text-xs text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[rgba(125,186,122,0.3)]"
							/>
							{draftCron !== savedCron && (
								<ResetBtn
									title="Reset to saved value"
									onClick={() => {
										setDraftCron(savedCron);
										setSaveError(null);
									}}
								/>
							)}
						</div>
						<p className="mt-1 text-[10px] text-[#4a5a4a]">
							Format: minute hour day month weekday.{" "}
							<a
								href={
									"https://crontab.guru/#" +
									encodeURIComponent(draftCron.replace(/\s+/g, "_"))
								}
								target="_blank"
								rel="noopener noreferrer"
								className="underline hover:text-[#6a7a6a]"
							>
								crontab.guru
							</a>
						</p>
					</div>

					{/* Timezone */}
					<div>
						<label
							htmlFor="tz"
							className="mb-1.5 block text-xs font-medium text-[#6a7a6a]"
						>
							Timezone
						</label>
						<div className="relative">
							<input
								id="tz"
								type="text"
								value={draftTimezone}
								onChange={set(setDraftTimezone)}
								placeholder="UTC"
								className="w-full rounded-lg border border-[#2a3a2a] bg-[#080b08] px-3 py-2 pr-8 text-xs text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[rgba(125,186,122,0.3)]"
							/>
							{draftTimezone !== savedTimezone && (
								<ResetBtn
									title="Reset to saved value"
									onClick={() => {
										setDraftTimezone(savedTimezone || browserTz.current);
										setSaveError(null);
									}}
								/>
							)}
						</div>
						<p className="mt-1 text-[10px] text-[#4a5a4a]">
							<a
								href="https://en.wikipedia.org/wiki/List_of_tz_database_time_zones"
								target="_blank"
								rel="noopener noreferrer"
								className="underline hover:text-[#6a7a6a]"
							>
								IANA timezone name.
							</a>{" "}
							Detected:{" "}
							<span className="text-[#6a7a6a]">{browserTz.current}</span>
						</p>
					</div>

					{/* Next run preview */}
					{draftEnabled && (
						<div
							className={`rounded-lg border px-3 py-2 ${
								cronDirty
									? "border-[#1a2a1a] bg-[#080b08] opacity-50"
									: settings?.next_scan_at
										? "border-[#1a2a1a] bg-[#080b08]"
										: "hidden"
							}`}
						>
							<p className="text-[10px] uppercase tracking-wider text-[#4a5a4a] mb-0.5">
								Next scan{" "}
								<span className="text-[#3a4a3a]">({draftTimezone})</span>
							</p>
							{cronDirty ? (
								<p className="text-xs text-[#4a5a4a] italic">
									Save to see updated schedule
								</p>
							) : settings?.next_scan_at ? (
								<p className="text-xs text-[#7dba7a]">
									{formatDate(settings.next_scan_at)}
								</p>
							) : null}
						</div>
					)}

					{/* Info text */}
					{!draftEnabled && (
						<p className="text-[10px] text-[#4a5a4a] leading-relaxed">
							When enabled, the scanner will run automatically on the schedule
							defined by your cron expression.
						</p>
					)}

					{/* Error */}
					{saveError && (
						<div className="rounded-lg border border-red-400/20 bg-red-400/5 px-3 py-2">
							<p className="text-xs text-red-400/90">{saveError}</p>
						</div>
					)}

					{/* Save / Cancel */}
					<div className="flex gap-2 pt-1">
						<button
							type="button"
							onClick={handleSave}
							disabled={!canSave || updateSettings.isPending}
							className="flex-1 rounded-lg bg-gradient-to-r from-[#7dba7a] to-[#5a8f5a] px-3 py-2 text-xs font-semibold text-[#080908] transition hover:from-[#8dca8a] hover:to-[#6a9f6a] disabled:cursor-not-allowed disabled:opacity-50"
						>
							{updateSettings.isPending ? "Saving..." : "Save"}
						</button>
						<button
							type="button"
							onClick={onClose}
							className="flex-1 rounded-lg border border-[#2a3a2a] px-3 py-2 text-xs font-medium text-[#6a7a6a] transition hover:bg-[#1a2a1a]"
						>
							Cancel
						</button>
					</div>
				</div>
			</div>
		</dialog>
	);
}
