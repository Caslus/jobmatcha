import { useRef, useState } from "react";
import { formatDate } from "#/lib/date.ts";
import { ScanScheduleFields } from "@/components/scan-settings/ScanScheduleFields";
import { Modal } from "@/components/ui/modal";
import { useSettings, useUpdateSettings } from "../../hooks/useApi";

interface Props {
	onClose: () => void;
	scanning: boolean;
	onStartScan: () => void;
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

	return (
		<Modal
			open
			onClose={onClose}
			title="Scan Settings"
			panelClassName="max-w-sm"
		>
			<div className="space-y-4 p-6">
				{/* Run Manual Scan */}
				<button
					type="button"
					onClick={() => {
						onStartScan();
						onClose();
					}}
					disabled={scanning}
					className="flex w-full items-center justify-center gap-2 rounded-lg bg-linear-to-r from-[#7dba7a] to-[#5a8f5a] px-3 py-2 text-sm font-semibold text-[#080908] transition hover:from-[#8dca8a] hover:to-[#6a9f6a] disabled:cursor-not-allowed disabled:opacity-50"
				>
					<svg
						width="14"
						height="14"
						viewBox="0 0 14 14"
						fill="none"
						aria-hidden="true"
						className={scanning ? "hidden" : ""}
					>
						<path d="M3 1l10 6-10 6V1z" fill="currentColor" />
					</svg>
					{scanning ? "Scanning..." : "Run Manual Scan"}
				</button>

				<hr className="border-[#1a2a1a]" />

				<ScanScheduleFields
					value={{
						scan_enabled: draftEnabled,
						scan_cron_expr: draftCron,
						scan_timezone: draftTimezone,
					}}
					onChange={(v) => {
						setDraftEnabled(v.scan_enabled);
						setDraftCron(v.scan_cron_expr);
						setDraftTimezone(v.scan_timezone);
						setSaveError(null);
					}}
					showInfo={false}
				/>

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
						className="flex-1 rounded-lg bg-linear-to-r from-[#7dba7a] to-[#5a8f5a] px-3 py-2 text-xs font-semibold text-[#080908] transition hover:from-[#8dca8a] hover:to-[#6a9f6a] disabled:cursor-not-allowed disabled:opacity-50"
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
		</Modal>
	);
}
