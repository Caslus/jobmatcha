import { createFileRoute, redirect } from "@tanstack/react-router";
import {
	ArrowDown,
	ArrowUp,
	ArrowUpDown,
	CircleAlert,
	CircleCheck,
	CircleHelp,
	Clock3,
	ShieldOff,
	Sparkles,
} from "lucide-react";
import { useMemo, useState } from "react";
import Header from "#/components/Header.tsx";
import { Tooltip } from "#/components/Tooltip.tsx";
import { Switch } from "#/components/ui/switch.tsx";
import type { CompanyListItem } from "@/types/api.gen";
import {
	authStatusQueryOptions,
	useCompanies,
	useUpdateCompaniesActiveBulk,
	useUpdateCompanyActive,
} from "../hooks/useApi";

export const Route = createFileRoute("/companies")({
	beforeLoad: async ({ context }) => {
		const status = await context.queryClient.ensureQueryData(
			authStatusQueryOptions(),
		);
		if (!status.authenticated) throw redirect({ to: "/" });
		if (!status.setup_complete) throw redirect({ to: "/onboarding" });
	},
	component: CompaniesPage,
});

type SortKey =
	| "name"
	| "location"
	| "ats_type"
	| "active"
	| "role_count"
	| "adapter_status"
	| "freshness_status"
	| "last_scan_attempt_at"
	| "last_new_role_discovery_at";
const columns: Array<[SortKey, string]> = [
	["name", "Company"],
	["location", "Location"],
	["ats_type", "Adapter"],
	["role_count", "Jobs"],
	["adapter_status", "Adapter status"],
	["freshness_status", "Freshness"],
	["last_scan_attempt_at", "Latest scan"],
	["last_new_role_discovery_at", "Latest discovery"],
	["active", "Enabled"],
];

function labelStatus(value: string) {
	return value.replaceAll("_", " ");
}
function date(value: unknown) {
	return value ? new Date(value as string).toLocaleString() : "—";
}

export function CompaniesPage() {
	const { data, isLoading, error } = useCompanies();
	const single = useUpdateCompanyActive();
	const bulk = useUpdateCompaniesActiveBulk();
	const [selected, setSelected] = useState<Set<number>>(new Set());
	const [sort, setSort] = useState<{ key: SortKey; ascending: boolean }>({
		key: "role_count",
		ascending: false,
	});
	const companies = useMemo(
		() =>
			[...(data?.data ?? [])].sort((a, b) => {
				const left = a[sort.key] ?? "";
				const right = b[sort.key] ?? "";
				const result =
					typeof left === "boolean" && typeof right === "boolean"
						? Number(left) - Number(right)
						: typeof left === "number" && typeof right === "number"
							? left - right
							: String(left).localeCompare(String(right));
				return sort.ascending ? result : -result;
			}),
		[data, sort],
	);
	const toggleSort = (key: SortKey) =>
		setSort((current) =>
			current.key === key
				? { key, ascending: !current.ascending }
				: { key, ascending: true },
		);
	const busy = single.isPending || bulk.isPending;
	const select = (id: number) =>
		setSelected((current) => {
			const next = new Set(current);
			next.has(id) ? next.delete(id) : next.add(id);
			return next;
		});
	const applyBulk = (active: boolean) =>
		bulk.mutate(
			{ companyIDs: [...selected], active },
			{ onSuccess: () => setSelected(new Set()) },
		);
	return (
		<div className="min-h-screen text-[#e8e8e8]">
			<Header />
			<main className="p-6">
				<div className="mb-5 flex items-center justify-between gap-4">
					<div>
						<h1 className="text-xl font-semibold">Companies</h1>
						<p className="mt-1 text-sm text-[#8b9b8b]">
							Manage the job sources included in future scans.
						</p>
					</div>
					{selected.size > 0 && (
						<div className="flex gap-2">
							<button
								type="button"
								disabled={busy}
								onClick={() => applyBulk(true)}
								className="rounded bg-[#315a38] px-3 py-2 text-sm disabled:opacity-50"
							>
								Enable {selected.size}
							</button>
							<button
								type="button"
								disabled={busy}
								onClick={() => applyBulk(false)}
								className="rounded bg-[#3f3030] px-3 py-2 text-sm disabled:opacity-50"
							>
								Disable {selected.size}
							</button>
						</div>
					)}
				</div>
				{error && (
					<p className="text-sm text-red-300">Could not load companies.</p>
				)}
				{isLoading ? (
					<p className="text-sm text-[#8b9b8b]">Loading companies…</p>
				) : (
					<div className="overflow-x-auto rounded-lg border border-[#1a2a1a]">
						<table className="w-full text-left text-sm">
							<thead className="bg-[#151d15] text-xs text-[#a9b5a9]">
								<tr>
									<th className="p-3">
										<input
											aria-label="Select all companies"
											type="checkbox"
											checked={
												companies.length > 0 &&
												selected.size === companies.length
											}
											onChange={() =>
												setSelected(
													selected.size === companies.length
														? new Set()
														: new Set(companies.map((company) => company.id)),
												)
											}
										/>
									</th>
									{columns.map(([key, title]) => (
										<th key={key} className="whitespace-nowrap p-3">
											<button
												type="button"
												onClick={() => toggleSort(key)}
												className="flex items-center gap-1"
											>
												{title}
												{sort.key === key ? (
													sort.ascending ? (
														<ArrowUp size={12} className="text-[#7dba7a]" />
													) : (
														<ArrowDown size={12} className="text-[#7dba7a]" />
													)
												) : (
													<ArrowUpDown size={12} />
												)}
											</button>
										</th>
									))}
								</tr>
							</thead>
							<tbody>
								{companies.map((company) => (
									<CompanyRow
										key={company.id}
										company={company}
										selected={selected.has(company.id)}
										onSelect={select}
										onToggle={(active) =>
											single.mutate({ id: company.id, active })
										}
										busy={busy}
									/>
								))}
							</tbody>
						</table>
					</div>
				)}
			</main>
		</div>
	);
}

function CompanyRow({
	company,
	selected,
	onSelect,
	onToggle,
	busy,
}: {
	company: CompanyListItem;
	selected: boolean;
	onSelect: (id: number) => void;
	onToggle: (active: boolean) => void;
	busy: boolean;
}) {
	return (
		<tr className="border-t border-[#1a2a1a] align-top hover:bg-[#121a12]">
			<td className="p-3">
				<input
					aria-label={`Select ${company.name}`}
					type="checkbox"
					checked={selected}
					onChange={() => onSelect(company.id)}
				/>
			</td>
			<td className="p-3 font-medium">{company.name}</td>
			<td className="p-3 text-[#a9b5a9]">{company.location || "—"}</td>
			<td className="p-3">{company.ats_type || "—"}</td>
			<td className="p-3 text-center tabular-nums">{company.role_count}</td>
			<td className="p-3">
				<StatusIndicator
					kind="adapter"
					status={company.adapter_status}
					failureDetail={company.last_scan_failure_detail}
				/>
			</td>
			<td className="p-3">
				<StatusIndicator kind="freshness" status={company.freshness_status} />
			</td>
			<td className="whitespace-nowrap p-3 text-xs text-[#a9b5a9]">
				{date(company.last_scan_attempt_at)}
			</td>
			<td className="whitespace-nowrap p-3 text-xs text-[#a9b5a9]">
				{date(company.last_new_role_discovery_at)}
			</td>
			<td className="p-3">
				<Switch
					aria-label={`${company.name} enabled`}
					checked={company.active}
					disabled={busy}
					onCheckedChange={onToggle}
				/>
			</td>
		</tr>
	);
}

function StatusIndicator({
	kind,
	status,
	failureDetail,
}: {
	kind: "adapter" | "freshness";
	status: string;
	failureDetail?: string | null;
}) {
	const detail =
		kind === "adapter"
			? ({
					healthy: "Latest scan completed successfully",
					failing: `Latest scan failed${failureDetail ? `: ${failureDetail}` : ""}`,
					unsupported: "No scanner is available for this adapter",
					unknown: "No scan attempt has been recorded yet",
				}[status] ?? labelStatus(status))
			: ({
					fresh: "New roles were found within the last 30 days",
					stale: "No new roles found in 30 days",
					no_activity_yet: "No roles have been discovered yet",
					not_applicable: "Freshness is not tracked for this company",
				}[status] ?? labelStatus(status));
	const Icon =
		kind === "adapter"
			? ({
					healthy: CircleCheck,
					failing: CircleAlert,
					unsupported: ShieldOff,
					unknown: CircleHelp,
				}[status] ?? CircleHelp)
			: ({
					fresh: Sparkles,
					stale: CircleAlert,
					no_activity_yet: Clock3,
					not_applicable: ShieldOff,
				}[status] ?? CircleHelp);
	const tone =
		status === "failing"
			? "text-red-300"
			: status === "stale" || status === "unsupported"
				? "text-amber-300"
				: status === "healthy" || status === "fresh"
					? "text-emerald-300"
					: "text-[#a9b5a9]";
	return (
		<Tooltip content={detail}>
			<button
				type="button"
				className={`inline-flex ${tone}`}
				aria-label={detail}
			>
				<Icon size={17} aria-hidden="true" />
			</button>
		</Tooltip>
	);
}
