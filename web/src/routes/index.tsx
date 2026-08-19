import { createFileRoute, redirect, useNavigate } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { useLogin } from "../hooks/useApi";
import { useAuthStore } from "../stores/auth";

export const Route = createFileRoute("/")({
	beforeLoad: () => {
		const { authenticated } = useAuthStore.getState();
		if (authenticated) {
			throw redirect({ to: "/dashboard" });
		}
	},
	component: LoginPage,
});

function LoginPage() {
	const navigate = useNavigate();
	const [password, setPassword] = useState("");
	const { authenticated, loading, check } = useAuthStore();
	const loginMutation = useLogin();

	useEffect(() => {
		check().then(() => {
			const state = useAuthStore.getState();
			if (state.authenticated) {
				navigate({ to: "/dashboard" });
			}
		});
	}, []);

	if (loading) {
		return (
			<div className="flex min-h-screen items-center justify-center bg-[#080908]">
				<div className="h-8 w-8 animate-spin rounded-full border-2 border-[#7dba7a] border-t-transparent" />
			</div>
		);
	}

	if (authenticated) {
		return null;
	}

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		loginMutation.mutate(password);
	};

	return (
		<div className="flex min-h-screen items-center justify-center bg-[#080908]">
			<div className="w-full max-w-sm rounded-2xl border border-[#1a2a1a] bg-[#0d120d] p-8 shadow-2xl">
				<div className="mb-8 text-center">
					<div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-[#7dba7a] to-[#4a7c4f] shadow-lg shadow-[#7dba7a]/20">
						<span className="text-2xl">🍵</span>
					</div>
					<h1 className="text-2xl font-bold text-[#e8e8e8]">jobmatcha</h1>
					<p className="mt-1 text-sm text-[#6a7a6a]">
						Sign in to your dashboard
					</p>
				</div>

				<form onSubmit={handleSubmit} className="space-y-4">
					<div>
						<input
							type="password"
							placeholder="Password"
							value={password}
							onChange={(e) => setPassword(e.target.value)}
							className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-4 py-3 text-sm text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[#7dba7a] focus:ring-1 focus:ring-[#7dba7a]/30"
							autoFocus
						/>
					</div>

					{loginMutation.isError && (
						<p className="text-sm text-red-400">Invalid password.</p>
					)}

					<button
						type="submit"
						disabled={loginMutation.isPending || !password}
						className="w-full rounded-lg bg-gradient-to-r from-[#7dba7a] to-[#5a8f5a] px-4 py-3 text-sm font-semibold text-[#080908] transition hover:from-[#8dca8a] hover:to-[#6a9f6a] disabled:cursor-not-allowed disabled:opacity-50"
					>
						{loginMutation.isPending ? "Signing in..." : "Sign in"}
					</button>
				</form>
			</div>
		</div>
	);
}
