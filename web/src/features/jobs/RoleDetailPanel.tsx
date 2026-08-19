import { FileText, FolderTree, Globe, MapPin, User, Zap } from "lucide-react";
import { usePatchRole, useRole } from "../../hooks/useApi";
import { matchLabel, scoreColor } from "../../lib/dashboard";
import type { RoleDetailResponse } from "../../types/api.gen";

interface RoleDetailPanelProps {
	selectedId: number | null;
	onBack: () => void;
}

export function RoleDetailPanel({ selectedId, onBack }: RoleDetailPanelProps) {
	const { data: role } = useRole(selectedId);
	const patchRole = usePatchRole();

	if (!role) {
		return (
			<aside className="w-full h-full overflow-y-auto">
				<div className="flex h-full items-center justify-center text-[#4a5a4a]">
					<p className="text-sm">Select a role to view details</p>
				</div>
			</aside>
		);
	}

	const handleToggleInterested = () => {
		patchRole.mutate({ id: role.id, is_interested: !role.is_interested });
	};

	const handleToggleHidden = () => {
		patchRole.mutate({ id: role.id, is_hidden: !role.is_hidden });
	};

	return (
		<aside className="w-full h-full overflow-y-auto">
			<div className="p-6">
				<button
					type="button"
					onClick={onBack}
					className="mb-4 text-sm text-[#6a7a6a] transition hover:text-[#e8e8e8]"
				>
					← Back to list
				</button>

				<div className="mb-6">
					<p className="text-sm text-[#6a7a6a]">{role.company_name}</p>
					<h3 className="mt-1 text-lg font-bold text-[#e8e8e8]">
						{role.title}
					</h3>
					<div className="mt-3 flex flex-wrap items-center gap-3 text-sm text-[#6a7a6a]">
						<span className="flex items-center gap-1">{role.location}</span>
						<span
							className={`inline-flex items-center gap-1 rounded-md border px-2 py-0.5 font-semibold ${scoreColor(role.match_percent)}`}
						>
							✦ Match: {matchLabel(role.match_percent)}
						</span>
					</div>
				</div>

				<div className="space-y-4">
					<div className="flex gap-2">
						<button
							type="button"
							onClick={handleToggleInterested}
							disabled={patchRole.isPending}
							className={`flex-1 rounded-lg border px-4 py-2 text-sm font-medium ${
								role.is_interested
									? "border-[#7dba7a] bg-[#7dba7a]/10 text-[#7dba7a]"
									: "border-[#2a3a2a] text-[#6a7a6a] hover:border-[#7dba7a]/50 hover:text-[#7dba7a]"
							}`}
						>
							{patchRole.isPending
								? "Updating..."
								: role.is_interested
									? "★ Interested"
									: "☆ Mark Interested"}
						</button>
						<button
							type="button"
							onClick={handleToggleHidden}
							disabled={patchRole.isPending}
							className="flex-1 rounded-lg border border-[#2a3a2a] px-4 py-2 text-sm font-medium text-[#6a7a6a] transition hover:border-red-400/50 hover:text-red-400"
						>
							{role.is_hidden ? "Unhide" : "Hide"}
						</button>
					</div>

					{role.match_details && <MatchAnalysisCard role={role} />}

					{role.url && (
						<div className="rounded-xl border border-[#1a2a1a] bg-[#0d120d] p-4">
							<p className="mb-2 text-xs font-medium text-[#6a7a6a]">Job URL</p>
							<a
								href={role.url}
								target="_blank"
								rel="noopener noreferrer"
								className="break-all text-sm text-[#7dba7a] transition hover:underline"
							>
								{role.url}
							</a>
						</div>
					)}

					<div className="rounded-xl border border-[#1a2a1a] bg-[#0d120d] p-4">
						<p className="mb-2 text-xs font-medium text-[#6a7a6a]">
							Description
						</p>
						<p className="text-sm leading-relaxed text-[#9a9a9a]">
							{role.description || "No description loaded"}
						</p>
					</div>
				</div>
			</div>
		</aside>
	);
}

function MatchAnalysisCard({ role }: { role: RoleDetailResponse }) {
	const d = role.match_details;
	if (!d) return null;
	return (
		<div className="rounded-xl border border-[#1a2a1a] bg-[#0d120d] p-4">
			<p className="mb-3 text-xs font-medium text-[#6a7a6a]">Match Analysis</p>

			<div className="space-y-1.5 mb-3">
				{role.match_reasons?.map((r: string) => {
					const [source, ...rest] = r.split(":");
					const kw = rest.join(":");
					let Icon = FileText,
						label = "",
						color = "text-[#7dba7a]",
						points = "";
					switch (source) {
						case "title":
							Icon = User;
							label = "Title";
							color = "text-orange-400";
							points = "+2";
							break;
						case "dept":
							Icon = FolderTree;
							label = "Department";
							color = "text-amber-400";
							points = "+1";
							break;
						case "loc":
							Icon = MapPin;
							label = "Loc";
							color = "text-amber-400";
							points = "+1";
							break;
						case "desc":
							Icon = FileText;
							label = "Description";
							color = "text-emerald-400";
							points = "+1";
							break;
						case "location":
							Icon = Globe;
							label = "Region";
							color = "text-blue-400";
							points = "✓";
							break;
						case "work_type":
							Icon = Zap;
							label = "Type";
							color = "text-purple-400";
							points = "+1";
							break;
						default:
							Icon = FileText;
							label = "";
							color = "text-[#6a7a6a]";
							points = "";
							break;
					}
					return (
						<div key={r} className="flex items-center justify-between text-sm">
							<div className="flex items-center gap-1.5">
								<Icon size={11} className={color} />
								<span className="text-[#c8c8c8]">
									{kw} <span className="text-[#6a7a6a]">({label})</span>
								</span>
							</div>
							{points && (
								<span
									className={`font-medium ${source === "location" ? "text-blue-400" : color}`}
								>
									{points}
								</span>
							)}
						</div>
					);
				})}
			</div>

			<div className="space-y-1 text-sm text-[#9a9a9a]">
				<div className="flex justify-between border-t border-[#1a2a1a] pt-2">
					<span>Include score</span>
					<span className="text-[#c8c8c8] font-medium">
						{d.include_score} pts
					</span>
				</div>
				{d.bonus_score > 0 && (
					<div className="flex justify-between">
						<span>Bonus (work type)</span>
						<span className="text-purple-400 font-medium">
							+{d.bonus_score}
						</span>
					</div>
				)}
				<div className="flex justify-between border-t border-[#1a2a1a] pt-2">
					<span>Raw score</span>
					<span className="text-[#c8c8c8] font-medium">
						{d.total_score} pts
					</span>
				</div>
				<div className="flex justify-between">
					<span>Recency</span>
					<span className="text-blue-400 font-medium">
						×{d.recency_factor.toFixed(2)}
					</span>
				</div>
				<div className="flex justify-between border-t border-[#1a2a1a] pt-2">
					<span>Adjusted score</span>
					<span className="text-[#c8c8c8] font-medium">
						{d.adjusted_score.toFixed(1)}
					</span>
				</div>
			</div>

			<div className="mt-3 flex items-center gap-3 rounded-lg bg-[#080908] px-3 py-2">
				<span className={`text-sm font-bold ${scoreColor(d.percent)}`}>
					{matchLabel(d.percent)}
				</span>
				<div className="flex-1 h-1.5 rounded-full bg-[#1a2a1a] overflow-hidden">
					<div
						className={`h-full rounded-full transition-all ${d.percent >= 70 ? "bg-[#7dba7a]" : d.percent >= 40 ? "bg-amber-400" : "bg-[#4a5a4a]"}`}
						style={{ width: `${d.percent}%` }}
					/>
				</div>
				<span className="text-sm text-[#6a7a6a]">
					{d.matched_keywords}/{d.total_keywords} keywords
				</span>
			</div>
		</div>
	);
}
