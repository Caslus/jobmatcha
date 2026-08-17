import { create } from "zustand";
import { authApi } from "../lib/api";

interface AuthState {
	authenticated: boolean;
	setupComplete: boolean;
	loading: boolean;
	check: () => Promise<void>;
	logout: () => Promise<void>;
}

function hasToken(): boolean {
	try {
		return !!localStorage.getItem("jobmatcha_token");
	} catch {
		return false;
	}
}

function clearToken() {
	try {
		localStorage.removeItem("jobmatcha_token");
	} catch {
		// localStorage unavailable
	}
}

export const useAuthStore = create<AuthState>((set) => ({
	authenticated: hasToken(),
	setupComplete: false,
	loading: true,

	check: async () => {
		const token = hasToken();
		if (!token) {
			set({ authenticated: false, setupComplete: false, loading: false });
			return;
		}
		try {
			const status = await authApi.status();
			set({
				authenticated: status.authenticated,
				setupComplete: status.setup_complete,
				loading: false,
			});
		} catch {
			set({ authenticated: false, setupComplete: false, loading: false });
		}
	},

	logout: () => {
		clearToken();
		set({ authenticated: false, setupComplete: false });
		return Promise.resolve();
	},
}));
