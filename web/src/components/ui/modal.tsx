import { X } from "lucide-react";
import type { ReactNode } from "react";
import { useEffect, useId, useRef } from "react";
import { cn } from "#/lib/utils.ts";

interface ModalProps {
	open: boolean;
	onClose: () => void;
	title: ReactNode;
	subtitle?: ReactNode;
	icon?: ReactNode;
	headerActions?: ReactNode;
	children: ReactNode;
	panelClassName?: string;
	headerClassName?: string;
	closeLabel?: string;
}

/**
 * Shared native-dialog shell for app overlays. Feature components own their
 * content while the modal consistently handles focus, Escape, and backdrop
 * dismissal.
 */
export function Modal({
	open,
	onClose,
	title,
	subtitle,
	icon,
	headerActions,
	children,
	panelClassName,
	headerClassName,
	closeLabel = "Close dialog",
}: ModalProps) {
	const dialogRef = useRef<HTMLDialogElement>(null);
	const titleId = useId();

	useEffect(() => {
		const dialog = dialogRef.current;
		if (!dialog) return;

		if (open && !dialog.open) dialog.showModal();
		if (!open && dialog.open) dialog.close();

		return () => {
			if (dialog.open) dialog.close();
		};
	}, [open]);

	return (
		<dialog
			ref={dialogRef}
			onClick={(event) => {
				if (event.target === dialogRef.current) onClose();
			}}
			onCancel={(event) => {
				event.preventDefault();
				onClose();
			}}
			onKeyDown={(event) => {
				if (event.key !== "Escape") return;
				event.preventDefault();
				onClose();
			}}
			aria-labelledby={titleId}
			className="fixed inset-0 z-50 m-0 hidden h-full w-full items-center justify-center border-0 bg-transparent p-4 open:flex"
		>
			<style>{`dialog::backdrop { background: rgba(0, 0, 0, 0.5); }`}</style>
			<div
				className={cn(
					"flex max-h-[92vh] w-full flex-col overflow-hidden rounded-xl border border-[#1a2a1a] bg-[#0d120d] shadow-2xl",
					panelClassName,
				)}
			>
				<header
					className={cn(
						"flex items-center justify-between gap-4 border-b border-[#1a2a1a] px-5 py-3",
						headerClassName,
					)}
				>
					<div className="min-w-0">
						<div className="flex items-center gap-2">
							{icon}
							<h2 id={titleId} className="text-sm font-semibold text-[#e8e8e8]">
								{title}
							</h2>
						</div>
						{subtitle && (
							<p className="mt-0.5 truncate text-xs text-[#6a7a6a]">
								{subtitle}
							</p>
						)}
					</div>
					<div className="flex shrink-0 items-center gap-2">
						{headerActions}
						<button
							type="button"
							onClick={onClose}
							className="rounded-md p-2 text-[#4a5a4a] transition hover:bg-[#1a2a1a] hover:text-[#6a7a6a]"
							aria-label={closeLabel}
						>
							<X size={16} />
						</button>
					</div>
				</header>
				{children}
			</div>
		</dialog>
	);
}
