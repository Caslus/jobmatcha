import { FileText, Upload } from "lucide-react";
import { useRef, useState } from "react";
import type { ParseResumeResponse } from "@/types/api.gen";
import { useParseResume } from "../../hooks/useApi";
import { LoadingAnimation } from "./LoadingAnimation";

interface Props {
	onData: (data: ParseResumeResponse) => void;
	onSkip: () => void;
}

export function ResumeUploadStep({ onData, onSkip }: Props) {
	const parseResume = useParseResume();
	const [dragOver, setDragOver] = useState(false);
	const [file, setFile] = useState<File | null>(null);
	const inputRef = useRef<HTMLInputElement>(null);

	const handleFile = (f: File) => {
		const ext = f.name.split(".").pop()?.toLowerCase();
		if (!ext || !["md", "txt", "pdf"].includes(ext)) return;
		setFile(f);
		parseResume.mutate(f, {
			onSuccess: (data) => onData(data),
		});
	};

	return (
		<div className="space-y-4">
			{/* Drop zone */}
			<button
				type="button"
				onDragOver={(e) => {
					e.preventDefault();
					setDragOver(true);
				}}
				onDragLeave={() => setDragOver(false)}
				onDrop={(e) => {
					e.preventDefault();
					setDragOver(false);
					const f = e.dataTransfer.files[0];
					if (f) handleFile(f);
				}}
				onClick={() => inputRef.current?.click()}
				className={`flex w-full cursor-pointer flex-col items-center gap-3 rounded-xl border-2 border-dashed p-8 transition ${
					dragOver
						? "border-[#7dba7a] bg-[rgba(125,186,122,0.05)]"
						: "border-[#2a3a2a] hover:border-[#4a5a4a]"
				}`}
			>
				<Upload
					size={28}
					className={dragOver ? "text-[#7dba7a]" : "text-[#4a5a4a]"}
				/>
				<div className="text-center">
					<p className="text-sm font-medium text-[#e8e8e8]">
						{file ? file.name : "Drop your resume here"}
					</p>
					<p className="mt-1 text-xs text-[#6a7a6a]">
						{file
							? `${(file.size / 1024).toFixed(0)} KB`
							: "Supports PDF, Markdown (.md), or text files"}
					</p>
				</div>
				<input
					ref={inputRef}
					type="file"
					accept=".md,.txt,.pdf"
					className="hidden"
					onChange={(e) => {
						const f = e.target.files?.[0];
						if (f) handleFile(f);
					}}
				/>
			</button>

			{/* File type icon */}
			{file && !parseResume.isPending && !parseResume.isError && (
				<div className="flex items-center gap-2 rounded-lg border border-[#1a2a1a] bg-[#080b08] px-3 py-2">
					<FileText size={14} className="text-[#7dba7a]" />
					<span className="text-xs text-[#e8e8e8]">{file.name}</span>
				</div>
			)}

			{/* Loading */}
			{parseResume.isPending && (
				<LoadingAnimation label="Analyzing resume..." />
			)}

			{/* Error */}
			{parseResume.isError && (
				<div className="rounded-lg border border-red-400/20 bg-red-400/5 px-3 py-2">
					<p className="text-xs text-red-400/90">
						{parseResume.error instanceof Error
							? parseResume.error.message
							: "Failed to parse resume"}
					</p>
				</div>
			)}

			{/* Manual entry button */}
			<button
				type="button"
				onClick={onSkip}
				className="text-xs text-[#4a5a4a] underline transition hover:text-[#6a7a6a]"
			>
				Enter details manually instead
			</button>
		</div>
	);
}
