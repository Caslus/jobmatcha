import { useMemo, useRef, useState } from "react";
import {
	type KeywordField,
	KeywordSection,
	WorkTypeToggle,
} from "@/components/keywords/KeywordSection";

interface ManualData {
	name: string;
	email: string;
	location: string;
	linkedin_url: string;
	github_url: string;
	include_keywords: string[];
	exclude_keywords: string[];
	work_types: string[];
}

interface Props {
	data: ManualData;
	onChange: (data: ManualData) => void;
}

const WORK_TYPE_OPTIONS = ["internship", "contract", "part-time", "full-time"];

export function ReviewStep({ data, onChange }: Props) {
	const [showInputs, setShowInputs] = useState<Set<string>>(new Set());
	const [markedForDelete, setMarkedForDelete] = useState<Set<string>>(
		new Set(),
	);
	const includeRef = useRef<HTMLInputElement>(null);
	const excludeRef = useRef<HTMLInputElement>(null);

	const inputRefs = useMemo(
		() => ({
			include: includeRef,
			exclude: excludeRef,
		}),
		[],
	);

	const set =
		(field: keyof ManualData) => (e: React.ChangeEvent<HTMLInputElement>) => {
			onChange({ ...data, [field]: e.target.value });
		};

	const stageAdd = (field: KeywordField, rawValue: string) => {
		const keywords = rawValue
			.split(",")
			.map((s) => s.trim().toLowerCase())
			.filter((s) => s.length > 0);
		if (keywords.length === 0) return;
		const target =
			field === "include" ? "include_keywords" : "exclude_keywords";
		const arr = [...data[target]];
		for (const kw of keywords) {
			if (!arr.includes(kw)) arr.push(kw);
		}
		onChange({ ...data, [target]: arr });
	};

	const handleChipClick = (kw: string, field: KeywordField) => {
		setMarkedForDelete((prev) => {
			const next = new Set(prev);
			if (next.has(kw)) next.delete(kw);
			else next.add(kw);
			return next;
		});
		const target =
			field === "include" ? "include_keywords" : "exclude_keywords";
		onChange({ ...data, [target]: data[target].filter((k) => k !== kw) });
	};

	return (
		<div className="space-y-4">
			{/* Profile fields */}
			<div className="grid grid-cols-2 gap-3">
				<div>
					<label
						htmlFor="review-name"
						className="mb-1 block text-xs font-medium text-[#6a7a6a]"
					>
						Name
					</label>
					<input
						id="review-name"
						type="text"
						value={data.name}
						onChange={set("name")}
						className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-3 py-2 text-xs text-[#e8e8e8] outline-none transition focus:border-[#7dba7a]"
					/>
				</div>
				<div>
					<label
						htmlFor="review-email"
						className="mb-1 block text-xs font-medium text-[#6a7a6a]"
					>
						Email
					</label>
					<input
						id="review-email"
						type="email"
						value={data.email}
						onChange={set("email")}
						className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-3 py-2 text-xs text-[#e8e8e8] outline-none transition focus:border-[#7dba7a]"
					/>
				</div>
				<div>
					<label
						htmlFor="review-location"
						className="mb-1 block text-xs font-medium text-[#6a7a6a]"
					>
						Location
					</label>
					<input
						id="review-location"
						type="text"
						value={data.location}
						onChange={set("location")}
						className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-3 py-2 text-xs text-[#e8e8e8] outline-none transition focus:border-[#7dba7a]"
					/>
				</div>
				<div>
					<label
						htmlFor="review-linkedin"
						className="mb-1 block text-xs font-medium text-[#6a7a6a]"
					>
						LinkedIn
					</label>
					<input
						id="review-linkedin"
						type="text"
						value={data.linkedin_url}
						onChange={set("linkedin_url")}
						placeholder="https://linkedin.com/in/..."
						className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-3 py-2 text-xs text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[#7dba7a]"
					/>
				</div>
				<div>
					<label
						htmlFor="review-github"
						className="mb-1 block text-xs font-medium text-[#6a7a6a]"
					>
						GitHub
					</label>
					<input
						id="review-github"
						type="text"
						value={data.github_url}
						onChange={set("github_url")}
						placeholder="https://github.com/..."
						className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-3 py-2 text-xs text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[#7dba7a]"
					/>
				</div>
			</div>

			<hr className="border-[#1a2a1a]" />

			{/* Include keywords */}
			<KeywordSection
				label="Keywords"
				field="include"
				keywords={data.include_keywords}
				showInput={showInputs.has("include")}
				onAddInput={() =>
					setShowInputs((prev) => {
						const n = new Set(prev);
						n.add("include");
						return n;
					})
				}
				onCloseInput={() =>
					setShowInputs((prev) => {
						const n = new Set(prev);
						n.delete("include");
						return n;
					})
				}
				inputRef={inputRefs.include}
				highlight="green"
				onStageAdd={stageAdd}
				onChipClick={handleChipClick}
				markedForDelete={markedForDelete}
			/>

			{/* Exclude keywords */}
			<KeywordSection
				label="Exclude"
				field="exclude"
				keywords={data.exclude_keywords}
				showInput={showInputs.has("exclude")}
				onAddInput={() =>
					setShowInputs((prev) => {
						const n = new Set(prev);
						n.add("exclude");
						return n;
					})
				}
				onCloseInput={() =>
					setShowInputs((prev) => {
						const n = new Set(prev);
						n.delete("exclude");
						return n;
					})
				}
				inputRef={inputRefs.exclude}
				highlight="red"
				onStageAdd={stageAdd}
				onChipClick={handleChipClick}
				markedForDelete={markedForDelete}
			/>

			{/* Work types */}
			<div>
				<span className="mb-1 block text-xs font-medium text-[#6a7a6a]">
					Work Types
				</span>
				<div className="flex flex-wrap gap-1.5">
					<WorkTypeToggle
						label="Any"
						selected={data.work_types.length === 0}
						onToggle={() => onChange({ ...data, work_types: [] })}
					/>
					{WORK_TYPE_OPTIONS.map((t) => (
						<WorkTypeToggle
							key={t}
							label={
								t === "full-time"
									? "Full-time"
									: t.charAt(0).toUpperCase() + t.slice(1)
							}
							selected={data.work_types.includes(t)}
							onToggle={() => {
								const next = data.work_types.includes(t)
									? data.work_types.filter((x) => x !== t)
									: [...data.work_types, t];
								onChange({ ...data, work_types: next });
							}}
						/>
					))}
				</div>
			</div>
		</div>
	);
}
