import { createFileRoute, redirect } from "@tanstack/react-router";
import {
	ArrowDown,
	ArrowUp,
	BriefcaseBusiness,
	Building2,
	Check,
	ChevronRight,
	CircleAlert,
	CircleHelp,
	Clock3,
	ExternalLink,
	Layers3,
	type LucideIcon,
	Pencil,
	Plus,
	Power,
	PowerOff,
	ShieldOff,
	Sparkles,
	Trash2,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import Header from "#/components/Header.tsx";
import { Tooltip } from "#/components/Tooltip.tsx";
import { Modal } from "#/components/ui/modal.tsx";
import { Switch } from "#/components/ui/switch.tsx";
import { timeAgo } from "#/lib/dashboard.ts";
import { formatDate } from "#/lib/date.ts";
import type {
	CareerBoardDiscoveryCandidate,
	CompanyListItem,
} from "@/types/api.gen";
import {
	authStatusQueryOptions,
	useCompanies,
	useCreateCareerBoard,
	useDeleteCareerBoard,
	useDeleteCompany,
	useDiscoverCareerBoards,
	useRegisterCareerBoards,
	useUpdateCareerBoardActive,
	useUpdateCareerBoardDetails,
	useUpdateCompaniesActiveBulk,
	useUpdateCompanyActive,
	useUpdateCompanyDetails,
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

function labelStatus(value: string) {
	return value.replaceAll("_", " ");
}
function relativeDate(value: string | null) {
	return value ? timeAgo(value) : "—";
}

type SortKey = "jobs" | "boards" | "name" | "activity";

const sortLabels: Record<SortKey, string> = {
	jobs: "Job count",
	boards: "Board count",
	name: "Name",
	activity: "Latest activity",
};

function activityTime(company: CompanyListItem) {
	return Math.max(
		company.last_scan_attempt_at
			? new Date(company.last_scan_attempt_at).getTime()
			: 0,
		company.last_new_role_discovery_at
			? new Date(company.last_new_role_discovery_at).getTime()
			: 0,
	);
}

export function CompaniesPage() {
	const { data, isLoading, error } = useCompanies();
	const single = useUpdateCompanyActive();
	const board = useUpdateCareerBoardActive();
	const updateCompany = useUpdateCompanyDetails();
	const deleteCompany = useDeleteCompany();
	const createBoard = useCreateCareerBoard();
	const updateBoard = useUpdateCareerBoardDetails();
	const deleteBoard = useDeleteCareerBoard();
	const bulk = useUpdateCompaniesActiveBulk();
	const discover = useDiscoverCareerBoards();
	const register = useRegisterCareerBoards();
	const [selected, setSelected] = useState<Set<number>>(new Set());
	const [boardCompany, setBoardCompany] = useState<CompanyListItem | null>(
		null,
	);
	const [discoveryOpen, setDiscoveryOpen] = useState(false);
	const [editingCompany, setEditingCompany] = useState<CompanyListItem | null>(
		null,
	);
	const [sortKey, setSortKey] = useState<SortKey>("jobs");
	const [sortDirection, setSortDirection] = useState<"asc" | "desc">("desc");
	const companies = useMemo(
		() =>
			[...(data?.data ?? [])].sort((a, b) => {
				const direction = sortDirection === "asc" ? 1 : -1;
				const comparison =
					sortKey === "jobs"
						? a.role_count - b.role_count
						: sortKey === "boards"
							? a.board_count - b.board_count
							: sortKey === "activity"
								? activityTime(a) - activityTime(b)
								: a.name.localeCompare(b.name);
				return comparison === 0
					? a.name.localeCompare(b.name)
					: comparison * direction;
			}),
		[data, sortDirection, sortKey],
	);
	const busy =
		single.isPending ||
		bulk.isPending ||
		board.isPending ||
		updateCompany.isPending ||
		deleteCompany.isPending ||
		createBoard.isPending ||
		updateBoard.isPending ||
		deleteBoard.isPending;
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
			<main className="mx-auto w-full max-w-5xl px-5 py-7 sm:px-7">
				<div className="mb-7 flex items-center justify-between gap-4">
					<div>
						<h1 className="text-xl font-semibold">Companies</h1>
						<p className="mt-1 text-sm text-[#8b9b8b]">
							Manage the job sources included in future scans.
						</p>
					</div>
					<div className="flex items-center gap-2">
						<button
							type="button"
							onClick={() => setDiscoveryOpen(true)}
							className="rounded-xl bg-[#1a2a1a] px-4 py-2 text-sm font-medium text-[#c8d5c8] transition hover:bg-[#253425] hover:text-[#e8eee8]"
						>
							Discover boards
						</button>
					</div>
				</div>
				{error && (
					<p className="text-sm text-red-300">Could not load companies.</p>
				)}
				{isLoading ? (
					<p className="text-sm text-[#8b9b8b]">Loading companies…</p>
				) : (
					<section aria-label="Registered companies">
						<div className="mb-3 flex flex-wrap items-center gap-2 px-1 text-xs text-[#718071]">
							<span className="mr-1">{companies.length} companies</span>
							<label className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1.5 transition hover:bg-[#1a2a1a] hover:text-[#c8d5c8]">
								<span className="sr-only">Sort companies by</span>
								<span>Sort:</span>
								<select
									value={sortKey}
									onChange={(event) =>
										setSortKey(event.target.value as SortKey)
									}
									className="bg-transparent font-medium text-[#c8d5c8] outline-none"
								>
									{Object.entries(sortLabels).map(([value, label]) => (
										<option key={value} value={value} className="bg-[#101710]">
											{label}
										</option>
									))}
								</select>
							</label>
							<button
								type="button"
								onClick={() =>
									setSortDirection((direction) =>
										direction === "asc" ? "desc" : "asc",
									)
								}
								aria-label={`Sort ${sortLabels[sortKey]} ${sortDirection === "asc" ? "ascending" : "descending"}; reverse order`}
								title="Reverse sort order"
								className="inline-flex size-7 items-center justify-center rounded-full transition hover:bg-[#1a2a1a] hover:text-[#c8d5c8]"
							>
								{sortDirection === "asc" ? (
									<ArrowUp size={14} />
								) : (
									<ArrowDown size={14} />
								)}
							</button>
							<button
								type="button"
								onClick={() =>
									setSelected(new Set(companies.map((company) => company.id)))
								}
								disabled={
									companies.length === 0 || selected.size === companies.length
								}
								className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1.5 transition hover:bg-[#1a2a1a] hover:text-[#c8d5c8] disabled:opacity-40"
							>
								<span className="flex size-4 items-center justify-center rounded-full border border-[#526052]">
									<Check size={10} />
								</span>
								Select all
							</button>
							<button
								type="button"
								onClick={() => setSelected(new Set())}
								disabled={selected.size === 0}
								className="rounded-full px-2.5 py-1.5 transition hover:bg-[#1a2a1a] hover:text-[#c8d5c8] disabled:opacity-40"
							>
								Clear selection
							</button>
							{selected.size > 0 && (
								<div className="ml-auto flex items-center gap-2">
									<span>{selected.size} selected</span>
									<button
										type="button"
										disabled={busy}
										onClick={() => applyBulk(true)}
										className="inline-flex items-center gap-1.5 rounded-full bg-[#315a38] px-3 py-1.5 font-medium text-[#e8eee8] transition hover:bg-[#3a6a42] disabled:opacity-50"
									>
										<Power size={13} />
										Enable
									</button>
									<button
										type="button"
										disabled={busy}
										onClick={() => applyBulk(false)}
										className="inline-flex items-center gap-1.5 rounded-full bg-[#3f3030] px-3 py-1.5 font-medium text-[#f0dada] transition hover:bg-[#553838] disabled:opacity-50"
									>
										<PowerOff size={13} />
										Disable
									</button>
								</div>
							)}
						</div>
						<div className="space-y-2.5">
							{companies.map((company) => (
								<CompanyRow
									key={company.id}
									company={company}
									selected={selected.has(company.id)}
									onSelect={select}
									onToggle={(active) =>
										single.mutate({ id: company.id, active })
									}
									onManageBoards={() => setBoardCompany(company)}
									onEdit={() => setEditingCompany(company)}
									busy={busy}
								/>
							))}
						</div>
					</section>
				)}
				<BoardModal
					company={boardCompany}
					busy={busy}
					onClose={() => setBoardCompany(null)}
					onToggle={(boardID, active) => {
						if (!boardCompany) return;
						const previousCompany = boardCompany;
						setBoardCompany({
							...previousCompany,
							career_boards: previousCompany.career_boards.map((candidate) =>
								candidate.id === boardID ? { ...candidate, active } : candidate,
							),
						});
						board.mutate(
							{ companyID: previousCompany.id, boardID, active },
							{
								onSuccess: (company) => setBoardCompany(company),
								onError: () => setBoardCompany(previousCompany),
							},
						);
					}}
					onCreate={(input) => {
						if (!boardCompany) return;
						createBoard.mutate(
							{ companyID: boardCompany.id, ...input },
							{ onSuccess: setBoardCompany },
						);
					}}
					onUpdate={(boardID, input) => {
						if (!boardCompany) return;
						updateBoard.mutate(
							{ companyID: boardCompany.id, boardID, ...input },
							{ onSuccess: setBoardCompany },
						);
					}}
					onDelete={(boardID) => {
						if (
							!boardCompany ||
							!window.confirm(
								"Delete this career board? Existing jobs will be kept as history.",
							)
						)
							return;
						deleteBoard.mutate(
							{ companyID: boardCompany.id, boardID },
							{ onSuccess: setBoardCompany },
						);
					}}
				/>
				<CompanyEditModal
					company={editingCompany}
					busy={busy}
					onClose={() => setEditingCompany(null)}
					onSave={(name) => {
						if (!editingCompany) return;
						updateCompany.mutate(
							{ id: editingCompany.id, name },
							{ onSuccess: () => setEditingCompany(null) },
						);
					}}
					onDelete={() => {
						if (
							!editingCompany ||
							!window.confirm(
								`Delete ${editingCompany.name}? Its job history will be kept, but hidden from the Jobs page.`,
							)
						)
							return;
						deleteCompany.mutate(
							{ id: editingCompany.id },
							{ onSuccess: () => setEditingCompany(null) },
						);
					}}
				/>
				<Modal
					open={discoveryOpen}
					onClose={() => setDiscoveryOpen(false)}
					title="Discover career boards"
					subtitle="Find public job sources, then add the ones you want under one company."
					icon={<Sparkles size={16} className="text-[#7dba7a]" />}
					panelClassName="max-w-4xl"
				>
					<DiscoveryPanel
						discover={discover}
						register={register}
						onRegistered={(company) => {
							setBoardCompany(company);
							setDiscoveryOpen(false);
						}}
					/>
				</Modal>
			</main>
		</div>
	);
}

function BoardModal({
	company,
	busy,
	onClose,
	onToggle,
	onCreate,
	onUpdate,
	onDelete,
}: {
	company: CompanyListItem | null;
	busy: boolean;
	onClose: () => void;
	onToggle: (boardID: number, active: boolean) => void;
	onCreate: (input: BoardInput) => void;
	onUpdate: (boardID: number, input: BoardInput) => void;
	onDelete: (boardID: number) => void;
}) {
	const [editing, setEditing] = useState<
		CompanyListItem["career_boards"][number] | null | "new"
	>(null);
	const saveBoard = (input: BoardInput) => {
		if (editing === "new") onCreate(input);
		else if (editing) onUpdate(editing.id, input);
		setEditing(null);
	};
	return (
		<Modal
			open={company !== null}
			onClose={onClose}
			title={company?.name ?? "Career boards"}
			subtitle="Career board sources"
			icon={<Layers3 size={16} className="text-[#7dba7a]" />}
			panelClassName="max-w-2xl"
			headerActions={
				<button
					type="button"
					onClick={() => setEditing("new")}
					disabled={!company || busy}
					className="inline-flex items-center gap-1.5 rounded-lg bg-[#1a2a1a] px-2.5 py-1.5 text-xs font-medium text-[#c8d5c8] transition hover:bg-[#253425] disabled:opacity-50"
				>
					<Plus size={14} /> Add board
				</button>
			}
		>
			<div className="max-h-[calc(92vh-4.5rem)] space-y-3 overflow-y-auto p-5 sm:p-6">
				{editing !== null && (
					<BoardForm
						board={editing === "new" ? null : editing}
						busy={busy}
						onCancel={() => setEditing(null)}
						onSave={saveBoard}
					/>
				)}
				{company?.career_boards.map((board) => (
					<div key={board.id} className="rounded-2xl bg-[#101710] p-4 sm:p-5">
						<div className="flex items-start justify-between gap-4">
							<div className="min-w-0">
								<div className="flex flex-wrap items-center gap-2">
									<span className="rounded-full bg-[#1d2d1d] px-2.5 py-1 text-xs font-medium text-[#b9cdb9]">
										{board.provider}
									</span>
									<span className="font-medium text-[#e8eee8]">
										{board.board_identifier}
									</span>
								</div>
								<a
									href={board.canonical_url}
									target="_blank"
									rel="noreferrer"
									title={board.canonical_url}
									className="mt-2 flex max-w-full items-center gap-1.5 truncate text-sm text-[#7dba7a] transition hover:text-[#a5d9a1]"
								>
									<span className="truncate">{board.canonical_url}</span>
									<ExternalLink size={14} className="shrink-0" />
								</a>
							</div>
							<div className="flex shrink-0 items-center gap-2">
								<button
									type="button"
									onClick={() => setEditing(board)}
									disabled={busy}
									aria-label={`Edit ${board.board_identifier} board`}
									className="rounded-lg p-2 text-[#8b9b8b] transition hover:bg-[#1d2d1d] hover:text-[#d9e3d9] disabled:opacity-40"
								>
									<Pencil size={15} />
								</button>
								<button
									type="button"
									onClick={() => onDelete(board.id)}
									disabled={busy}
									aria-label={`Delete ${board.board_identifier} board`}
									className="rounded-lg p-2 text-[#a87f7f] transition hover:bg-red-400/10 hover:text-red-200 disabled:opacity-40"
								>
									<Trash2 size={15} />
								</button>
								<span className="text-xs text-[#718071]">
									{board.active ? "Enabled" : "Disabled"}
								</span>
								<Switch
									aria-label={`${company?.name} ${board.board_identifier} enabled`}
									checked={board.active}
									disabled={
										busy ||
										!company?.active ||
										board.adapter_status === "unsupported"
									}
									onCheckedChange={(active) => onToggle(board.id, active)}
								/>
							</div>
						</div>
						<div className="mt-4 grid grid-cols-2 gap-x-5 gap-y-4 border-t border-[#203020] pt-4 sm:grid-cols-4">
							<BoardStatus kind="adapter" status={board.adapter_status} />
							<BoardStatus kind="freshness" status={board.freshness_status} />
							<BoardTime label="Last scan" value={board.last_scan_attempt_at} />
							<BoardTime
								label="Latest discovery"
								value={board.last_new_role_discovery_at}
							/>
						</div>
						{board.last_scan_failure_detail && (
							<div className="mt-4 rounded-xl bg-red-400/8 px-3 py-2 text-sm text-red-200">
								<span className="font-medium">Latest scan failed:</span>{" "}
								{board.last_scan_failure_detail}
							</div>
						)}
					</div>
				))}
				{company?.career_boards.length === 0 && (
					<p className="py-8 text-center text-sm text-[#718071]">
						No career boards registered.
					</p>
				)}
			</div>
		</Modal>
	);
}

type BoardInput = {
	provider: string;
	board_identifier: string;
	canonical_url: string;
};

function BoardForm({
	board,
	busy,
	onCancel,
	onSave,
}: {
	board: CompanyListItem["career_boards"][number] | null;
	busy: boolean;
	onCancel: () => void;
	onSave: (input: BoardInput) => void;
}) {
	const [provider, setProvider] = useState(board?.provider ?? "greenhouse");
	const [identifier, setIdentifier] = useState(board?.board_identifier ?? "");
	const [url, setURL] = useState(board?.canonical_url ?? "");
	return (
		<form
			className="rounded-2xl bg-[#172217] p-4"
			onSubmit={(event) => {
				event.preventDefault();
				onSave({ provider, board_identifier: identifier, canonical_url: url });
			}}
		>
			<p className="text-sm font-semibold text-[#e8eee8]">
				{board ? "Edit board" : "Add career board"}
			</p>
			<div className="mt-3 grid gap-2 sm:grid-cols-3">
				<input
					aria-label="Board provider"
					required
					value={provider}
					onChange={(event) => setProvider(event.target.value)}
					placeholder="Provider"
					className="rounded-xl bg-[#0d120d] px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-[#7dba7a]/40"
				/>
				<input
					aria-label="Board identifier"
					required
					value={identifier}
					onChange={(event) => setIdentifier(event.target.value)}
					placeholder="Board identifier"
					className="rounded-xl bg-[#0d120d] px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-[#7dba7a]/40"
				/>
				<input
					aria-label="Board URL"
					required
					type="url"
					value={url}
					onChange={(event) => setURL(event.target.value)}
					placeholder="https://…"
					className="rounded-xl bg-[#0d120d] px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-[#7dba7a]/40"
				/>
			</div>
			<div className="mt-3 flex items-center gap-2">
				<button
					type="submit"
					disabled={busy}
					className="rounded-lg bg-[#7dba7a] px-3 py-2 text-sm font-semibold text-[#080908] disabled:opacity-50"
				>
					Save board
				</button>
				<button
					type="button"
					onClick={onCancel}
					className="rounded-lg px-3 py-2 text-sm text-[#8b9b8b] hover:bg-[#203020]"
				>
					Cancel
				</button>
			</div>
		</form>
	);
}

function CompanyEditModal({
	company,
	busy,
	onClose,
	onSave,
	onDelete,
}: {
	company: CompanyListItem | null;
	busy: boolean;
	onClose: () => void;
	onSave: (name: string) => void;
	onDelete: () => void;
}) {
	const [name, setName] = useState("");
	useEffect(() => {
		if (company) {
			setName(company.name);
		}
	}, [company]);
	return (
		<Modal
			open={company !== null}
			onClose={onClose}
			title="Edit company"
			subtitle="Update company details or remove this company."
			icon={<Pencil size={16} className="text-[#7dba7a]" />}
			panelClassName="max-w-lg"
		>
			<form
				className="space-y-4 p-5"
				onSubmit={(event) => {
					event.preventDefault();
					onSave(name);
				}}
			>
				<label className="block text-sm text-[#c8d5c8]">
					Name
					<input
						aria-label="Company name"
						required
						value={name}
						onChange={(event) => setName(event.target.value)}
						className="mt-1.5 w-full rounded-xl bg-[#101710] px-3 py-2.5 outline-none focus:ring-2 focus:ring-[#7dba7a]/40"
					/>
				</label>
				<div className="flex items-center justify-between gap-3 pt-2">
					<button
						type="button"
						disabled={busy}
						onClick={onDelete}
						className="inline-flex items-center gap-1.5 rounded-lg px-3 py-2 text-sm text-red-300 transition hover:bg-red-400/10 disabled:opacity-50"
					>
						<Trash2 size={15} /> Delete company
					</button>
					<div className="flex gap-2">
						<button
							type="button"
							onClick={onClose}
							className="rounded-lg px-3 py-2 text-sm text-[#8b9b8b] hover:bg-[#1a2a1a]"
						>
							Cancel
						</button>
						<button
							type="submit"
							disabled={busy}
							className="rounded-lg bg-[#7dba7a] px-3 py-2 text-sm font-semibold text-[#080908] disabled:opacity-50"
						>
							Save changes
						</button>
					</div>
				</div>
			</form>
		</Modal>
	);
}

function BoardStatus({
	kind,
	status,
}: {
	kind: "adapter" | "freshness";
	status: string;
}) {
	const config =
		kind === "adapter"
			? adapterStatusDetail(status)
			: freshnessStatusDetail(status);
	const Icon = config.icon;
	return (
		<div className="min-w-0">
			<p className="text-[11px] font-medium tracking-wide text-[#718071] uppercase">
				{kind === "adapter" ? "Adapter" : "Freshness"}
			</p>
			<p
				className={`mt-1 flex items-center gap-1.5 text-sm font-medium ${config.tone}`}
			>
				<Icon size={15} aria-hidden="true" />
				{config.label}
			</p>
			<p className="mt-1 text-xs leading-4 text-[#718071]">{config.detail}</p>
		</div>
	);
}

function BoardTime({ label, value }: { label: string; value: string | null }) {
	return (
		<div className="min-w-0">
			<p className="text-[11px] font-medium tracking-wide text-[#718071] uppercase">
				{label}
			</p>
			<p
				className="mt-1 truncate text-sm text-[#c8d5c8]"
				title={value ? formatDate(value) : undefined}
			>
				{relativeDate(value)}
			</p>
		</div>
	);
}

type StatusDetail = {
	label: string;
	detail: string;
	tone: string;
	icon: LucideIcon;
};

function adapterStatusDetail(status: string): StatusDetail {
	return (
		{
			healthy: {
				label: "Healthy",
				detail: "Scanner is available and the latest scan succeeded.",
				tone: "text-emerald-300",
				icon: Sparkles,
			},
			failing: {
				label: "Failing",
				detail: "The latest scan did not complete successfully.",
				tone: "text-red-300",
				icon: CircleAlert,
			},
			unsupported: {
				label: "Unsupported",
				detail: "No scanner is available for this source.",
				tone: "text-amber-300",
				icon: ShieldOff,
			},
			unknown: {
				label: "Not scanned yet",
				detail: "This source has not completed a scan yet.",
				tone: "text-[#a9b5a9]",
				icon: Clock3,
			},
		}[status] ?? {
			label: labelStatus(status),
			detail: "Scanner status is not available.",
			tone: "text-[#a9b5a9]",
			icon: CircleHelp,
		}
	);
}

function freshnessStatusDetail(status: string): StatusDetail {
	return (
		{
			fresh: {
				label: "Fresh",
				detail: "New roles were found within the last 30 days.",
				tone: "text-emerald-300",
				icon: Sparkles,
			},
			stale: {
				label: "Stale",
				detail: "No new roles have been found in 30 days.",
				tone: "text-amber-300",
				icon: CircleAlert,
			},
			no_activity_yet: {
				label: "No activity yet",
				detail: "This source has not found any roles yet.",
				tone: "text-[#a9b5a9]",
				icon: Clock3,
			},
			not_applicable: {
				label: "Not applicable",
				detail: "Freshness is unavailable while this source cannot scan.",
				tone: "text-[#a9b5a9]",
				icon: ShieldOff,
			},
			unknown: {
				label: "Unknown",
				detail: "No freshness information is available yet.",
				tone: "text-[#a9b5a9]",
				icon: CircleHelp,
			},
		}[status] ?? {
			label: labelStatus(status),
			detail: "Freshness information is not available.",
			tone: "text-[#a9b5a9]",
			icon: CircleHelp,
		}
	);
}

function DiscoveryPanel({
	discover,
	register,
	onRegistered,
}: {
	discover: ReturnType<typeof useDiscoverCareerBoards>;
	register: ReturnType<typeof useRegisterCareerBoards>;
	onRegistered: (company: CompanyListItem) => void;
}) {
	const [careersURL, setCareersURL] = useState("");
	const [selected, setSelected] = useState<Set<string>>(new Set());
	const [employerName, setEmployerName] = useState("");
	const [separate, setSeparate] = useState<Set<string>>(new Set());
	const [separateNames, setSeparateNames] = useState<Record<string, string>>(
		{},
	);
	const candidates = discover.data?.candidates ?? [];
	const discoveryError =
		discover.error instanceof Error
			? discover.error.message
			: "Could not analyze the careers URL.";
	const selectableURLs = candidates
		.filter(
			(candidate) =>
				candidate.validation_status === "ready" ||
				candidate.validation_status === "unsupported",
		)
		.map((candidate) => candidate.canonical_url);
	const selectAll = () => setSelected(new Set(selectableURLs));
	const unselectAll = () => setSelected(new Set());
	const toggle = (url: string) =>
		setSelected((current) => {
			const next = new Set(current);
			next.has(url) ? next.delete(url) : next.add(url);
			return next;
		});
	const toggleSeparate = (url: string, fallbackName: string) =>
		setSeparate((current) => {
			const next = new Set(current);
			if (next.has(url)) {
				next.delete(url);
			} else {
				next.add(url);
				setSeparateNames((names) => ({
					...names,
					[url]: names[url] || fallbackName,
				}));
			}
			return next;
		});
	const sharedSelectedCount = candidates.filter(
		(candidate) =>
			selected.has(candidate.canonical_url) &&
			!separate.has(candidate.canonical_url),
	).length;
	const addSelected = () => {
		const registrations = candidates
			.filter((candidate) => selected.has(candidate.canonical_url))
			.map((candidate) => ({
				company_name: separate.has(candidate.canonical_url)
					? separateNames[candidate.canonical_url]?.trim() ||
						candidate.board_identifier
					: employerName.trim(),
				careers_url: careersURL,
				provider: candidate.provider,
				board_identifier: candidate.board_identifier,
				canonical_url: candidate.canonical_url,
				separate: separate.has(candidate.canonical_url),
			}));
		register.mutate(
			{ candidates: registrations },
			{
				onSuccess: (result) => {
					const company = result.data.find((item) =>
						item.career_boards.some((board) =>
							registrations.some(
								(registration) =>
									registration.provider === board.provider &&
									registration.board_identifier === board.board_identifier,
							),
						),
					);
					setSelected(new Set());
					setEmployerName("");
					setSeparate(new Set());
					setSeparateNames({});
					setCareersURL("");
					discover.reset();
					if (company) onRegistered(company);
				},
			},
		);
	};
	return (
		<div className="max-h-[calc(92vh-4.5rem)] overflow-y-auto p-5 sm:p-6">
			<div>
				<form
					className="flex flex-col gap-2 sm:flex-row"
					onSubmit={(event) => {
						event.preventDefault();
						setSelected(new Set());
						setEmployerName("");
						setSeparate(new Set());
						setSeparateNames({});
						discover.mutate(careersURL, {
							onSuccess: (result) => {
								setEmployerName(result.employer_name_suggestion);
								setSelected(
									new Set(
										result.candidates
											.filter(
												(candidate) =>
													candidate.validation_status === "ready" ||
													candidate.validation_status === "unsupported",
											)
											.map((candidate) => candidate.canonical_url),
									),
								);
							},
						});
					}}
				>
					<input
						required
						type="url"
						value={careersURL}
						onChange={(event) => setCareersURL(event.target.value)}
						placeholder="https://company.example/careers"
						className="min-w-0 flex-1 rounded-xl bg-[#080b08] px-4 py-3 text-sm ring-1 ring-[#2a3a2a] outline-none transition focus:ring-2 focus:ring-[#7dba7a]/50"
					/>
					<button
						type="submit"
						disabled={discover.isPending}
						className="rounded-xl bg-linear-to-r from-[#7dba7a] to-[#5a8f5a] px-4 py-3 text-sm font-semibold text-[#080908] transition hover:from-[#8dca8a] hover:to-[#6a9f6a] disabled:cursor-not-allowed disabled:opacity-50"
					>
						{discover.isPending ? "Discovering…" : "Discover"}
					</button>
				</form>
				{discover.error && (
					<p className="mt-2 text-sm text-red-300">{discoveryError}</p>
				)}
			</div>
			{candidates.length > 0 && (
				<div className="mt-7">
					<div className="grid gap-6 lg:grid-cols-[minmax(14rem,0.65fr)_minmax(0,1.8fr)] lg:items-start lg:gap-8">
						<div className="rounded-2xl bg-[#172217] p-4">
							<div className="flex items-center gap-2 text-sm font-medium text-[#d9e3d9]">
								<Building2 size={16} className="text-[#7dba7a]" />
								Add to company
							</div>
							<input
								value={employerName}
								onChange={(event) => setEmployerName(event.target.value)}
								placeholder="Company name"
								className="mt-3 w-full rounded-xl bg-[#0d120d] px-3 py-2.5 text-sm outline-none transition placeholder:text-[#526052] focus:ring-2 focus:ring-[#7dba7a]/40"
							/>
						</div>
						<div className="min-w-0">
							<div className="flex items-center justify-between px-1">
								<p className="font-medium text-[#d9e3d9]">Choose sources</p>
								<span className="rounded-full bg-[#1a2a1a] px-2.5 py-1 text-xs text-[#8b9b8b]">
									{selected.size} selected
								</span>
							</div>
							<div className="mt-3 space-y-2">
								{candidates.map((candidate) => (
									<DiscoveryCandidateCard
										key={candidate.canonical_url}
										candidate={candidate}
										selected={selected.has(candidate.canonical_url)}
										separate={separate.has(candidate.canonical_url)}
										separateName={separateNames[candidate.canonical_url] ?? ""}
										onSelect={() => toggle(candidate.canonical_url)}
										onToggleSeparate={() =>
											toggleSeparate(
												candidate.canonical_url,
												candidate.board_identifier,
											)
										}
										onSeparateNameChange={(name) =>
											setSeparateNames((current) => ({
												...current,
												[candidate.canonical_url]: name,
											}))
										}
									/>
								))}
							</div>
						</div>
					</div>
					<div className="mt-6 flex flex-wrap items-center gap-2">
						<button
							type="button"
							disabled={
								selected.size === 0 ||
								(sharedSelectedCount > 0 && !employerName.trim()) ||
								register.isPending
							}
							onClick={addSelected}
							className="rounded-lg bg-linear-to-r from-[#7dba7a] to-[#5a8f5a] px-4 py-2 text-sm font-semibold text-[#080908] transition hover:from-[#8dca8a] hover:to-[#6a9f6a] disabled:cursor-not-allowed disabled:opacity-50"
						>
							{register.isPending
								? "Adding sources…"
								: `Add ${selected.size} source${selected.size === 1 ? "" : "s"}${
										sharedSelectedCount > 0 && employerName.trim()
											? ` to ${employerName.trim()}`
											: ""
									}`}
						</button>
						<button
							type="button"
							disabled={selected.size === selectableURLs.length}
							onClick={selectAll}
							className="rounded-lg px-3 py-2 text-sm text-[#6a7a6a] transition hover:bg-[#1a2a1a] hover:text-[#c8d5c8] disabled:opacity-50"
						>
							Select all
						</button>
						<button
							type="button"
							disabled={selected.size === 0}
							onClick={unselectAll}
							className="rounded-lg px-3 py-2 text-sm text-[#6a7a6a] transition hover:bg-[#1a2a1a] hover:text-[#c8d5c8] disabled:opacity-50"
						>
							Unselect all
						</button>
						{discover.data?.incomplete && (
							<span className="text-xs text-amber-300">
								Results may be incomplete.
							</span>
						)}
					</div>
				</div>
			)}
		</div>
	);
}

function DiscoveryCandidateCard({
	candidate,
	selected,
	onSelect,
	separate,
	separateName,
	onToggleSeparate,
	onSeparateNameChange,
}: {
	candidate: CareerBoardDiscoveryCandidate;
	selected: boolean;
	onSelect: () => void;
	separate: boolean;
	separateName: string;
	onToggleSeparate: () => void;
	onSeparateNameChange: (value: string) => void;
}) {
	const selectable =
		candidate.validation_status === "ready" ||
		candidate.validation_status === "unsupported";
	return (
		<div
			className={`rounded-2xl p-3.5 text-sm transition-all ${
				selected
					? "bg-[#101710] ring-1 ring-[#7dba7a]/55"
					: "bg-[#101710] hover:bg-[#142014]"
			}`}
		>
			<div className="flex items-start gap-3">
				<button
					type="button"
					aria-label={`${selected ? "Unselect" : "Select"} ${candidate.board_identifier}`}
					aria-pressed={selected}
					disabled={!selectable}
					onClick={onSelect}
					className={`mt-0.5 flex size-5 shrink-0 items-center justify-center rounded-full border transition ${
						selected
							? "border-[#7dba7a] bg-[#7dba7a] text-[#080908]"
							: "border-[#526052] text-transparent hover:border-[#7dba7a] disabled:opacity-35"
					}`}
				>
					<Check size={13} strokeWidth={3} />
				</button>
				<div className="min-w-0 flex-1">
					<div className="flex items-center gap-2">
						<span className="text-base leading-5 font-semibold text-[#e8eee8]">
							{candidate.board_identifier}
						</span>
						{candidate.evidence_urls.length > 0 && (
							<span
								className="text-[#718071]"
								title={`Found on ${candidate.evidence_urls[0]}`}
							>
								<CircleHelp size={14} aria-label="Discovery evidence" />
							</span>
						)}
					</div>
					<div className="mt-1.5 flex flex-wrap items-center gap-2">
						<span className="text-sm text-[#8b9b8b]">{candidate.provider}</span>
						{candidate.validation_status !== "ready" && (
							<span
								className={`rounded-full px-2 py-0.5 text-[11px] ${
									candidate.validation_status === "unsupported"
										? "bg-amber-400/10 text-amber-300"
										: "bg-red-400/10 text-red-300"
								}`}
							>
								{labelStatus(candidate.validation_status)}
							</span>
						)}
					</div>
				</div>
				<a
					href={candidate.canonical_url}
					target="_blank"
					rel="noreferrer"
					title="Open source board"
					aria-label={`Open ${candidate.board_identifier} job board`}
					className="flex size-8 shrink-0 items-center justify-center rounded-full text-[#7dba7a] transition hover:bg-[#253425] hover:text-[#a5d9a1]"
				>
					<ExternalLink size={16} />
				</a>
				{selectable && selected && (
					<button
						type="button"
						onClick={onToggleSeparate}
						className="self-center rounded-full bg-[#0d120d] px-3 py-1.5 text-xs text-[#8b9b8b] transition hover:bg-[#253425] hover:text-[#d9e3d9]"
					>
						{separate ? "Use shared company" : "Add separately"}
					</button>
				)}
			</div>
			{separate && (
				<div className="ml-8 mt-2.5 flex items-center gap-3 rounded-xl bg-[#0d120d] px-3 py-2">
					<label
						htmlFor={`separate-company-${candidate.board_identifier}`}
						className="shrink-0 text-xs font-medium text-[#a9b5a9]"
					>
						Separate company
					</label>
					<input
						id={`separate-company-${candidate.board_identifier}`}
						value={separateName}
						onChange={(event) => onSeparateNameChange(event.target.value)}
						className="min-w-0 flex-1 bg-transparent py-1 text-sm outline-none placeholder:text-[#526052]"
					/>
				</div>
			)}
			{candidate.validation_error && (
				<p className="mt-1 text-xs text-red-300">
					{candidate.validation_error}
				</p>
			)}
		</div>
	);
}

function CompanyRow({
	company,
	selected,
	onSelect,
	onToggle,
	onManageBoards,
	onEdit,
	busy,
}: {
	company: CompanyListItem;
	selected: boolean;
	onSelect: (id: number) => void;
	onToggle: (active: boolean) => void;
	onManageBoards: () => void;
	onEdit: () => void;
	busy: boolean;
}) {
	return (
		<article className="group flex flex-col gap-4 rounded-2xl bg-[#101710] p-4 transition hover:bg-[#142014] sm:flex-row sm:items-center sm:p-5">
			<button
				type="button"
				aria-label={`${selected ? "Unselect" : "Select"} ${company.name}`}
				aria-pressed={selected}
				onClick={() => onSelect(company.id)}
				className={`flex size-5 shrink-0 items-center justify-center rounded-full border transition ${selected ? "border-[#7dba7a] bg-[#7dba7a] text-[#080908]" : "border-[#526052] text-transparent hover:border-[#7dba7a]"}`}
			>
				<Check size={13} strokeWidth={3} />
			</button>
			<div className="min-w-0 flex-1">
				<div className="flex flex-wrap items-center gap-x-2 gap-y-1">
					<h2 className="font-semibold text-[#e8eee8]">{company.name}</h2>
					{!company.active && (
						<span className="rounded-full bg-[#2b2525] px-2 py-0.5 text-[11px] text-[#d8aaaa]">
							Disabled
						</span>
					)}
				</div>
			</div>
			<div className="flex flex-wrap items-center gap-3 sm:flex-nowrap sm:gap-4">
				<button
					type="button"
					onClick={onEdit}
					disabled={busy}
					aria-label={`Edit ${company.name}`}
					className="rounded-xl p-2 text-[#8b9b8b] transition hover:bg-[#1b2a1b] hover:text-[#d9e3d9] disabled:opacity-40"
				>
					<Pencil size={16} />
				</button>
				<button
					type="button"
					onClick={onManageBoards}
					className="inline-flex items-center gap-2 rounded-xl bg-[#1b2a1b] px-3 py-2 text-sm font-medium text-[#c8d5c8] transition hover:bg-[#284028] hover:text-[#e8eee8]"
					aria-label={`Manage ${company.name} career boards`}
				>
					<Layers3 size={16} className="text-[#7dba7a]" />
					{company.board_count} board{company.board_count === 1 ? "" : "s"}
					<ChevronRight size={15} className="text-[#8b9b8b]" />
				</button>
				<span className="inline-flex items-center gap-1.5 text-sm text-[#a9b5a9]">
					<BriefcaseBusiness size={15} className="text-[#718071]" />
					{company.role_count} jobs
				</span>
				<FreshnessIndicator status={company.freshness_status} />
			</div>
			<div className="flex shrink-0 items-center sm:pl-3">
				<Switch
					aria-label={`${company.name} enabled`}
					checked={company.active}
					disabled={busy}
					onCheckedChange={onToggle}
				/>
			</div>
		</article>
	);
}

function FreshnessIndicator({ status }: { status: string }) {
	const detail =
		{
			fresh: "At least one board found new roles within the last 30 days",
			stale: "At least one board has no new roles in 30 days",
			failing: "At least one board's latest scan failed",
			no_activity_yet: "No board has discovered roles yet",
			not_applicable: "Freshness is not tracked for this company",
			unknown: "No board activity has been recorded yet",
		}[status] ?? labelStatus(status);
	const Icon =
		{
			fresh: Sparkles,
			stale: CircleAlert,
			failing: CircleAlert,
			no_activity_yet: Clock3,
			not_applicable: ShieldOff,
			unknown: CircleHelp,
		}[status] ?? CircleHelp;
	const tone =
		status === "failing"
			? "text-red-300"
			: status === "stale"
				? "text-amber-300"
				: status === "fresh"
					? "text-emerald-300"
					: "text-[#a9b5a9]";
	return (
		<Tooltip content={detail}>
			<span className={`inline-flex ${tone}`} role="img" aria-label={detail}>
				<Icon size={17} aria-hidden="true" />
			</span>
		</Tooltip>
	);
}
