import { CheckCircle, XCircle } from "lucide-react";
import { useState } from "react";
import { useValidateKey } from "../../hooks/useApi";
import { LoadingAnimation } from "./LoadingAnimation";

interface Props {
	value: {
		provider: string;
		apiKey: string;
		enabled: boolean;
	};
	onChange: (data: {
		provider: string;
		apiKey: string;
		enabled: boolean;
	}) => void;
	onValidated: () => void;
	onSkipped: () => void;
	hideSkip?: boolean;
	hasExistingKey?: boolean;
}

export function AIProviderStep({
	value,
	onChange,
	onValidated,
	onSkipped,
	hideSkip,
	hasExistingKey,
}: Props) {
	const validateKey = useValidateKey();
	const [apiKey, setApiKey] = useState(value.apiKey);
	const [valid, setValid] = useState<boolean | null>(null);
	const [keyChanged, setKeyChanged] = useState(false);
	const wasPreFilled = hasExistingKey && !value.apiKey;

	const handleValidate = async () => {
		if (!apiKey.trim()) return;
		setValid(null);
		const resp = await validateKey.mutateAsync({
			provider: value.provider,
			api_key: apiKey,
		});
		setValid(resp.valid);
		if (resp.valid) {
			onChange({ ...value, apiKey, enabled: true });
			onValidated();
		}
	};

	return (
		<div className="space-y-4">
			{/* Provider selector */}
			<div>
				<label
					htmlFor="ai-provider"
					className="mb-1 block text-xs font-medium text-[#6a7a6a]"
				>
					AI Provider
				</label>
				<select
					id="ai-provider"
					value={value.provider}
					onChange={(e) => onChange({ ...value, provider: e.target.value })}
					className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-4 py-3 text-sm text-[#e8e8e8] outline-none transition focus:border-[#7dba7a]"
				>
					<option value="openrouter">OpenRouter</option>
				</select>
			</div>

			{/* API key */}
			<div>
				<label
					htmlFor="ai-api-key"
					className="mb-1 block text-xs font-medium text-[#6a7a6a]"
				>
					API Key
				</label>
				<div className="flex gap-2">
					<input
						id="ai-api-key"
						type="password"
						value={apiKey}
						onChange={(e) => {
							setApiKey(e.target.value);
							setValid(null);
							setKeyChanged(true);
						}}
						placeholder={wasPreFilled ? "sk-••••••••" : "sk-or-v1-..."}
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

				{wasPreFilled && !keyChanged && (
					<p className="mt-1 text-[10px] text-[#4a5a4a]">
						Key is already saved. Just click Continue to proceed.
					</p>
				)}

				{/* Validation status */}
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
					<p className="mt-2 flex items-center gap-1.5 text-xs text-red-400">
						<XCircle size={12} />
						Invalid key
					</p>
				)}
			</div>

			{/* Skip (only if AI was not previously configured) */}
			{!hideSkip && (
				<button
					type="button"
					onClick={onSkipped}
					className="text-xs text-[#4a5a4a] underline transition hover:text-[#6a7a6a]"
				>
					Skip AI setup — I'll configure later
				</button>
			)}
		</div>
	);
}
