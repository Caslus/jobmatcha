import ky, { type HTTPError } from "ky";
import type {
	AIInfoResponse,
	AIUpdateRequest,
	AIValidateKeyRequest,
	AIValidateKeyResponse,
	AuthLoginResponse,
	AuthStatusResponse,
	ChangePasswordRequest,
	ErrorResponse,
	OnboardingCompleteRequest,
	ParseResumeResponse,
	RoleDetailResponse,
	RoleListResponse,
	ScanJobResponse,
	SettingsResponse,
	SettingsUpdateRequest,
	TailoredResumeResponse,
} from "@/types/api.gen";

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "/api";

export const api = ky.create({
	prefix: API_BASE,
	credentials: "same-origin",
	hooks: {
		beforeError: [
			({ error }) => {
				const httpError = error as HTTPError;
				const data = httpError.data as ErrorResponse | undefined;
				if (data?.error) {
					httpError.message = data.error;
				}
				return error;
			},
		],
	},
});

// Auth helpers
export const authApi = {
	login: (password: string) =>
		api.post("auth/login", { json: { password } }).json<AuthLoginResponse>(),

	logout: async () => {
		return api.post("auth/logout").json();
	},

	status: () => api.get("auth/status").json<AuthStatusResponse>(),

	changePassword: (currentPassword: string, newPassword: string) =>
		api
			.post("auth/change-password", {
				json: {
					current_password: currentPassword,
					new_password: newPassword,
				} satisfies ChangePasswordRequest,
			})
			.json(),
};

// Roles
export const rolesApi = {
	list: (page = 1, perPage = 25) =>
		api
			.get("roles", {
				searchParams: { page: String(page), per_page: String(perPage) },
			})
			.json<RoleListResponse>(),

	getByID: (id: number) => api.get(`roles/${id}`).json<RoleDetailResponse>(),

	patch: (
		id: number,
		updates: { is_hidden?: boolean; is_interested?: boolean },
	) => api.patch(`roles/${id}`, { json: updates }).json(),

	tailor: (id: number) =>
		api
			.post(`roles/${id}/tailor`, { timeout: 105_000 })
			.json<TailoredResumeResponse>(),

	getTailoredResume: (id: number) =>
		api
			.get(`roles/${id}/tailored-resume`)
			.json<TailoredResumeResponse | null>(),
};

// Settings
export const settingsApi = {
	get: () => api.get("settings").json<SettingsResponse>(),
	update: (data: SettingsUpdateRequest) =>
		api.put("settings", { json: data }).json(),
};

// Scan
export const scanApi = {
	start: () => api.post("scan").json<ScanJobResponse>(),
	get: (id: number) => api.get(`scan/${id}`).json<ScanJobResponse>(),
	getLatest: () => api.get("scan/latest").json<ScanJobResponse>(),
};

// AI
export const aiApi = {
	validateKey: (data: AIValidateKeyRequest) =>
		api.post("ai/validate-key", { json: data }).json<AIValidateKeyResponse>(),
	getSettings: () => api.get("settings/ai").json<AIInfoResponse>(),
	updateSettings: (data: AIUpdateRequest) =>
		api.put("settings/ai", { json: data }).json(),
	parseResume: (file: File) => {
		const form = new FormData();
		form.append("file", file);
		return api
			.post("ai/parse-resume", { body: form, timeout: 105_000 })
			.json<ParseResumeResponse>();
	},
};

// Onboarding
export const onboardingApi = {
	complete: (data: OnboardingCompleteRequest) =>
		api.post("onboarding/complete", { json: data }).json(),
};
