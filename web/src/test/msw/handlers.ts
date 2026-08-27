import { HttpResponse, http } from "msw";
import type {
	AIInfoResponse,
	AuthStatusResponse,
	RoleDetailResponse,
	RoleListResponse,
	ScanJobResponse,
	SettingsResponse,
} from "../../types/api.gen";
import { fixtures } from "./fixtures";

// Ky resolves the same-origin API base differently in jsdom and the browser;
// this wildcard intentionally accepts both `/api/...` and absolute test URLs.
const apiPath = (path: string) => `*/${path}`;

export const apiHandlers = {
	authStatus: (response: AuthStatusResponse = fixtures.authStatus) =>
		http.get(apiPath("auth/status"), () => HttpResponse.json(response)),
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
};

export const handlers = Object.values(apiHandlers).map((createHandler) =>
	createHandler(),
);
