import { createFileRoute, redirect, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { useChangePassword } from "../hooks/useApi";
import { useAuthStore } from "../stores/auth";

export const Route = createFileRoute("/onboarding")({
	beforeLoad: () => {
		const { authenticated } = useAuthStore.getState();
		if (!authenticated) {
			throw redirect({ to: "/" });
		}
	},
	component: OnboardingPage,
});

function OnboardingPage() {
	const navigate = useNavigate();
	const changePasswordMutation = useChangePassword();
	const [current, setCurrent] = useState("");
	const [next, setNext] = useState("");
	const [confirm, setConfirm] = useState("");
	const [error, setError] = useState("");

	const handleSubmit = async (e: React.FormEvent) => {
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

		changePasswordMutation.mutate(
			{ currentPassword: current, newPassword: next },
			{
				onSuccess: () => navigate({ to: "/dashboard" }),
				onError: (err) =>
					setError((err as Error).message || "Failed to change password"),
			},
		);
	};

	return (
		<div className="flex min-h-screen items-center justify-center bg-[#080908]">
			<div className="w-full max-w-sm rounded-2xl border border-[#1a2a1a] bg-[#0d120d] p-8 shadow-2xl">
				<div className="mb-8 text-center">
					<div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-[#7dba7a] to-[#4a7c4f] shadow-lg shadow-[#7dba7a]/20">
						<span className="text-2xl">🔐</span>
					</div>
					<h1 className="text-2xl font-bold text-[#e8e8e8]">
						Set Your Password
					</h1>
					<p className="mt-1 text-sm text-[#6a7a6a]">
						Change the default password to continue
					</p>
				</div>

				<form onSubmit={handleSubmit} className="space-y-4">
					<div>
						<label className="mb-1 block text-xs font-medium text-[#6a7a6a]">
							Current Password
						</label>
						<input
							type="password"
							value={current}
							onChange={(e) => setCurrent(e.target.value)}
							className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-4 py-3 text-sm text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[#7dba7a] focus:ring-1 focus:ring-[#7dba7a]/30"
							autoFocus
						/>
					</div>
					<div>
						<label className="mb-1 block text-xs font-medium text-[#6a7a6a]">
							New Password
						</label>
						<input
							type="password"
							value={next}
							onChange={(e) => setNext(e.target.value)}
							className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-4 py-3 text-sm text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[#7dba7a] focus:ring-1 focus:ring-[#7dba7a]/30"
						/>
					</div>
					<div>
						<label className="mb-1 block text-xs font-medium text-[#6a7a6a]">
							Confirm New Password
						</label>
						<input
							type="password"
							value={confirm}
							onChange={(e) => setConfirm(e.target.value)}
							className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-4 py-3 text-sm text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[#7dba7a] focus:ring-1 focus:ring-[#7dba7a]/30"
						/>
					</div>

					{error && <p className="text-sm text-red-400">{error}</p>}

					<button
						type="submit"
						disabled={
							changePasswordMutation.isPending || !current || !next || !confirm
						}
						className="w-full rounded-lg bg-gradient-to-r from-[#7dba7a] to-[#5a8f5a] px-4 py-3 text-sm font-semibold text-[#080908] transition hover:from-[#8dca8a] hover:to-[#6a9f6a] disabled:cursor-not-allowed disabled:opacity-50"
					>
						{changePasswordMutation.isPending ? "Setting..." : "Set Password"}
					</button>
				</form>
			</div>
		</div>
	);
}
