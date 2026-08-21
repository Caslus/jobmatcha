interface Props {
	checked: boolean;
	onChange: () => void;
}

export function ToggleSwitch({ checked, onChange }: Props) {
	return (
		<button
			type="button"
			role="switch"
			aria-checked={checked}
			onClick={onChange}
			className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full transition-colors ${
				checked ? "bg-[#7dba7a]" : "bg-[#2a3a2a]"
			}`}
		>
			<span
				className={`inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform ${
					checked ? "translate-x-4.5" : "translate-x-0.75"
				}`}
			/>
		</button>
	);
}
