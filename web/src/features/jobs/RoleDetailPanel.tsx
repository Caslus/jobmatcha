import {
	Bookmark,
	Calendar,
	ExternalLink,
	FileText,
	FolderTree,
	Globe,
	MapPin,
	User,
	WandSparkles,
	X,
	Zap,
} from "lucide-react";
import { useState } from "react";
import { Tooltip } from "../../components/Tooltip";
import {
	useAISettings,
	usePatchRole,
	useRole,
	useTailoredResume,
	useTailorResume,
} from "../../hooks/useApi";
import { matchLabel, scoreColor, timeAgo } from "../../lib/dashboard";
import { formatDate } from "../../lib/date";
import type { DescriptionFormat } from "../../lib/description";
import type { ResumeDocument, RoleDetailResponse } from "../../types/api.gen";
import { TailoredResumePreview } from "../resumes/TailoredResumePreview";
import { JobDescription } from "./JobDescription";

interface RoleDetailPanelProps {
	selectedId: number | null;
	onBack: () => void;
}

export function RoleDetailPanel({ selectedId, onBack }: RoleDetailPanelProps) {
	const { data: role } = useRole(selectedId);
	const { data: aiSettings } = useAISettings();
	const patchRole = usePatchRole();
	const tailoredResume = useTailoredResume(selectedId);
	const tailorResume = useTailorResume();
	const [previewOpen, setPreviewOpen] = useState(false);
	const [previewDocument, setPreviewDocument] = useState<ResumeDocument | null>(
		null,
	);

	if (!role) {
		return (
			<aside className="w-full h-full overflow-y-auto">
				<div className="flex h-full items-center justify-center text-[#4a5a4a]">
					<p className="text-sm">Select a role to view details</p>
				</div>
			</aside>
		);
	}

	const handleToggleBookmark = () => {
		patchRole.mutate({ id: role.id, is_interested: !role.is_interested });
	};

	const handleTailor = () => {
		tailorResume.mutate(role.id, {
			onSuccess: (tailored) => {
				setPreviewDocument(tailored.document);
				setPreviewOpen(true);
			},
		});
	};

	const displayedResume = previewDocument ?? tailoredResume.data?.document;
	const canTailor = aiSettings?.enabled === true;
	const profileLinks = [
		aiSettings?.user_linkedin,
		aiSettings?.user_github,
	].filter((link): link is string => Boolean(link?.trim()));

	return (
		<aside className="w-full h-full overflow-y-auto">
			<div className="p-6">
				<div className="mb-6">
					<div className="flex items-center justify-between gap-3">
						<div>
							<p className="text-sm text-[#6a7a6a]">{role.company_name}</p>
							<h3 className="mt-1 text-lg font-bold text-[#e8e8e8]">
								{role.title}
							</h3>
						</div>
						<button
							type="button"
							onClick={onBack}
							className="rounded-lg p-1 text-[#6a7a6a] transition hover:bg-[#1a2a1a] hover:text-[#e8e8e8]"
						>
							<X size={16} />
						</button>
					</div>
					<div className="mt-3 flex items-start gap-3 text-sm text-[#6a7a6a]">
						<div className="flex flex-col gap-1">
							<span className="flex items-center gap-1">
								<Globe size={12} />
								{role.location}
							</span>
							{role.department && (
								<span className="flex items-center gap-1">
									<FolderTree size={12} />
									{role.department}
								</span>
							)}
							{role.posted_at && (
								<span className="flex items-center gap-1">
									<Calendar size={12} />
									<Tooltip content={formatDate(role.posted_at)}>
										<span className="border-b border-dotted border-[#4a5a4a]">
											{timeAgo(role.posted_at)}
										</span>
									</Tooltip>
								</span>
							)}
						</div>
					</div>
				</div>

				<div className="space-y-4">
					<div className="grid gap-2 sm:grid-cols-2">
						<button
							type="button"
							onClick={handleToggleBookmark}
							disabled={patchRole.isPending}
							className={`flex flex-1 items-center justify-center gap-2 rounded-lg border px-4 py-2 text-sm font-medium transition ${
								role.is_interested
									? "border-[#7dba7a] bg-[#7dba7a]/10 text-[#7dba7a]"
									: "border-[#2a3a2a] text-[#6a7a6a] hover:border-[#7dba7a]/50 hover:text-[#7dba7a]"
							}`}
						>
							{patchRole.isPending ? (
								"Updating..."
							) : (
								<>
									<Bookmark
										size={14}
										className={role.is_interested ? "fill-[#7dba7a]" : ""}
									/>
									{role.is_interested ? "Bookmarked" : "Bookmark"}
								</>
							)}
						</button>
						{role.url && (
							<a
								href={role.url}
								target="_blank"
								rel="noopener noreferrer"
								className="flex flex-1 items-center justify-center gap-2 rounded-lg border border-[#2a3a2a] bg-[#0d120d] px-4 py-2 text-sm text-[#9a9a9a] transition hover:border-[#7dba7a]/50 hover:text-[#e8e8e8]"
							>
								<ExternalLink size={14} />
								<span>Visit Website</span>
							</a>
						)}
					</div>
					<button
						type="button"
						onClick={handleTailor}
						disabled={!canTailor || tailorResume.isPending}
						title={
							canTailor
								? undefined
								: "Enable an AI provider in Settings to tailor resumes"
						}
						className="flex w-full items-center justify-center gap-2 rounded-lg bg-linear-to-r from-[#7dba7a] to-[#5a8f5a] px-4 py-2.5 text-sm font-semibold text-[#080908] transition hover:from-[#8dca8a] hover:to-[#6a9f6a] disabled:cursor-not-allowed disabled:opacity-60"
					>
						<WandSparkles size={15} />
						{tailorResume.isPending
							? "Tailoring resume..."
							: canTailor
								? "Tailor resume"
								: "AI provider disabled"}
					</button>
					{tailorResume.isError && (
						<div className="rounded-lg border border-red-400/20 bg-red-400/5 px-3 py-2 text-xs text-red-400/90">
							{tailorResume.error instanceof Error
								? tailorResume.error.message
								: "Could not tailor your resume. Please try again."}
						</div>
					)}
					{tailoredResume.data && (
						<button
							type="button"
							onClick={() => {
								setPreviewDocument(tailoredResume.data?.document ?? null);
								setPreviewOpen(true);
							}}
							className="flex w-full items-center justify-center gap-2 rounded-lg border border-[#335533] bg-[#0a0f0a] px-4 py-2 text-sm font-medium text-[#9bd398] transition hover:bg-[#173017]"
						>
							View tailored resume
						</button>
					)}
					{displayedResume && (
						<TailoredResumePreview
							open={previewOpen}
							document={displayedResume}
							jobTitle={role.title}
							profileLinks={profileLinks}
							onClose={() => {
								setPreviewOpen(false);
								setPreviewDocument(null);
							}}
						/>
					)}

					{role.match_details && <MatchAnalysisCard role={role} />}

					<div className="rounded-xl border border-[#1a2a1a] bg-[#0d120d] p-4">
						<p className="mb-2 text-xs font-medium text-[#6a7a6a]">
							Description
						</p>
						<JobDescription
							description={role.description || ""}
							format={(role.description_format as DescriptionFormat) || "plain"}
						/>
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
			<div className="flex items-center justify-between mb-3 w-full">
				<p className="text-xs font-medium text-[#6a7a6a]">Match Analysis</p>
				<span className="text-xs text-[#6a7a6a]">
					{d.matched_keywords}/{d.total_keywords} keywords
				</span>
			</div>

			<div className="relative flex items-center justify-center gap-3">
				<div className="h-px flex-1 bg-linear-to-r from-transparent via-[#2a3a2a] to-[#2a3a2a]" />
				<span
					className={`inline-flex items-center gap-1.5 rounded-md border px-3 py-1 text-sm font-semibold 
					${scoreColor(role.match_percent)}`}
				>
					✦ Match: {matchLabel(role.match_percent)}
				</span>
				<div className="h-px flex-1 bg-linear-to-r from-[#2a3a2a] via-[#2a3a2a] to-transparent" />
			</div>

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

			<div className="mt-3 flex items-center gap-3 rounded-lg bg-[#080908] border border-[#1a2a1a] px-3 py-2">
				<span
					className={`text-sm font-bold rounded-md p-1 ${scoreColor(d.percent)}`}
				>
					{matchLabel(d.percent)}
				</span>
				<div className="flex-1 h-1.5 rounded-full bg-[#1a2a1a] overflow-hidden">
					<div
						className={`h-full rounded-full transition-all ${d.percent === 100 ? "bg-purple-400" : d.percent >= 70 ? "bg-[#7dba7a]" : d.percent >= 40 ? "bg-amber-400" : "bg-[#4a5a4a]"}`}
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
