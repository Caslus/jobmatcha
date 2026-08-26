import { LogOut, Settings } from "lucide-react";
import { useState } from "react";
import { useLogout } from "#/hooks/useApi.ts";
import wordmark from "@/assets/wordmark.svg";
import { SettingsPanel } from "@/features/settings/SettingsPanel.tsx";

export default function Header() {
	const [showSettings, setShowSettings] = useState(false);
	const logout = useLogout();

	return (
		<>
			<header className="flex items-center justify-between border-b border-[#1a2a1a] px-6 py-3">
				<div className="flex items-center gap-3">
					<img
						src={wordmark}
						alt="jobmatcha"
						draggable="false"
						className="mx-auto h-10 select-none"
					/>
				</div>
				<div className="flex items-center gap-2">
					<button
						type="button"
						onClick={() => setShowSettings(true)}
						className="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs text-[#6a7a6a] transition hover:bg-[#1a2a1a] hover:text-[#e8e8e8]"
						title="Settings"
					>
						<Settings size={14} />
					</button>
					<button
						type="button"
						onClick={() => logout.mutate()}
						className="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs text-[#6a7a6a] transition hover:bg-[#1a2a1a] hover:text-[#e8e8e8]"
					>
						<LogOut size={14} /> Logout
					</button>
				</div>
			</header>
			{showSettings && <SettingsPanel onClose={() => setShowSettings(false)} />}
		</>
	);
}
