import ky, { type HTTPError } from "ky";
import type {
	AIInfoResponse,
	AIUpdateRequest,
	AIValidateKeyRequest,
	AIValidateKeyResponse,
	AuthStatusResponse,
	AuthTokenResponse,
	ErrorResponse,
	ParseResumeResponse,
	RoleDetailResponse,
	RoleListResponse,
	ScanJobResponse,
	SettingsResponse,
	SettingsUpdateRequest,
	TailoredResumeResponse,
} from "@/types/api.gen";

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "/api";

function getToken(): string | null {
	if (typeof window === "undefined") return null;
	try {
		return localStorage.getItem("jobmatcha_token");
	} catch {
		return null;
	}
}

function setToken(token: string) {
	if (typeof window === "undefined") return;
	try {
		localStorage.setItem("jobmatcha_token", token);
	} catch {
		// localStorage unavailable
	}
}

function clearToken() {
	if (typeof window === "undefined") return;
	try {
		localStorage.removeItem("jobmatcha_token");
	} catch {
		// localStorage unavailable
	}
}

export const api = ky.create({
	prefix: API_BASE,
	hooks: {
		beforeRequest: [
			({ request }) => {
				const token = getToken();
				if (token) {
					request.headers.set("Authorization", `Bearer ${token}`);
				}
			},
		],
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
	login: async (password: string) => {
		const res = await api
			.post("auth/login", { json: { password } })
			.json<AuthTokenResponse>();
		setToken(res.token);
		return res;
	},

	logout: async () => {
		const res = await api.post("auth/logout").json();
		clearToken();
		return res;
	},

	status: () => api.get("auth/status").json<AuthStatusResponse>(),

	changePassword: (currentPassword: string, newPassword: string) =>
		api
			.post("auth/change-password", {
				json: { current_password: currentPassword, new_password: newPassword },
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
	complete: (data: Record<string, unknown>) =>
		api.post("onboarding/complete", { json: data }).json(),
};
