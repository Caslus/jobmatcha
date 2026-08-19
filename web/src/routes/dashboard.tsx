import { createFileRoute, redirect, useNavigate } from "@tanstack/react-router";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import Header from "#/components/Header.tsx";
import { RoleDetailPanel } from "../features/jobs/RoleDetailPanel";
import { RoleList } from "../features/jobs/RoleList";
import { PreferencesPanel } from "../features/preferences";
import { useAuthStore } from "../stores/auth";

// ── Resizable Panel Layout ─────────────────────────

const STORAGE_KEY = "jobmatcha_panel_widths";
const MIN_PCT = 10;
const COLLAPSE_THRESHOLD = 8;
const MAX_PCT = 80;
const DEFAULT_LEFT = 20;
const DEFAULT_RIGHT = 40;

interface PanelState {
	leftPct: number;
	rightPct: number;
	collapsedLeft: boolean;
	collapsedRight: boolean;
}

function loadState(): PanelState {
	try {
		const saved = localStorage.getItem(STORAGE_KEY);
		if (saved) {
			const data = JSON.parse(saved);
			const left = Math.max(0, Math.min(MAX_PCT, data.left ?? DEFAULT_LEFT));
			const right = Math.max(0, Math.min(MAX_PCT, data.right ?? DEFAULT_RIGHT));
			return {
				leftPct: left || DEFAULT_LEFT,
				rightPct: right || DEFAULT_RIGHT,
				collapsedLeft: left === 0,
				collapsedRight: right === 0,
			};
		}
	} catch {
		/* ignore */
	}
	return {
		leftPct: DEFAULT_LEFT,
		rightPct: DEFAULT_RIGHT,
		collapsedLeft: false,
		collapsedRight: false,
	};
}

function saveState(left: number, right: number) {
	try {
		localStorage.setItem(STORAGE_KEY, JSON.stringify({ left, right }));
	} catch {
		/* ignore */
	}
}

function ResizeHandle({
	onMouseDown,
}: {
	side: "left" | "right";
	onMouseDown: (e: React.MouseEvent) => void;
}) {
	return (
		// biome-ignore lint/a11y/useSemanticElements: Cannot use <hr> because a layout splitter requires child elements for the visual handle
		<div
			className="w-px shrink-0 relative cursor-col-resize bg-[#1a2a1a] hover:bg-[#7dba7a]/30 transition-colors self-stretch"
			onMouseDown={onMouseDown}
			tabIndex={-1}
			role="separator"
			aria-valuenow={0}
		>
			<div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-1 h-6 bg-[#4a5a4a]" />
		</div>
	);
}

function CollapseButton({
	side,
	onRestore,
}: {
	side: "left" | "right";
	onRestore: () => void;
}) {
	const isRight = side === "right";
	return (
		<button
			type="button"
			onClick={onRestore}
			className="w-5 shrink-0 flex items-center justify-center self-stretch bg-[#1a2a1a] hover:bg-[#7dba7a]/30 transition-colors cursor-pointer"
			title={isRight ? "Show details" : "Show preferences"}
		>
			{isRight ? (
				<ChevronLeft size={12} className="text-[#6a7a6a]" />
			) : (
				<ChevronRight size={12} className="text-[#6a7a6a]" />
			)}
		</button>
	);
}

// ── Route Definition ───────────────────────────────

export const Route = createFileRoute("/dashboard")({
	beforeLoad: () => {
		const { authenticated } = useAuthStore.getState();
		if (!authenticated) throw redirect({ to: "/" });
	},
	component: DashboardPage,
});

// ── Dashboard Page ─────────────────────────────────

function DashboardPage() {
	const navigate = useNavigate();
	const { check } = useAuthStore();
	const [selectedId, setSelectedId] = useState<number | null>(null);
	const containerRef = useRef<HTMLDivElement>(null);
	const init = loadState();
	const [leftPct, setLeftPct] = useState(init.leftPct);
	const [rightPct, setRightPct] = useState(init.rightPct);
	const [collapsedLeft, setCollapsedLeft] = useState(init.collapsedLeft);
	const [collapsedRight, setCollapsedRight] = useState(init.collapsedRight);
	const dragging = useRef<"left" | "right" | null>(null);

	useEffect(() => {
		check().then(() => {
			const { setupComplete } = useAuthStore.getState();
			if (!setupComplete) navigate({ to: "/onboarding" });
		});
	}, [check, navigate]);

	const onMouseDown = useCallback(
		(side: "left" | "right") => (e: React.MouseEvent) => {
			e.preventDefault();
			dragging.current = side;
			document.body.style.cursor = "col-resize";
			document.body.style.userSelect = "none";
		},
		[],
	);

	useEffect(() => {
		const onMove = (e: MouseEvent) => {
			if (!dragging.current || !containerRef.current) return;
			const rect = containerRef.current.getBoundingClientRect();
			const pct = ((e.clientX - rect.left) / rect.width) * 100;
			if (dragging.current === "left") {
				const maxLeft = 100 - rightPct - MIN_PCT;
				if (pct < COLLAPSE_THRESHOLD) {
					setCollapsedLeft(true);
					setLeftPct(0);
				} else {
					setCollapsedLeft(false);
					setLeftPct(Math.max(MIN_PCT, Math.min(maxLeft, pct)));
				}
			} else {
				const rightEdge = 100 - pct;
				const maxRight = 100 - leftPct - MIN_PCT;
				if (rightEdge < COLLAPSE_THRESHOLD) {
					setCollapsedRight(true);
					setRightPct(0);
				} else {
					setCollapsedRight(false);
					setRightPct(Math.max(MIN_PCT, Math.min(maxRight, rightEdge)));
				}
			}
		};
		const onUp = () => {
			if (dragging.current) {
				dragging.current = null;
				document.body.style.cursor = "";
				document.body.style.userSelect = "";
				saveState(leftPct, rightPct);
			}
		};
		window.addEventListener("mousemove", onMove);
		window.addEventListener("mouseup", onUp);
		return () => {
			window.removeEventListener("mousemove", onMove);
			window.removeEventListener("mouseup", onUp);
		};
	}, [leftPct, rightPct]);

	const restoreLeft = () => {
		setCollapsedLeft(false);
		setLeftPct(DEFAULT_LEFT);
		saveState(DEFAULT_LEFT, rightPct);
	};

	const restoreRight = () => {
		setCollapsedRight(false);
		setRightPct(DEFAULT_RIGHT);
		saveState(leftPct, DEFAULT_RIGHT);
	};

	return (
		<div className="flex h-screen flex-col text-[#e8e8e8]">
			<Header />

			<div ref={containerRef} className="flex flex-1 overflow-hidden">
				{!collapsedLeft && (
					<div style={{ width: `${leftPct}%` }} className="shrink-0 min-w-0">
						<PreferencesPanel />
					</div>
				)}

				{collapsedLeft ? (
					<CollapseButton side="left" onRestore={restoreLeft} />
				) : (
					<ResizeHandle side="left" onMouseDown={onMouseDown("left")} />
				)}

				<RoleList selectedId={selectedId} onSelect={setSelectedId} />

				{collapsedRight ? (
					<CollapseButton side="right" onRestore={restoreRight} />
				) : (
					<ResizeHandle side="right" onMouseDown={onMouseDown("right")} />
				)}

				{!collapsedRight && (
					<div style={{ width: `${rightPct}%` }} className="shrink-0 min-w-0">
						<RoleDetailPanel
							selectedId={selectedId}
							onBack={() => setSelectedId(null)}
						/>
					</div>
				)}
			</div>
		</div>
	);
}
