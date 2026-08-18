export function timeAgo(dateStr: string | null): string {
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

export function scoreColor(percent: number): string {
	if (percent >= 70)
		return "text-green-400 bg-green-400/10 border-green-400/20";
	if (percent >= 40)
		return "text-amber-400 bg-amber-400/10 border-amber-400/20";
	return "text-gray-500 bg-gray-500/10 border-gray-500/20";
}

export function matchLabel(percent: number): string {
	return percent > 0 ? `${percent}%` : "–";
}
