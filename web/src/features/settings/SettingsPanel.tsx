import { useNavigate } from "@tanstack/react-router";
import { CheckCircle, Settings as SettingsIcon } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { Modal } from "@/components/ui/modal";
import type { AIUpdateRequest } from "@/types/api.gen";
import {
	useAISettings,
	useChangePassword,
	useSettings,
	useUpdateAISettings,
	useValidateKey,
} from "../../hooks/useApi";
import { LoadingAnimation } from "../onboarding/LoadingAnimation";

type Tab = "ai" | "profile" | "password" | "preferences";

const TABS: { id: Tab; label: string }[] = [
	{ id: "ai", label: "AI Provider" },
	{ id: "profile", label: "Profile" },
	{ id: "password", label: "Password" },
	{ id: "preferences", label: "Keywords" },
];

interface Props {
	onClose: () => void;
}

// ── AI Provider Tab ──────────────────────────

function AITab() {
	const { data: aiSettings, isLoading } = useAISettings();
	const updateAi = useUpdateAISettings();
	const validateKey = useValidateKey();
	const [enabled, setEnabled] = useState(false);
	const [apiKey, setApiKey] = useState("");
	const [provider, setProvider] = useState("openrouter");
	const [valid, setValid] = useState<boolean | null>(null);
	const [saved, setSaved] = useState(false);
	const [init, setInit] = useState(false);

	if (aiSettings && !init) {
		setInit(true);
		setEnabled(aiSettings.enabled);
		setProvider(aiSettings.provider || "openrouter");
	}

	const handleValidate = async () => {
		if (!apiKey.trim()) return;
		setValid(null);
		const resp = await validateKey.mutateAsync({
			provider,
			api_key: apiKey,
		});
		setValid(resp.valid);
	};

	const handleSave = () => {
		const payload: AIUpdateRequest = { enabled };
		if (valid && apiKey.trim()) {
			payload.api_key = apiKey;
			payload.provider = provider;
		}
		updateAi.mutate(payload, {
			onSuccess: () => {
				setSaved(true);
				setTimeout(() => setSaved(false), 2000);
			},
		});
	};

	const canSave = (() => {
		if (!enabled) {
			// Toggling from off to on without a key — just save enabled state
			return aiSettings?.enabled !== enabled;
		}
		// A previously saved key can be enabled again without re-entering it.
		// A newly entered key must still be validated before saving.
		return (
			aiSettings?.has_api_key === true ||
			(valid === true && apiKey.trim().length > 0)
		);
	})();

	if (isLoading) return <LoadingAnimation label="Loading..." />;

	return (
		<div className="space-y-4">
			{/* Enabled toggle — always shown first */}
			<label className="flex items-center justify-between cursor-pointer">
				<span className="text-xs font-medium text-[#e8e8e8]">Enabled</span>
				<button
					type="button"
					role="switch"
					aria-checked={enabled}
					onClick={() => {
						setEnabled(!enabled);
						setSaved(false);
					}}
					className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full transition-colors ${
						enabled ? "bg-[#7dba7a]" : "bg-[#2a3a2a]"
					}`}
				>
					<span
						className={`inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform ${
							enabled ? "translate-x-4.5" : "translate-x-0.75"
						}`}
					/>
				</button>
			</label>

			{/* Fields only shown when enabled */}
			{enabled && (
				<>
					{/* Provider selector */}
					<div>
						<label
							htmlFor="settings-provider"
							className="mb-1 block text-xs font-medium text-[#6a7a6a]"
						>
							AI Provider
						</label>
						<select
							id="settings-provider"
							value={provider}
							onChange={(e) => {
								setProvider(e.target.value);
								setValid(null);
								setSaved(false);
							}}
							className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-4 py-3 text-sm text-[#e8e8e8] outline-none transition focus:border-[#7dba7a]"
						>
							<option value="openrouter">OpenRouter</option>
						</select>
					</div>

					{/* API key */}
					<div>
						<label
							htmlFor="settings-api-key"
							className="mb-1 block text-xs font-medium text-[#6a7a6a]"
						>
							API Key
						</label>
						<div className="flex gap-2">
							<input
								id="settings-api-key"
								type="password"
								value={apiKey}
								onChange={(e) => {
									setApiKey(e.target.value);
									setValid(null);
									setSaved(false);
								}}
								placeholder={
									aiSettings?.has_api_key ? "sk-••••••••" : "sk-or-v1-..."
								}
								className="flex-1 rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-4 py-3 text-sm text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[#7dba7a]"
							/>
							<button
								type="button"
								onClick={handleValidate}
								disabled={!apiKey.trim() || validateKey.isPending}
								className="rounded-lg bg-[#1a2a1a] px-4 py-3 text-xs font-medium text-[#6a7a6a] transition hover:bg-[#2a3a2a] disabled:cursor-not-allowed disabled:opacity-40"
							>
								Validate
							</button>
						</div>

						{aiSettings?.has_api_key && !apiKey.trim() && (
							<p className="mt-1 text-[10px] text-[#4a5a4a]">
								A key is already saved. You can enable the provider and save, or
								replace the key below.
							</p>
						)}

						{validateKey.isPending && (
							<LoadingAnimation label="Validating key..." />
						)}
						{valid === true && (
							<p className="mt-2 flex items-center gap-1.5 text-xs text-[#7dba7a]">
								<CheckCircle size={12} />
								Key is valid
							</p>
						)}
						{valid === false && (
							<p className="mt-2 text-xs text-red-400">Invalid key</p>
						)}
					</div>
				</>
			)}

			<button
				type="button"
				onClick={handleSave}
				disabled={!canSave || updateAi.isPending}
				className="w-full rounded-lg bg-linear-to-r from-[#7dba7a] to-[#5a8f5a] px-4 py-3 text-sm font-semibold text-[#080908] transition hover:from-[#8dca8a] hover:to-[#6a9f6a] disabled:cursor-not-allowed disabled:opacity-50"
			>
				{saved ? "Saved ✓" : updateAi.isPending ? "Saving..." : "Save"}
			</button>
		</div>
	);
}

// ── Profile Tab ──────────────────────────────

function ProfileTab() {
	const { data: aiSettings, isLoading } = useAISettings();
	const updateAi = useUpdateAISettings();
	const [draft, setDraft] = useState<Record<string, string>>({});
	const init = useRef(false);

	useEffect(() => {
		if (aiSettings && !init.current) {
			init.current = true;
			setDraft({
				user_name: aiSettings.user_name,
				user_email: aiSettings.user_email,
				user_location: aiSettings.user_location,
				user_linkedin: aiSettings.user_linkedin,
				user_github: aiSettings.user_github,
			});
		}
	}, [aiSettings]);

	const set = (field: string) => (e: React.ChangeEvent<HTMLInputElement>) => {
		setDraft((prev) => ({ ...prev, [field]: e.target.value }));
	};

	const handleSave = () => {
		updateAi.mutate(draft);
	};

	if (isLoading) return <LoadingAnimation label="Loading..." />;

	return (
		<div className="space-y-3">
			{[
				{ field: "user_name", label: "Name" },
				{ field: "user_email", label: "Email", type: "email" },
				{ field: "user_location", label: "Location" },
				{ field: "user_linkedin", label: "LinkedIn URL" },
				{ field: "user_github", label: "GitHub URL" },
			].map(({ field, label, type }) => (
				<div key={field}>
					<label
						htmlFor={field}
						className="mb-1 block text-xs font-medium text-[#6a7a6a]"
					>
						{label}
					</label>
					<input
						id={field}
						type={type ?? "text"}
						value={draft[field] ?? ""}
						onChange={set(field)}
						placeholder={label}
						className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-3 py-2 text-xs text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[#7dba7a]"
					/>
				</div>
			))}
			<button
				type="button"
				onClick={handleSave}
				disabled={updateAi.isPending}
				className="w-full rounded-lg bg-linear-to-r from-[#7dba7a] to-[#5a8f5a] px-3 py-2 text-xs font-semibold text-[#080908] transition hover:from-[#8dca8a] hover:to-[#6a9f6a] disabled:opacity-50"
			>
				{updateAi.isPending ? "Saving..." : "Save"}
			</button>
		</div>
	);
}

// ── Preferences Tab ──────────────────────────

function PreferencesTab() {
	const { data: settings, isLoading } = useSettings();

	if (isLoading) return <LoadingAnimation label="Loading..." />;

	return (
		<div className="space-y-3">
			<p className="text-xs text-[#6a7a6a]">
				Keywords and preferences are managed from the dashboard sidebar. Open
				the dashboard to edit include/exclude keywords and work types.
			</p>
			{settings && (
				<div className="space-y-2">
					<div className="rounded-lg border border-[#1a2a1a] bg-[#080b08] px-3 py-2">
						<p className="text-[10px] uppercase tracking-wider text-[#4a5a4a] mb-1">
							Include
						</p>
						<p className="text-xs text-[#e8e8e8]">
							{settings.include_keywords?.join(", ") || "None set"}
						</p>
					</div>
					<div className="rounded-lg border border-[#1a2a1a] bg-[#080b08] px-3 py-2">
						<p className="text-[10px] uppercase tracking-wider text-[#4a5a4a] mb-1">
							Exclude
						</p>
						<p className="text-xs text-[#e8e8e8]">
							{settings.exclude_keywords?.join(", ") || "None set"}
						</p>
					</div>
					<div className="rounded-lg border border-[#1a2a1a] bg-[#080b08] px-3 py-2">
						<p className="text-[10px] uppercase tracking-wider text-[#4a5a4a] mb-1">
							Max Age
						</p>
						<p className="text-xs text-[#e8e8e8]">
							{settings.max_days_old === 0
								? "Any date"
								: `${settings.max_days_old} days`}
						</p>
					</div>
				</div>
			)}
		</div>
	);
}

// ── Password Tab ──────────────────────────

function PasswordTab() {
	const changePassword = useChangePassword();
	const [current, setCurrent] = useState("");
	const [next, setNext] = useState("");
	const [confirm, setConfirm] = useState("");
	const [error, setError] = useState("");
	const [success, setSuccess] = useState(false);

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		setError("");
		setSuccess(false);

		if (next.length < 6) {
			setError("Password must be at least 6 characters");
			return;
		}
		if (next !== confirm) {
			setError("Passwords do not match");
			return;
		}

		changePassword.mutate(
			{ currentPassword: current, newPassword: next },
			{
				onSuccess: () => {
					setSuccess(true);
					setCurrent("");
					setNext("");
					setConfirm("");
					setTimeout(() => setSuccess(false), 3000);
				},
				onError: (err) => {
					setError(
						err instanceof Error ? err.message : "Failed to change password",
					);
				},
			},
		);
	};

	return (
		<form onSubmit={handleSubmit} className="space-y-4">
			<div>
				<label
					htmlFor="settings-pw-current"
					className="mb-1 block text-xs font-medium text-[#6a7a6a]"
				>
					Current Password
				</label>
				<input
					id="settings-pw-current"
					type="password"
					value={current}
					onChange={(e) => {
						setCurrent(e.target.value);
						setError("");
						setSuccess(false);
					}}
					className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-3 py-2 text-xs text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[#7dba7a]"
				/>
			</div>
			<div>
				<label
					htmlFor="settings-pw-new"
					className="mb-1 block text-xs font-medium text-[#6a7a6a]"
				>
					New Password
				</label>
				<input
					id="settings-pw-new"
					type="password"
					value={next}
					onChange={(e) => {
						setNext(e.target.value);
						setError("");
						setSuccess(false);
					}}
					className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-3 py-2 text-xs text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[#7dba7a]"
				/>
			</div>
			<div>
				<label
					htmlFor="settings-pw-confirm"
					className="mb-1 block text-xs font-medium text-[#6a7a6a]"
				>
					Confirm New Password
				</label>
				<input
					id="settings-pw-confirm"
					type="password"
					value={confirm}
					onChange={(e) => {
						setConfirm(e.target.value);
						setError("");
						setSuccess(false);
					}}
					className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-3 py-2 text-xs text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[#7dba7a]"
				/>
			</div>

			{error && <p className="text-xs text-red-400">{error}</p>}
			{success && <p className="text-xs text-[#7dba7a]">Password changed ✓</p>}

			<button
				type="submit"
				disabled={changePassword.isPending || !current || !next || !confirm}
				className="w-full rounded-lg bg-linear-to-r from-[#7dba7a] to-[#5a8f5a] px-3 py-2 text-xs font-semibold text-[#080908] transition hover:from-[#8dca8a] hover:to-[#6a9f6a] disabled:opacity-50"
			>
				{changePassword.isPending ? "Changing..." : "Change Password"}
			</button>
		</form>
	);
}

// ── Settings Panel ───────────────────────────

export function SettingsPanel({ onClose }: Props) {
	const navigate = useNavigate();
	const [tab, setTab] = useState<Tab>("ai");

	const handleReRunOnboarding = () => {
		onClose();
		navigate({ to: "/onboarding" });
	};

	return (
		<Modal
			open
			onClose={onClose}
			title="Settings"
			icon={<SettingsIcon size={14} className="text-[#6a7a6a]" />}
			panelClassName="max-w-md max-h-[90vh]"
		>
			<div className="flex min-h-0 flex-1 flex-col">
				{/* Tabs */}
				<div className="flex border-b border-[#1a2a1a]">
					{TABS.map((t) => (
						<button
							key={t.id}
							type="button"
							onClick={() => setTab(t.id)}
							className={`flex-1 px-3 py-2 text-xs font-medium transition ${
								tab === t.id
									? "border-b-2 border-[#7dba7a] text-[#e8e8e8]"
									: "text-[#4a5a4a] hover:text-[#6a7a6a]"
							}`}
						>
							{t.label}
						</button>
					))}
				</div>

				{/* Tab content */}
				<div className="flex-1 overflow-y-auto p-5">
					{tab === "ai" && <AITab />}
					{tab === "profile" && <ProfileTab />}
					{tab === "password" && <PasswordTab />}
					{tab === "preferences" && <PreferencesTab />}
				</div>

				{/* Footer: Re-run onboarding */}
				<div className="border-t border-[#1a2a1a] px-5 py-3">
					<button
						type="button"
						onClick={handleReRunOnboarding}
						className="w-full rounded-lg border border-dashed border-[#2a3a2a] px-3 py-2 text-xs text-[#4a5a4a] transition hover:border-[#4a5a4a] hover:text-[#6a7a6a]"
					>
						Run onboarding again
					</button>
				</div>
			</div>
		</Modal>
	);
}
