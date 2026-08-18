import { Bookmark, Clock } from "lucide-react";
import { useState } from "react";
import { useRoles } from "../../hooks/useApi";
import { matchLabel, scoreColor, timeAgo } from "../../lib/dashboard";
import type { RoleListItem } from "../../types/api.gen";

interface RoleListProps {
	selectedId: number | null;
	onSelect: (id: number) => void;
}

export function RoleList({ selectedId, onSelect }: RoleListProps) {
	const [page, setPage] = useState(1);
	const { data: rolesData, isLoading } = useRoles(page);
	const roles = (rolesData?.data ?? []) as RoleListItem[];
	const total = rolesData?.pagination?.total ?? 0;
	const totalPages = Math.ceil(total / 25);

	return (
		<main className="flex flex-1 flex-col overflow-hidden">
			<div className="flex items-center justify-between border-b border-[#1a2a1a] px-6 py-3">
				<div className="flex items-center gap-2">
					<h2 className="text-sm font-semibold">Roles</h2>
					<span className="rounded-full bg-[#1a2a1a] px-2 py-0.5 text-xs text-[#6a7a6a]">
						{total}
					</span>
				</div>
				<div className="flex items-center gap-2 text-xs text-[#6a7a6a]">
					<span className="w-16 text-right">Match</span>
				</div>
			</div>

			<div className="flex-1 overflow-y-auto">
				{isLoading ? (
					<div className="space-y-2 p-4">
						{[1, 2, 3, 4, 5].map((i) => (
							<div
								key={i}
								className="h-16 animate-pulse rounded-lg bg-[#0d120d]"
							/>
						))}
					</div>
				) : roles.length === 0 ? (
					<div className="flex h-full flex-col items-center justify-center text-[#4a5a4a]">
						<p className="text-sm">No roles found</p>
						<p className="mt-1 text-xs">Run a scan to find jobs</p>
					</div>
				) : (
					<div className="space-y-1 p-2">
						{roles.map((role) => (
							<button
								key={role.id}
								onClick={() => onSelect(role.id)}
								className={`w-full rounded-xl border px-4 py-3 text-left transition ${
									selectedId === role.id
										? "border-[#7dba7a]/40 bg-[#1a2a1a]"
										: "border-transparent bg-transparent hover:bg-[#0d120d]"
								}`}
							>
								<div className="flex items-start justify-between gap-3">
									<div className="min-w-0 flex-1">
										<p className="truncate text-sm font-medium text-[#e8e8e8]">
											{role.title}
										</p>
										<p className="mt-0.5 text-xs text-[#6a7a6a]">
											{role.company_name}
											{role.location && (
												<>
													<span className="mx-1">·</span>
													{role.location}
												</>
											)}
										</p>
										<p className="mt-1 flex items-center gap-1 text-[10px] text-[#4a5a4a]">
											<Clock size={10} />
											{timeAgo(role.posted_at)}
										</p>
									</div>
									<div className="flex flex-col items-end gap-1">
										<span
											className={`inline-flex min-w-[2rem] items-center justify-center rounded-md border px-2 py-0.5 text-xs font-semibold ${scoreColor(role.match_percent ?? 0)}`}
										>
											{matchLabel(role.match_percent ?? 0)}
										</span>
										{role.is_interested && (
											<Bookmark
												size={12}
												className="text-[#7dba7a]"
												fill="#7dba7a"
											/>
										)}
									</div>
								</div>
							</button>
						))}
					</div>
				)}
			</div>

			{totalPages > 1 && (
				<div className="flex items-center justify-between border-t border-[#1a2a1a] px-6 py-3">
					<button
						onClick={() => setPage((p) => Math.max(1, p - 1))}
						disabled={page <= 1}
						className="rounded-lg px-3 py-1.5 text-xs text-[#6a7a6a] transition hover:bg-[#1a2a1a] disabled:opacity-30"
					>
						Previous
					</button>
					<span className="text-xs text-[#4a5a4a]">
						Page {page} of {totalPages}
					</span>
					<button
						onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
						disabled={page >= totalPages}
						className="rounded-lg px-3 py-1.5 text-xs text-[#6a7a6a] transition hover:bg-[#1a2a1a] disabled:opacity-30"
					>
						Next
					</button>
				</div>
			)}
		</main>
	);
}
