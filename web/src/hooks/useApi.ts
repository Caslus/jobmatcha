import {
	queryOptions,
	useMutation,
	useQuery,
	useQueryClient,
} from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import type { AIInfoResponse } from "@/types/api.gen";
import {
	aiApi,
	authApi,
	onboardingApi,
	rolesApi,
	scanApi,
	settingsApi,
} from "../lib/api";
import { queryKeys } from "../lib/queryKeys";

export const authStatusQueryOptions = () =>
	queryOptions({
		queryKey: queryKeys.auth.status,
		queryFn: () => authApi.status(),
		retry: false,
		staleTime: 1000 * 60 * 5,
	});

// ---- Auth hooks ----

export function useAuthStatus() {
	return useQuery(authStatusQueryOptions());
}

export function useLogin() {
	const queryClient = useQueryClient();
	const navigate = useNavigate();

	return useMutation({
		mutationFn: (password: string) => authApi.login(password),
		onSuccess: async () => {
			await queryClient.invalidateQueries({ queryKey: queryKeys.auth.status });
			navigate({ to: "/dashboard" });
		},
	});
}

export function useLogout() {
	const queryClient = useQueryClient();
	const navigate = useNavigate();

	return useMutation({
		mutationFn: () => authApi.logout(),
		onSuccess: () => {
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

export function useTailoredResume(id: number | null) {
	return useQuery({
		queryKey: queryKeys.roles.tailoredResume(id as number),
		queryFn: () => rolesApi.getTailoredResume(id as number),
		enabled: id !== null,
		retry: false,
		staleTime: 1000 * 60 * 5,
	});
}

export function useTailorResume() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (id: number) => rolesApi.tailor(id),
		onSuccess: (tailored) => {
			queryClient.setQueryData(
				queryKeys.roles.tailoredResume(tailored.role_id),
				tailored,
			);
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
		onSuccess: (_, updates) => {
			queryClient.setQueryData<AIInfoResponse>(
				queryKeys.ai.settings,
				(current) =>
					current
						? {
								...current,
								provider: updates.provider ?? current.provider,
								enabled: updates.enabled ?? current.enabled,
								has_api_key: updates.api_key ? true : current.has_api_key,
								user_name: updates.user_name ?? current.user_name,
								user_email: updates.user_email ?? current.user_email,
								user_location: updates.user_location ?? current.user_location,
								user_linkedin: updates.user_linkedin ?? current.user_linkedin,
								user_github: updates.user_github ?? current.user_github,
							}
						: current,
			);
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
		mutationFn: (data: Parameters<typeof onboardingApi.complete>[0]) =>
			onboardingApi.complete(data),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.auth.status });
			queryClient.invalidateQueries({ queryKey: queryKeys.settings.all });
		},
	});
}
