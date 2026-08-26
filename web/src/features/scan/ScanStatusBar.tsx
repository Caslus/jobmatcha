import { Clock, Settings } from "lucide-react";
import { useState } from "react";
import { formatDate } from "#/lib/date.ts";
import {
	useLatestScan,
	useScan,
	useSettings,
	useStartScan,
} from "../../hooks/useApi";
import { ScanSettingsModal } from "./ScanSettingsModal";

export function ScanStatusBar() {
	const [showModal, setShowModal] = useState(false);
	const { data: settings } = useSettings();
	const latestScan = useLatestScan();
	const startedScan = useStartScan();
	const scan = useScan(startedScan.data?.id ?? latestScan.data?.id ?? null);
	const lastScan = scan.data ?? latestScan.data;
	const scanning =
		lastScan?.status === "pending" || lastScan?.status === "running";
	const startScan = () => startedScan.mutate();
	const nextScanTime = settings?.next_scan_at
		? formatDate(settings.next_scan_at)
		: null;

	return (
		<>
			<div className="border-t border-[#1a2a1a] p-3 space-y-2">
				<div className="flex items-center justify-between">
					<span className="text-sm font-medium uppercase tracking-wider text-[#4a5a4a]">
						Last Scan
					</span>
					<button
						type="button"
						onClick={() => setShowModal(true)}
						className="text-[#4a5a4a] transition hover:text-[#6a7a6a]"
						title="Scan settings"
					>
						<Settings size={16} />
					</button>
				</div>

				{lastScan ? (
					<div className="text-xs text-[#6a7a6a] leading-relaxed">
						{lastScan.status === "completed" ? (
							<>
								<p className="text-[#9a9a9a] font-medium">
									{lastScan.total_new_roles} new roles
								</p>
								<p>
									{lastScan.total_roles} total
									{lastScan.total_companies > 0 &&
										` · ${lastScan.total_companies} companies`}
									{lastScan.duration_ms > 0 &&
										` · ${(lastScan.duration_ms / 1000).toFixed(1)}s`}
								</p>
								{lastScan.completed_at && (
									<p className="text-[#4a5a4a]">
										{formatDate(lastScan.completed_at)}
									</p>
								)}
							</>
						) : lastScan.status === "failed" ? (
							<>
								<p className="text-red-400 font-medium">Scan failed</p>
								{lastScan.error && (
									<p className="text-[#4a5a4a] truncate">{lastScan.error}</p>
								)}
							</>
						) : lastScan.status === "pending" ||
							lastScan.status === "running" ? (
							<p className="text-amber-400">Scanning...</p>
						) : null}
					</div>
				) : (
					<p className="text-xs text-[#3a4a3a]">Never scanned</p>
				)}

				{/* Scheduled indicator */}
				{settings?.scan_enabled && (
					<div className="flex items-center gap-1.5 text-xs text-[#7dba7a]/70">
						<Clock size={11} />
						<span>
							{nextScanTime
								? `Next scan: ${nextScanTime}`
								: `Scheduled (${settings.scan_timezone})`}
						</span>
					</div>
				)}

				{/* Progress bar */}
				{scanning && lastScan && lastScan.total_companies > 0 && (
					<div>
						<div className="h-1.5 rounded-full bg-[#1a2a1a] overflow-hidden">
							<div
								className="h-full rounded-full bg-linear-to-r from-[#7dba7a] to-[#5a8f5a] transition-all duration-500"
								style={{
									width: `${Math.round((lastScan.completed_companies / lastScan.total_companies) * 100)}%`,
								}}
							/>
						</div>
						<p className="mt-1 text-xs text-[#4a5a4a] text-right">
							{lastScan.completed_companies}/{lastScan.total_companies}{" "}
							companies
						</p>
					</div>
				)}
			</div>

			{showModal && (
				<ScanSettingsModal
					onClose={() => setShowModal(false)}
					scanning={scanning}
					onStartScan={startScan}
				/>
			)}
		</>
	);
}
