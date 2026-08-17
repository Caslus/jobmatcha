import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { authApi, rolesApi } from "../lib/api";
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
		queryKey: queryKeys.roles.detail(id!),
		queryFn: () => rolesApi.getByID(id!),
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
		},
	});
}
