import { ScanScheduleFields } from "@/components/scan-settings/ScanScheduleFields";

interface ScanSettings {
	scan_enabled: boolean;
	scan_cron_expr: string;
	scan_timezone: string;
}

interface Props {
	value: ScanSettings;
	onChange: (data: ScanSettings) => void;
}

export function ScanSettingsStep({ value, onChange }: Props) {
	return <ScanScheduleFields value={value} onChange={onChange} showInfo />;
}
