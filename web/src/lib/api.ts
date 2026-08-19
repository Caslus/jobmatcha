import ky from "ky";
import type {
	AuthStatusResponse,
	AuthTokenResponse,
	ErrorResponse,
	RoleDetailResponse,
	RoleListResponse,
	ScanJobResponse,
	SettingsResponse,
	SettingsUpdateRequest,
} from "@/types/api.gen";

const API_BASE = "http://localhost:8181/api";

function getToken(): string | null {
	try {
		return localStorage.getItem("jobmatcha_token");
	} catch {
		return null;
	}
}

function setToken(token: string) {
	try {
		localStorage.setItem("jobmatcha_token", token);
	} catch {
		// localStorage unavailable
	}
}

function clearToken() {
	try {
		localStorage.removeItem("jobmatcha_token");
	} catch {
		// localStorage unavailable
	}
}

export const api = ky.create({
	prefix: API_BASE,
	headers: { "Content-Type": "application/json" },
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
				const data = (error as any).data as ErrorResponse | undefined;
				if (data?.error) {
					(error as any).message = data.error;
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
