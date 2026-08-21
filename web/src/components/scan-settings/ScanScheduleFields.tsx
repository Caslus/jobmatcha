import { ToggleSwitch } from "@/components/ui/ToggleSwitch";

interface ScanSettings {
	scan_enabled: boolean;
	scan_cron_expr: string;
	scan_timezone: string;
}

interface Props {
	value: ScanSettings;
	onChange: (data: ScanSettings) => void;
	showInfo?: boolean;
}

export function ScanScheduleFields({
	value,
	onChange,
	showInfo = true,
}: Props) {
	const browserTz = Intl.DateTimeFormat().resolvedOptions().timeZone;

	return (
		<div className="space-y-4">
			{/* Toggle */}
			{/* biome-ignore lint/a11y/noLabelWithoutControl: ToggleSwitch button acts as checkbox control */}
			<label className="flex items-center justify-between cursor-pointer">
				<span className="text-xs font-medium text-[#e8e8e8]">
					Enable scheduled scanning
				</span>
				<ToggleSwitch
					checked={value.scan_enabled}
					onChange={() =>
						onChange({ ...value, scan_enabled: !value.scan_enabled })
					}
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
				<input
					id="cron-expr"
					type="text"
					value={value.scan_cron_expr}
					onChange={(e) =>
						onChange({ ...value, scan_cron_expr: e.target.value })
					}
					placeholder="0 6 * * *"
					className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-3 py-2 text-xs text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[#7dba7a]"
				/>
				<p className="mt-1 text-[10px] text-[#4a5a4a]">
					Format: minute hour day month weekday.{" "}
					<a
						href={`https://crontab.guru/#${encodeURIComponent(value.scan_cron_expr.replace(/\s+/g, "_"))}`}
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
				<input
					id="tz"
					type="text"
					value={value.scan_timezone}
					onChange={(e) =>
						onChange({ ...value, scan_timezone: e.target.value })
					}
					placeholder="UTC"
					className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-3 py-2 text-xs text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[#7dba7a]"
				/>
				<p className="mt-1 text-[10px] text-[#4a5a4a]">
					<a
						href="https://en.wikipedia.org/wiki/List_of_tz_database_time_zones"
						target="_blank"
						rel="noopener noreferrer"
						className="underline hover:text-[#6a7a6a]"
					>
						IANA timezone name.
					</a>{" "}
					Detected: <span className="text-[#6a7a6a]">{browserTz}</span>
				</p>
			</div>

			{/* Info */}
			{showInfo && !value.scan_enabled && (
				<p className="text-[10px] text-[#4a5a4a] leading-relaxed">
					When enabled, the scanner will run automatically on the schedule
					defined by your cron expression.
				</p>
			)}
		</div>
	);
}
