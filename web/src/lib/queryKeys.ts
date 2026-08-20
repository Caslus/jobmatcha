export const queryKeys = {
	auth: {
		status: ["auth", "status"] as const,
	},
	roles: {
		all: ["roles"] as const,
		lists: () => [...queryKeys.roles.all, "list"] as const,
		list: (page: number, perPage: number) =>
			[...queryKeys.roles.lists(), { page, perPage }] as const,
		details: () => [...queryKeys.roles.all, "detail"] as const,
		detail: (id: number) => [...queryKeys.roles.details(), id] as const,
	},
	settings: {
		all: ["settings"] as const,
	},
	scan: {
		all: ["scan"] as const,
		latest: () => [...queryKeys.scan.all, "latest"] as const,
		detail: (id: number) => [...queryKeys.scan.all, "detail", id] as const,
	},
};
