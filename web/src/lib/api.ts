import ky, { type HTTPError } from "ky";
import type {
	AIInfoResponse,
	AIUpdateRequest,
	AIValidateKeyRequest,
	AIValidateKeyResponse,
	AuthLoginResponse,
	AuthStatusResponse,
	CareerBoardDiscoveryRequest,
	CareerBoardDiscoveryResponse,
	CareerBoardRegistrationRequest,
	CareerBoardUpsertRequest,
	ChangePasswordRequest,
	CompanyActiveUpdateRequest,
	CompanyBulkActiveUpdateRequest,
	CompanyDetailsUpdateRequest,
	CompanyListItem,
	CompanyListResponse,
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
const responseErrors = new WeakMap<Request, string>();

export const api = ky.create({
	prefix: API_BASE,
	credentials: "same-origin",
	hooks: {
		afterResponse: [
			async ({ request, response }) => {
				if (!response.ok) {
					try {
						const data = (await response.clone().json()) as ErrorResponse;
						if (data?.error) responseErrors.set(request, data.error);
					} catch {
						// Some error responses have no JSON body. Preserve Ky's default message.
					}
				}
			},
		],
		beforeError: [
			async ({ error, request }) => {
				const httpError = error as HTTPError;
				if (!httpError.response) return error;
				const message = responseErrors.get(request);
				if (message) httpError.message = message;
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

export const companiesApi = {
	list: () => api.get("companies").json<CompanyListResponse>(),
	updateActive: (id: number, data: CompanyActiveUpdateRequest) =>
		api.patch(`companies/${id}`, { json: data }).json<CompanyListItem>(),
	updateBoardActive: (
		companyID: number,
		boardID: number,
		data: CompanyActiveUpdateRequest,
	) =>
		api
			.patch(`companies/${companyID}/boards/${boardID}`, { json: data })
			.json<CompanyListItem>(),
	updateActiveBulk: (data: CompanyBulkActiveUpdateRequest) =>
		api.put("companies/active", { json: data }).json<CompanyListResponse>(),
	updateDetails: (id: number, data: CompanyDetailsUpdateRequest) =>
		api
			.patch(`companies/${id}/details`, { json: data })
			.json<CompanyListItem>(),
	delete: (id: number) => api.delete(`companies/${id}`).json(),
	createBoard: (companyID: number, data: CareerBoardUpsertRequest) =>
		api
			.post(`companies/${companyID}/boards`, { json: data })
			.json<CompanyListItem>(),
	updateBoardDetails: (
		companyID: number,
		boardID: number,
		data: CareerBoardUpsertRequest,
	) =>
		api
			.patch(`companies/${companyID}/boards/${boardID}/details`, { json: data })
			.json<CompanyListItem>(),
	deleteBoard: (companyID: number, boardID: number) =>
		api
			.delete(`companies/${companyID}/boards/${boardID}`)
			.json<CompanyListItem>(),
	discover: (data: CareerBoardDiscoveryRequest) =>
		api
			.post("companies/discover", { json: data, timeout: 120_000 })
			.json<CareerBoardDiscoveryResponse>(),
	register: (data: CareerBoardRegistrationRequest) =>
		api.post("companies/register", { json: data }).json<CompanyListResponse>(),
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
