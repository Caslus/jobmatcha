import { motion } from "motion/react";

interface Step {
	id: string;
	title: string;
	description: string;
}

interface Props {
	steps: Step[];
	current: number;
	children: React.ReactNode;
	canGoBack: boolean;
	canGoNext: boolean;
	onBack: () => void;
	onNext: () => void;
	isLast?: boolean;
	onFinish?: () => void;
	isSubmitting?: boolean;
	finishLabel?: string;
}

export function StepManager({
	steps,
	current,
	children,
	canGoBack,
	canGoNext,
	onBack,
	onNext,
	isLast,
	onFinish,
	isSubmitting,
	finishLabel = "Launch",
}: Props) {
	const step = steps[current];
	const progress = ((current + 1) / steps.length) * 100;

	return (
		<div className="w-full max-w-md mx-auto">
			{/* Progress bar */}
			<div className="mb-6">
				<div className="flex justify-between mb-2">
					{steps.map((s, i) => (
						<button
							key={s.id}
							type="button"
							disabled
							className={`flex h-8 w-8 items-center justify-center rounded-full text-xs font-bold transition ${
								i <= current
									? "bg-[#7dba7a] text-[#080908]"
									: "bg-[#1a2a1a] text-[#4a5a4a]"
							}`}
						>
							{i + 1}
						</button>
					))}
				</div>
				<div className="h-1 rounded-full bg-[#1a2a1a]">
					<div
						className="h-full rounded-full bg-[#7dba7a] transition-all duration-500"
						style={{ width: `${progress}%` }}
					/>
				</div>
			</div>

			{/* Step header */}
			<div className="mb-6 text-center">
				<h2 className="text-lg font-bold text-[#e8e8e8]">{step.title}</h2>
				<p className="mt-1 text-sm text-[#6a7a6a]">{step.description}</p>
			</div>

			{/* Step content */}
			<motion.div
				key={current}
				initial={{ opacity: 0, x: 20 }}
				animate={{ opacity: 1, x: 0 }}
				exit={{ opacity: 0, x: -20 }}
				transition={{ duration: 0.2 }}
			>
				{children}
			</motion.div>

			{/* Navigation */}
			<div className="mt-8 flex gap-3">
				{current > 0 ? (
					<button
						type="button"
						onClick={onBack}
						disabled={!canGoBack}
						className="flex-1 rounded-lg border border-[#2a3a2a] px-4 py-2.5 text-sm font-medium text-[#6a7a6a] transition hover:bg-[#1a2a1a] disabled:cursor-not-allowed disabled:opacity-40"
					>
						Back
					</button>
				) : (
					<div className="flex-1" />
				)}

				{isLast ? (
					<button
						type="button"
						onClick={onFinish}
						disabled={!canGoNext || isSubmitting}
						className="flex-1 rounded-lg bg-linear-to-r from-[#7dba7a] to-[#5a8f5a] px-4 py-2.5 text-sm font-semibold text-[#080908] transition hover:from-[#8dca8a] hover:to-[#6a9f6a] disabled:cursor-not-allowed disabled:opacity-50"
					>
						{isSubmitting ? "Starting..." : finishLabel}
					</button>
				) : (
					<button
						type="button"
						onClick={onNext}
						disabled={!canGoNext}
						className="flex-1 rounded-lg bg-linear-to-r from-[#7dba7a] to-[#5a8f5a] px-4 py-2.5 text-sm font-semibold text-[#080908] transition hover:from-[#8dca8a] hover:to-[#6a9f6a] disabled:cursor-not-allowed disabled:opacity-50"
					>
						Continue
					</button>
				)}
			</div>
		</div>
	);
}
