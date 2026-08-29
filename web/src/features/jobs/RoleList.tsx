import { Bookmark, ChevronDown, Clock } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { usePatchRole, useRoles } from "../../hooks/useApi";
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
	const listRef = useRef<HTMLDivElement>(null);
	const [scrollProgress, setScrollProgress] = useState(0);
	const [animating, setAnimating] = useState(false);
	const animTimer = useRef<ReturnType<typeof setTimeout>>(undefined);
	const patchRole = usePatchRole();

	const goToPage = (fn: (p: number) => number) => {
		setPage(fn);
		listRef.current?.scrollTo({ top: 0 });
		setAnimating(true);
		setScrollProgress(0);
		clearTimeout(animTimer.current);
		animTimer.current = setTimeout(() => setAnimating(false), 200);
	};

	useEffect(() => {
		const el = listRef.current;
		if (!el) return;
		const onScroll = () => {
			const pct =
				el.scrollHeight - el.clientHeight > 0
					? el.scrollTop / (el.scrollHeight - el.clientHeight)
					: 0;
			setScrollProgress(pct);
		};
		el.addEventListener("scroll", onScroll, { passive: true });
		return () => el.removeEventListener("scroll", onScroll);
	}, []);

	return (
		<main className="flex flex-1 flex-col overflow-hidden">
			<div className="flex items-center justify-between border-b border-[#1a2a1a] px-6 py-3">
				<div className="flex items-center gap-2">
					<h2 className="text-base font-semibold">Roles</h2>
					<span className="rounded-full bg-[#1a2a1a] px-2 py-0.5 text-xs text-[#6a7a6a]">
						{total}
					</span>
				</div>
				<div className="flex items-center gap-2 text-xs text-[#6a7a6a]">
					<span className="w-16 text-right pr-5">Match</span>
				</div>
			</div>

			<div ref={listRef} className="flex-1 overflow-y-auto scrollbar-hide">
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
						<p className="text-base">No roles found</p>
						<p className="mt-1 text-xs">Run a scan to find jobs</p>
					</div>
				) : (
					<div className="space-y-1 p-2">
						{roles.map((role) => (
							<div
								key={role.id}
								className={`flex items-center rounded-xl border px-4 py-3 transition ${
									selectedId === role.id
										? "border-[#7dba7a]/40 bg-[#1a2a1a]"
										: "border-transparent bg-transparent hover:bg-[#0d120d]"
								}`}
							>
								<button
									type="button"
									onClick={() => onSelect(role.id)}
									className="min-w-0 flex-1 text-left"
								>
									<div className="min-w-0">
										<p className="line-clamp-2 break-words text-sm leading-5 font-medium text-[#e8e8e8]">
											{role.title}
										</p>
										<p className="mt-0.5 text-sm text-[#6a7a6a]">
											{role.company_name}
											{role.location && (
												<>
													<span className="mx-1">·</span>
													{role.location}
												</>
											)}
										</p>
										<p className="mt-1 flex items-center gap-1 text-xs text-[#4a5a4a]">
											<Clock size={10} />
											{timeAgo(role.posted_at)}
										</p>
									</div>
								</button>
								<div className="flex shrink-0 items-center gap-1.5">
									<button
										type="button"
										onClick={(e) => {
											e.stopPropagation();
											patchRole.mutate({
												id: role.id,
												is_interested: !role.is_interested,
											});
										}}
										className="cursor-pointer rounded p-0.5 transition hover:text-[#7dba7a]"
										aria-label={
											role.is_interested
												? "Remove bookmark"
												: "Bookmark this role"
										}
									>
										<Bookmark
											size={14}
											className={`hover:fill-[#7dba7a44] ${
												role.is_interested
													? "fill-[#7dba7a] text-[#7dba7a]"
													: "text-[#4a5a4a]"
											}`}
										/>
									</button>
									<span
										className={`inline-flex min-w-8 items-center justify-center rounded-md border px-2 py-0.5 text-sm font-semibold ${scoreColor(role.match_percent ?? 0)}`}
									>
										✦ {matchLabel(role.match_percent ?? 0)}
									</span>
								</div>
							</div>
						))}
					</div>
				)}
			</div>

			{totalPages > 1 && (
				<div className="relative border-t border-[#1a2a1a] px-6 py-3">
					<div
						className="pointer-events-none absolute bottom-full left-1/2 mb-2 -translate-x-1/2 transition-opacity duration-300"
						style={{
							opacity: scrollProgress < 0.95 ? 1 : 0,
						}}
					>
						<div
							className={`flex items-center justify-center rounded-full bg-[#1a2a1a] p-1.5 aspect-square animate-bounce`}
						>
							<ChevronDown size={14} className="text-[#7dba7a]" />
						</div>
					</div>
					<div className="grid grid-cols-3 items-center">
						<button
							type="button"
							onClick={() => goToPage((p) => Math.max(1, p - 1))}
							disabled={page <= 1}
							className="justify-self-start rounded-lg px-3 py-1.5 text-sm text-[#6a7a6a] transition hover:bg-[#1a2a1a] disabled:opacity-30"
						>
							Previous
						</button>
						<div className="flex flex-col items-center gap-1.5 justify-self-center">
							<span className="text-xs text-[#4a5a4a]">
								Page {page} of {totalPages}
							</span>
							<div className="h-1 w-32 rounded-full bg-[#1a2a1a]">
								<div
									className={`h-full rounded-full bg-[#7dba7a] ${animating ? "transition-all duration-200" : ""}`}
									style={{ width: `${scrollProgress * 100}%` }}
								/>
							</div>
						</div>
						<button
							type="button"
							onClick={() => goToPage((p) => Math.min(totalPages, p + 1))}
							disabled={page >= totalPages}
							className="justify-self-end rounded-lg px-3 py-1.5 text-sm text-[#6a7a6a] transition hover:bg-[#1a2a1a] disabled:opacity-30"
						>
							Next
						</button>
					</div>
				</div>
			)}
		</main>
	);
}
