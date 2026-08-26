interface Props {
	value: {
		currentPassword: string;
		newPassword: string;
		confirmPassword: string;
	};
	onChange: (data: {
		currentPassword: string;
		newPassword: string;
		confirmPassword: string;
	}) => void;
	onSubmit: () => void;
	isSubmitting?: boolean;
}

export function PasswordStep({
	value,
	onChange,
	onSubmit,
	isSubmitting,
}: Props) {
	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
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
					value={value.currentPassword}
					onChange={(e) =>
						onChange({ ...value, currentPassword: e.target.value })
					}
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
					value={value.newPassword}
					onChange={(e) => onChange({ ...value, newPassword: e.target.value })}
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
					value={value.confirmPassword}
					onChange={(e) =>
						onChange({ ...value, confirmPassword: e.target.value })
					}
					className="w-full rounded-lg border border-[#2a3a2a] bg-[#0a0f0a] px-4 py-3 text-sm text-[#e8e8e8] placeholder-[#4a5a4a] outline-none transition focus:border-[#7dba7a] focus:ring-1 focus:ring-[#7dba7a]/30"
					placeholder="Confirm new password"
				/>
			</div>

			<button type="submit" className="sr-only" disabled={isSubmitting}>
				Set Password
			</button>
		</form>
	);
}
