import { useState } from "react";

interface Props {
	onChange: (data: { currentPassword: string; newPassword: string }) => void;
	onSubmit: () => void;
	isSubmitting?: boolean;
}

export function PasswordStep({ onChange, onSubmit, isSubmitting }: Props) {
	const [current, setCurrent] = useState("");
	const [next, setNext] = useState("");
	const [confirm, setConfirm] = useState("");
	const [error, setError] = useState("");

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		setError("");

		if (next.length < 6) {
			setError("Password must be at least 6 characters");
			return;
		}
		if (next !== confirm) {
			setError("Passwords do not match");
			return;
		}

		onChange({ currentPassword: current, newPassword: next });
		onSubmit();
	};

	return (
		<form onSubmit={handleSubmit} className="space-y-4">
			<div>
				<label
					htmlFor="pw-current"
					className="mb-1 block text-xs font-medium text-[#6a7a6a]"
				>
					Current Password
				</label>
				<input
					id="pw-current"
					type="password"
					value={current}
					onChange={(e) => setCurrent(e.target.value)}
					className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-4 py-3 text-sm text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[#7dba7a] focus:ring-1 focus:ring-[#7dba7a]/30"
					placeholder="Enter current password"
				/>
			</div>
			<div>
				<label
					htmlFor="pw-new"
					className="mb-1 block text-xs font-medium text-[#6a7a6a]"
				>
					New Password
				</label>
				<input
					id="pw-new"
					type="password"
					value={next}
					onChange={(e) => setNext(e.target.value)}
					className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-4 py-3 text-sm text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[#7dba7a] focus:ring-1 focus:ring-[#7dba7a]/30"
					placeholder="New password (min 6 chars)"
				/>
			</div>
			<div>
				<label
					htmlFor="pw-confirm"
					className="mb-1 block text-xs font-medium text-[#6a7a6a]"
				>
					Confirm New Password
				</label>
				<input
					id="pw-confirm"
					type="password"
					value={confirm}
					onChange={(e) => setConfirm(e.target.value)}
					className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-4 py-3 text-sm text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[#7dba7a] focus:ring-1 focus:ring-[#7dba7a]/30"
					placeholder="Confirm new password"
				/>
			</div>

			{error && <p className="text-sm text-red-400">{error}</p>}

			<button
				type="submit"
				disabled={isSubmitting || !current || !next || !confirm}
				className="hidden w-full rounded-lg bg-linear-to-r from-[#7dba7a] to-[#5a8f5a] px-4 py-3 text-sm font-semibold text-[#080908] transition hover:from-[#8dca8a] hover:to-[#6a9f6a] disabled:cursor-not-allowed disabled:opacity-50"
			>
				{isSubmitting ? "Setting..." : "Set Password"}
			</button>
		</form>
	);
}
