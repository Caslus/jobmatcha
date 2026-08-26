import { createFileRoute, redirect, useNavigate } from "@tanstack/react-router";
import { X } from "lucide-react";
import { Particles } from "#/components/ui/particles.tsx";
import { OnboardingWizard } from "../features/onboarding/OnboardingWizard";
import { authStatusQueryOptions, useAuthStatus } from "../hooks/useApi";

export const Route = createFileRoute("/onboarding")({
	beforeLoad: async ({ context }) => {
		const status = await context.queryClient.ensureQueryData(
			authStatusQueryOptions(),
		);
		if (!status.authenticated) {
			throw redirect({ to: "/" });
		}
	},
	component: OnboardingPage,
});

function OnboardingPage() {
	const navigate = useNavigate();
	const { data: status } = useAuthStatus();
	const isReRun = status?.setup_complete ?? false;

	return (
		<div className="flex min-h-screen items-center justify-center">
			<div className="fixed -z-10 h-[200%] w-[200%]">
				<Particles
					className="w-full h-full opacity-20"
					color="#7dba7a"
					vy={-0.3}
					size={20}
					staticity={70}
					quantity={40}
				/>
			</div>
			<div
				className="relative w-full max-w-sm rounded-2xl border border-[#1a2a1a] bg-[#0d120d] p-8 shadow-2xl animate-[cardIn_1.8s_cubic-bezier(0.16,1,0.3,1)]"
				style={{ animationFillMode: "both" }}
			>
				<style>{`
					@keyframes cardIn {
						0% { opacity: 0; transform: scale(0.85) translateY(40px); filter: blur(6px); }
						30% { opacity: 1; filter: blur(0); }
						100% { opacity: 1; transform: scale(1) translateY(0); filter: blur(0); }
					}
					@keyframes glowDecay {
						0% { box-shadow: 0 0 0 0 rgba(125,186,122,0); }
						12% { box-shadow: 0 0 80px -10px rgba(125,186,122,0.12); }
						100% { box-shadow: 0 0 0 0 rgba(125,186,122,0); }
					}
				`}</style>
				<div
					className="absolute inset-0 rounded-2xl pointer-events-none animate-[glowDecay_5s_ease-out]"
					style={{ animationFillMode: "both" }}
				/>
				{isReRun && (
					<button
						type="button"
						onClick={() => navigate({ to: "/dashboard" })}
						className="absolute top-2 right-2 flex h-7 w-7 items-center justify-center rounded-full bg-[#1a2a1a] text-[#6a7a6a] transition hover:bg-[#2a3a2a] hover:text-[#e8e8e8]"
					>
						<X size={14} />
					</button>
				)}
				<div className="mb-8 text-center">
					<div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-linear-to-br from-[#7dba7a] to-[#4a7c4f] shadow-lg shadow-[#7dba7a]/20">
						<span className="text-2xl">🔐</span>
					</div>
				</div>
				<OnboardingWizard />
			</div>
		</div>
	);
}
