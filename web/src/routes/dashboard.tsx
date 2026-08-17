import { createFileRoute, redirect, useNavigate } from "@tanstack/react-router";
import { Bookmark, Clock, LogOut } from "lucide-react";
import { useEffect, useState } from "react";
import { useRoles } from "../hooks/useApi";
import { authApi } from "../lib/api";
import { useAuthStore } from "../stores/auth";

export const Route = createFileRoute("/dashboard")({
	beforeLoad: () => {
		const { authenticated } = useAuthStore.getState();
		if (!authenticated) {
			throw redirect({ to: "/" });
		}
	},
	component: DashboardPage,
});

interface Role {
	id: number;
	company_name: string;
	title: string;
	location: string;
	posted_at: string | null;
	relevance_score: number;
	is_interested: boolean;
	is_hidden: boolean;
}

function timeAgo(dateStr: string | null): string {
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

function scoreColor(score: number): string {
	if (score >= 3) return "text-green-400 bg-green-400/10 border-green-400/20";
	if (score >= 1) return "text-amber-400 bg-amber-400/10 border-amber-400/20";
	return "text-gray-500 bg-gray-500/10 border-gray-500/20";
}

function DashboardPage() {
	const navigate = useNavigate();
	const { check } = useAuthStore();
	const [selectedRole, setSelectedRole] = useState<Role | null>(null);
	const [page, setPage] = useState(1);

	useEffect(() => {
		check().then(() => {
			const { setupComplete } = useAuthStore.getState();
			if (!setupComplete) {
				navigate({ to: "/onboarding" });
			}
		});
	}, []);

	const handleLogout = async () => {
		try {
			await authApi.logout();
		} catch {
			// Ignore errors — the token is still cleared
		}
		useAuthStore.getState().logout();
		window.location.href = "/";
	};

	const { data: rolesData, isLoading: rolesLoading } = useRoles(page);
	const roles = rolesData?.data ?? [];
	const total = rolesData?.pagination?.total ?? 0;
	const totalPages = Math.ceil(total / 25);

	return (
		<div className="flex h-screen flex-col bg-[#080908] text-[#e8e8e8]">
			<header className="flex items-center justify-between border-b border-[#1a2a1a] px-6 py-3">
				<div className="flex items-center gap-3">
					<div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-[#7dba7a] to-[#4a7c4f]">
						<span className="text-sm">🍵</span>
					</div>
					<span className="text-lg font-bold text-[#e8e8e8]">jobmatcha</span>
				</div>
				<div className="flex items-center gap-4">
					<span className="text-xs text-[#6a7a6a]">{total} openings</span>
					<button
						onClick={handleLogout}
						className="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs text-[#6a7a6a] transition hover:bg-[#1a2a1a] hover:text-[#e8e8e8]"
					>
						<LogOut size={14} />
						Logout
					</button>
				</div>
			</header>

			<div className="flex flex-1 overflow-hidden">
				<aside className="w-72 shrink-0 border-r border-[#1a2a1a] p-4">
					<h3 className="mb-4 text-xs font-semibold uppercase tracking-wider text-[#6a7a6a]">
						Preferences
					</h3>
					<div className="space-y-4">
						<div>
							<label className="mb-1 block text-xs text-[#6a7a6a]">
								Include
							</label>
							<div className="flex flex-wrap gap-1.5">
								{["engineer", "go", "backend", "rust"].map((kw) => (
									<span
										key={kw}
										className="rounded-md border border-[#2a4a2a] bg-[#1a2a1a] px-2 py-1 text-xs text-[#7dba7a]"
									>
										+{kw}
									</span>
								))}
							</div>
						</div>
						<div>
							<label className="mb-1 block text-xs text-[#6a7a6a]">
								Exclude
							</label>
							<div className="flex flex-wrap gap-1.5">
								{["frontend", "manager"].map((kw) => (
									<span
										key={kw}
										className="rounded-md border border-[#3a2a2a] bg-[#2a1a1a] px-2 py-1 text-xs text-red-400"
									>
										-{kw}
									</span>
								))}
							</div>
						</div>
						<div>
							<label className="mb-1 block text-xs text-[#6a7a6a]">
								Location
							</label>
							<span className="rounded-md border border-[#2a3a2a] bg-[#1a1a2a] px-2 py-1 text-xs text-amber-400">
								Tokyo
							</span>
						</div>
					</div>
				</aside>

				<main className="flex flex-1 flex-col overflow-hidden">
					<div className="flex items-center justify-between border-b border-[#1a2a1a] px-6 py-3">
						<div className="flex items-center gap-2">
							<h2 className="text-sm font-semibold">Roles</h2>
							<span className="rounded-full bg-[#1a2a1a] px-2 py-0.5 text-xs text-[#6a7a6a]">
								{total}
							</span>
						</div>
						<div className="flex items-center gap-2 text-xs text-[#6a7a6a]">
							<span>Role</span>
							<span className="w-16 text-right">Match</span>
						</div>
					</div>

					<div className="flex-1 overflow-y-auto">
						{rolesLoading ? (
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
								{/* {roles.map((role) => (
                  <button
                    key={role.id}
                    onClick={() => setSelectedRole(role)}
                    className={`w-full rounded-xl border px-4 py-3 text-left transition ${
                      selectedRole?.id === role.id
                        ? 'border-[#7dba7a]/40 bg-[#1a2a1a]'
                        : 'border-transparent bg-transparent hover:bg-[#0d120d]'
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
                          className={`inline-flex min-w-[2rem] items-center justify-center rounded-md border px-2 py-0.5 text-xs font-semibold ${scoreColor(role.relevance_score)}`}
                        >
                          {role.relevance_score}
                        </span>
                        {role.is_interested && (
                          <Bookmark size={12} className="text-[#7dba7a]" fill="#7dba7a" />
                        )}
                      </div>
                    </div>
                  </button>
                ))} */}
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

				<aside className="w-96 shrink-0 border-l border-[#1a2a1a] overflow-y-auto">
					{selectedRole ? (
						<div className="p-6">
							<div className="mb-6">
								<p className="text-xs text-[#6a7a6a]">
									{selectedRole.company_name}
								</p>
								<h3 className="mt-1 text-lg font-bold text-[#e8e8e8]">
									{selectedRole.title}
								</h3>
								<div className="mt-3 flex flex-wrap items-center gap-3 text-xs text-[#6a7a6a]">
									<span className="flex items-center gap-1">
										{selectedRole.location}
									</span>
									<span
										className={`inline-flex items-center gap-1 rounded-md border px-2 py-0.5 font-semibold ${scoreColor(selectedRole.relevance_score)}`}
									>
										✦ Match Score: {selectedRole.relevance_score}
									</span>
								</div>
							</div>

							<div className="space-y-4">
								<div className="flex gap-2">
									<button
										className={`flex-1 rounded-lg border px-4 py-2 text-xs font-medium transition ${
											selectedRole.is_interested
												? "border-[#7dba7a] bg-[#7dba7a]/10 text-[#7dba7a]"
												: "border-[#2a3a2a] text-[#6a7a6a] hover:border-[#7dba7a]/50 hover:text-[#7dba7a]"
										}`}
									>
										{selectedRole.is_interested
											? "★ Interested"
											: "☆ Interested"}
									</button>
									<button className="flex-1 rounded-lg border border-[#2a3a2a] px-4 py-2 text-xs font-medium text-[#6a7a6a] transition hover:border-red-400/50 hover:text-red-400">
										{selectedRole.is_hidden ? "Hidden" : "Hide"}
									</button>
								</div>

								<div className="rounded-xl border border-[#1a2a1a] bg-[#0d120d] p-4">
									<p className="mb-2 text-xs font-medium text-[#6a7a6a]">
										Job URL
									</p>
									<p className="break-all text-xs text-[#7dba7a]">
										https://boards.greenhouse.io/jobs/123
									</p>
								</div>

								<div className="rounded-xl border border-[#1a2a1a] bg-[#0d120d] p-4">
									<p className="mb-2 text-xs font-medium text-[#6a7a6a]">
										Description
									</p>
									<p className="text-xs leading-relaxed text-[#9a9a9a]">
										No description loaded
									</p>
								</div>
							</div>
						</div>
					) : (
						<div className="flex h-full items-center justify-center text-[#4a5a4a]">
							<p className="text-xs">Select a role to view details</p>
						</div>
					)}
				</aside>
			</div>
		</div>
	);
}
