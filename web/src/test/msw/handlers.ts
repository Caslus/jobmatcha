import { HttpResponse, http } from "msw";
import type {
	AIInfoResponse,
	AuthStatusResponse,
	ParseResumeResponse,
	RoleDetailResponse,
	RoleListResponse,
	ScanJobResponse,
	SettingsResponse,
} from "../../types/api.gen";
import { fixtures } from "./fixtures";

const apiPath = (path: string) => `*/api/${path}`;

export const apiHandlers = {
	authStatus: (response: AuthStatusResponse = fixtures.authStatus) =>
		http.get(apiPath("auth/status"), () => HttpResponse.json(response)),
	login: (response = { authenticated: true }) =>
		http.post(apiPath("auth/login"), () => HttpResponse.json(response)),
	logout: () => http.post(apiPath("auth/logout"), () => HttpResponse.json({})),
	changePassword: () =>
		http.post(apiPath("auth/change-password"), () => HttpResponse.json({})),
	roles: (response: RoleListResponse = fixtures.roleList) =>
		http.get(apiPath("roles"), () => HttpResponse.json(response)),
	role: (response: RoleDetailResponse = fixtures.roleDetail) =>
		http.get(apiPath("roles/:id"), () => HttpResponse.json(response)),
	settings: (response: SettingsResponse = fixtures.settings) =>
		http.get(apiPath("settings"), () => HttpResponse.json(response)),
	scanLatest: (response: ScanJobResponse = fixtures.scanJob) =>
		http.get(apiPath("scan/latest"), () => HttpResponse.json(response)),
	aiSettings: (response: AIInfoResponse = fixtures.aiSettings) =>
		http.get(apiPath("settings/ai"), () => HttpResponse.json(response)),
	validateKey: (valid = true) =>
		http.post(apiPath("ai/validate-key"), () => HttpResponse.json({ valid })),
	parseResume: (response?: ParseResumeResponse) =>
		http.post(apiPath("ai/parse-resume"), () =>
			HttpResponse.json(response ?? {}),
		),
	completeOnboarding: () =>
		http.post(apiPath("onboarding/complete"), () => HttpResponse.json({})),
};

export const handlers = Object.values(apiHandlers).map((createHandler) =>
	createHandler(),
);
