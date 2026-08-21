import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import {
	aiApi,
	authApi,
	onboardingApi,
	rolesApi,
	scanApi,
	settingsApi,
} from "../lib/api";
import { queryKeys } from "../lib/queryKeys";
import { useAuthStore } from "../stores/auth";

// ---- Auth hooks ----

export function useAuthStatus() {
	return useQuery({
		queryKey: queryKeys.auth.status,
		queryFn: () => authApi.status(),
		retry: false,
		staleTime: 1000 * 60 * 5,
	});
}

export function useLogin() {
	const queryClient = useQueryClient();
	const navigate = useNavigate();
	const { check } = useAuthStore();

	return useMutation({
		mutationFn: (password: string) => authApi.login(password),
		onSuccess: async () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.auth.status });
			await check();
			// Navigate regardless of check() result — the token is set
			navigate({ to: "/dashboard" });
		},
	});
}

export function useLogout() {
	const queryClient = useQueryClient();
	const navigate = useNavigate();
	const { logout } = useAuthStore();

	return useMutation({
		mutationFn: () => authApi.logout(),
		onSuccess: () => {
			logout();
			queryClient.invalidateQueries({ queryKey: queryKeys.auth.status });
			navigate({ to: "/" });
		},
	});
}

export function useChangePassword() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: ({
			currentPassword,
			newPassword,
		}: {
			currentPassword: string;
			newPassword: string;
		}) => authApi.changePassword(currentPassword, newPassword),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.auth.status });
		},
	});
}

// ---- Roles hooks ----

export function useRoles(page: number, perPage = 25) {
	return useQuery({
		queryKey: queryKeys.roles.list(page, perPage),
		queryFn: () => rolesApi.list(page, perPage),
		staleTime: 1000 * 60 * 2, // 2 minutes
		placeholderData: (previousData) => previousData,
	});
}

export function useRole(id: number | null) {
	return useQuery({
		queryKey: queryKeys.roles.detail(id as number),
		queryFn: () => rolesApi.getByID(id as number),
		enabled: id !== null,
		staleTime: 1000 * 60 * 5,
	});
}

export function usePatchRole() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: ({
			id,
			...updates
		}: {
			id: number;
			is_hidden?: boolean;
			is_interested?: boolean;
		}) => rolesApi.patch(id, updates),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.roles.lists() });
			queryClient.invalidateQueries({ queryKey: queryKeys.roles.details() });
		},
	});
}

// ---- Settings hooks ----

export function useSettings() {
	return useQuery({
		queryKey: queryKeys.settings.all,
		queryFn: () => settingsApi.get(),
		staleTime: 1000 * 60 * 5,
	});
}

export function useUpdateSettings() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (data: Parameters<typeof settingsApi.update>[0]) =>
			settingsApi.update(data),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.settings.all });
			queryClient.invalidateQueries({ queryKey: queryKeys.roles.lists() });
		},
	});
}

// ---- Scan hooks ----

export function useLatestScan() {
	return useQuery({
		queryKey: queryKeys.scan.latest(),
		queryFn: () => scanApi.getLatest(),
		retry: false,
		staleTime: 1000 * 30, // 30 seconds — poll-friendly
	});
}

export function useScan(id: number | null) {
	return useQuery({
		queryKey: queryKeys.scan.detail(id as number),
		queryFn: () => scanApi.get(id as number),
		enabled: id !== null,
		refetchInterval: (query) =>
			query.state.data?.status === "pending" ||
			query.state.data?.status === "running"
				? 2000
				: false,
	});
}

export function useStartScan() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: () => scanApi.start(),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.scan.latest() });
		},
	});
}

// ---- AI hooks ----

export function useAISettings() {
	return useQuery({
		queryKey: queryKeys.ai.settings,
		queryFn: () => aiApi.getSettings(),
		staleTime: 1000 * 60 * 5,
	});
}

export function useUpdateAISettings() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (data: Parameters<typeof aiApi.updateSettings>[0]) =>
			aiApi.updateSettings(data),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.ai.settings });
		},
	});
}

export function useValidateKey() {
	return useMutation({
		mutationFn: (data: Parameters<typeof aiApi.validateKey>[0]) =>
			aiApi.validateKey(data),
	});
}

export function useParseResume() {
	return useMutation({
		mutationFn: (file: File) => aiApi.parseResume(file),
	});
}

// ---- Onboarding hooks ----

export function useCompleteOnboarding() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (data: Record<string, unknown>) => onboardingApi.complete(data),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.auth.status });
			queryClient.invalidateQueries({ queryKey: queryKeys.settings.all });
		},
	});
}
