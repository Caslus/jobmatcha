import { motion } from "motion/react";

interface Props {
	label?: string;
}

export function LoadingAnimation({ label }: Props) {
	return (
		<div className="flex flex-col items-center justify-center gap-3 py-8">
			<div className="flex gap-1.5">
				{[0, 1, 2].map((i) => (
					<motion.div
						key={i}
						className="h-3 w-3 rounded-full bg-[#7dba7a]"
						animate={{
							scale: [1, 1.5, 1],
							opacity: [0.4, 1, 0.4],
						}}
						transition={{
							duration: 1.2,
							repeat: Number.POSITIVE_INFINITY,
							delay: i * 0.2,
							ease: "easeInOut",
						}}
					/>
				))}
			</div>
			{label && <p className="text-xs font-medium text-[#6a7a6a]">{label}</p>}
		</div>
	);
}
