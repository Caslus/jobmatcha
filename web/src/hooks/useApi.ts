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
	companiesApi,
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

// ---- Companies hooks ----

export function useCompanies() {
	return useQuery({
		queryKey: queryKeys.companies.list(),
		queryFn: () => companiesApi.list(),
		staleTime: 1000 * 60 * 2,
	});
}

export function useUpdateCompanyActive() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({ id, active }: { id: number; active: boolean }) =>
			companiesApi.updateActive(id, { active }),
		onSuccess: () =>
			queryClient.invalidateQueries({ queryKey: queryKeys.companies.list() }),
	});
}

export function useUpdateCareerBoardActive() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({
			companyID,
			boardID,
			active,
		}: {
			companyID: number;
			boardID: number;
			active: boolean;
		}) => companiesApi.updateBoardActive(companyID, boardID, { active }),
		onSuccess: () =>
			queryClient.invalidateQueries({ queryKey: queryKeys.companies.list() }),
	});
}

export function useUpdateCompaniesActiveBulk() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({
			companyIDs,
			active,
		}: {
			companyIDs: number[];
			active: boolean;
		}) => companiesApi.updateActiveBulk({ company_ids: companyIDs, active }),
		onSuccess: () =>
			queryClient.invalidateQueries({ queryKey: queryKeys.companies.list() }),
	});
}

function useCompanyListMutation<TData, TVariables>(
	mutationFn: (variables: TVariables) => Promise<TData>,
) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn,
		onSuccess: () =>
			queryClient.invalidateQueries({ queryKey: queryKeys.companies.list() }),
	});
}

export function useUpdateCompanyDetails() {
	return useCompanyListMutation(({ id, name }: { id: number; name: string }) =>
		companiesApi.updateDetails(id, { name }),
	);
}

export function useDeleteCompany() {
	return useCompanyListMutation(({ id }: { id: number }) =>
		companiesApi.delete(id),
	);
}

export function useCreateCareerBoard() {
	return useCompanyListMutation(
		({
			companyID,
			provider,
			board_identifier,
			canonical_url,
		}: {
			companyID: number;
			provider: string;
			board_identifier: string;
			canonical_url: string;
		}) =>
			companiesApi.createBoard(companyID, {
				provider,
				board_identifier,
				canonical_url,
			}),
	);
}

export function useUpdateCareerBoardDetails() {
	return useCompanyListMutation(
		({
			companyID,
			boardID,
			provider,
			board_identifier,
			canonical_url,
		}: {
			companyID: number;
			boardID: number;
			provider: string;
			board_identifier: string;
			canonical_url: string;
		}) =>
			companiesApi.updateBoardDetails(companyID, boardID, {
				provider,
				board_identifier,
				canonical_url,
			}),
	);
}

export function useDeleteCareerBoard() {
	return useCompanyListMutation(
		({ companyID, boardID }: { companyID: number; boardID: number }) =>
			companiesApi.deleteBoard(companyID, boardID),
	);
}

export function useDiscoverCareerBoards() {
	return useMutation({
		mutationFn: (careersURL: string) =>
			companiesApi.discover({ careers_url: careersURL }),
	});
}

export function useRegisterCareerBoards() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (data: Parameters<typeof companiesApi.register>[0]) =>
			companiesApi.register(data),
		onSuccess: () =>
			queryClient.invalidateQueries({ queryKey: queryKeys.companies.list() }),
	});
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
