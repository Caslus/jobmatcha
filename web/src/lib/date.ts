export function formatDate(iso: string): string {
	const d = new Date(iso);
	return (
		d.toLocaleDateString("en-CA") +
		" · " +
		d.toLocaleTimeString("en-US", {
			hour: "2-digit",
			minute: "2-digit",
			second: "2-digit",
			hour12: false,
		})
	);
}

export function formatShortDate(iso: string): string {
	return new Date(iso).toLocaleString("en-US", {
		hour: "2-digit",
		minute: "2-digit",
		hour12: false,
		month: "long",
		day: "numeric",
	});
}
