import { Settings } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import type { ScanJobResponse } from "#/types/api.gen.ts";
import { scanApi } from "../../lib/api";

export function ScanStatusBar() {
	const [scanJob, setScanJob] = useState<ScanJobResponse | null>(null);
	const [scanning, setScanning] = useState(false);
	const [showScheduled, setShowScheduled] = useState(false);
	const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

	// Load latest scan on mount — resume polling if still in progress
	useEffect(() => {
		scanApi
			.getLatest()
			.then((job) => {
				if (job) {
					setScanJob(job);
					if (job.status === "pending" || job.status === "running") {
						setScanning(true);
						startPolling(job.id);
					}
				}
			})
			.catch(() => {});
	}, []);

	const startPolling = (scanId: number) => {
		if (intervalRef.current) clearInterval(intervalRef.current);
		intervalRef.current = setInterval(async () => {
			try {
				const job = await scanApi.get(scanId);
				setScanJob(job);
				if (job.status === "completed" || job.status === "failed") {
					setScanning(false);
					if (intervalRef.current) clearInterval(intervalRef.current);
				}
			} catch {
				setScanning(false);
				if (intervalRef.current) clearInterval(intervalRef.current);
			}
		}, 2000);
	};

	const startScan = async () => {
		try {
			const resp = await scanApi.start();
			console.log("Scan started:", resp);
			setScanning(true);
			startPolling(resp.id);
		} catch {
			setScanning(false);
		}
	};

	// Cleanup interval on unmount
	useEffect(() => {
		return () => {
			if (intervalRef.current) clearInterval(intervalRef.current);
		};
	}, []);

	const lastScan = scanJob;
	const lastScanDate = lastScan?.completed_at
		? new Date(lastScan.completed_at).toLocaleDateString("en-CA") +
			", " +
			new Date(lastScan.completed_at).toLocaleTimeString("en-US", {
				hour: "2-digit",
				minute: "2-digit",
				second: "2-digit",
				hour12: false,
			})
		: null;

	return (
		<>
			<div className="border-t border-[#1a2a1a] p-3 space-y-2">
				<div className="flex items-center justify-between">
					<span className="text-sm font-medium uppercase tracking-wider text-[#4a5a4a]">
						Last Scan
					</span>
					<button
						onClick={() => setShowScheduled(true)}
						className="text-[#4a5a4a] transition hover:text-[#6a7a6a]"
						title="Scheduled scan settings"
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
								{lastScanDate && (
									<p className="text-[#4a5a4a]">{lastScanDate}</p>
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

				{/* Progress bar */}
				{scanning && scanJob && scanJob.total_companies > 0 && (
					<div>
						<div className="h-1.5 rounded-full bg-[#1a2a1a] overflow-hidden">
							<div
								className="h-full rounded-full bg-gradient-to-r from-[#7dba7a] to-[#5a8f5a] transition-all duration-500"
								style={{
									width: `${Math.round((scanJob.completed_companies / scanJob.total_companies) * 100)}%`,
								}}
							/>
						</div>
						<p className="mt-1 text-xs text-[#4a5a4a] text-right">
							{scanJob.completed_companies}/{scanJob.total_companies} companies
						</p>
					</div>
				)}

				<button
					onClick={startScan}
					disabled={scanning}
					className="w-full rounded-lg bg-gradient-to-r from-[#7dba7a] to-[#5a8f5a] px-3 py-2 text-sm font-semibold text-[#080908] transition hover:from-[#8dca8a] hover:to-[#6a9f6a] disabled:cursor-not-allowed disabled:opacity-50"
				>
					{scanning ? "Scanning..." : "New scan"}
				</button>
			</div>

			{/* Scheduled Scan Modal */}
			{showScheduled && (
				<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
					<div className="w-full max-w-sm rounded-xl border border-[#1a2a1a] bg-[#0d120d] p-6 shadow-2xl">
						<div className="mb-4 flex items-center justify-between">
							<h3 className="text-sm font-semibold text-[#e8e8e8]">
								Scheduled Scan Settings
							</h3>
							<button
								onClick={() => setShowScheduled(false)}
								className="text-[#4a5a4a] transition hover:text-[#6a7a6a]"
							>
								<svg width="14" height="14" viewBox="0 0 14 14" fill="none">
									<path
										d="M3 3l8 8M11 3l-8 8"
										stroke="currentColor"
										strokeWidth="1.5"
										strokeLinecap="round"
									/>
								</svg>
							</button>
						</div>
						<p className="text-xs text-[#6a7a6a]">
							Scheduled scanning is not yet available. This will allow automatic
							daily/weekly scans.
						</p>
						<button
							onClick={() => setShowScheduled(false)}
							className="mt-4 w-full rounded-lg border border-[#2a3a2a] px-3 py-2 text-xs font-medium text-[#6a7a6a] transition hover:bg-[#1a2a1a]"
						>
							Close
						</button>
					</div>
				</div>
			)}
		</>
	);
}
