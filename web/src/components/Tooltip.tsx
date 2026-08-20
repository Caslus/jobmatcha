import { type ReactNode, useRef, useState } from "react";
import { createPortal } from "react-dom";

interface TooltipProps {
	content: ReactNode;
	children: ReactNode;
}

export function Tooltip({ content, children }: TooltipProps) {
	const [show, setShow] = useState(false);
	const [pos, setPos] = useState({ top: 0, left: 0 });
	const ref = useRef<HTMLDivElement>(null);

	const handleMouseEnter = () => {
		if (ref.current) {
			const rect = ref.current.getBoundingClientRect();
			setPos({ top: rect.top - 8, left: rect.left + rect.width / 2 });
		}
		setShow(true);
	};

	const handleMouseLeave = () => setShow(false);

	return (
		<>
			{/* biome-ignore lint/a11y/noStaticElementInteractions: hover-only tooltip trigger, not a clickable control */}
			<div
				ref={ref}
				onMouseEnter={handleMouseEnter}
				onMouseLeave={handleMouseLeave}
				className="cursor-help"
			>
				{children}
			</div>
			{show &&
				createPortal(
					<div
						className="pointer-events-none fixed z-9999 -translate-x-1/2 -translate-y-full whitespace-nowrap rounded-lg border border-[#2a3a2a] bg-[#1a2a1a] px-3 py-1.5 text-xs text-[#c8c8c8] shadow-lg"
						style={{ top: pos.top, left: pos.left }}
					>
						{content}
					</div>,
					document.body,
				)}
		</>
	);
}
