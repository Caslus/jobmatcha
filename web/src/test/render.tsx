import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
	createMemoryHistory,
	createRootRouteWithContext,
	createRoute,
	createRouter,
	Outlet,
	RouterProvider,
} from "@tanstack/react-router";
import { type RenderOptions, render } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactElement } from "react";

interface TestRouterContext {
	queryClient: QueryClient;
}

type ProviderRenderOptions = Omit<RenderOptions, "wrapper"> & {
	queryClient?: QueryClient;
};

type RouterRenderOptions = ProviderRenderOptions & {
	initialEntries?: string[];
};

export function createTestQueryClient() {
	return new QueryClient({
		defaultOptions: {
			queries: { retry: false },
			mutations: { retry: false },
		},
	});
}

export function renderWithProviders(
	ui: ReactElement,
	{
		queryClient = createTestQueryClient(),
		...options
	}: ProviderRenderOptions = {},
) {
	return {
		user: userEvent.setup(),
		queryClient,
		...render(ui, {
			...options,
			wrapper: ({ children }) => (
				<QueryClientProvider client={queryClient}>
					{children}
				</QueryClientProvider>
			),
		}),
	};
}

export function renderWithRouter(
	ui: ReactElement,
	{
		initialEntries = ["/"],
		queryClient = createTestQueryClient(),
		...options
	}: RouterRenderOptions = {},
) {
	const rootRoute = createRootRouteWithContext<TestRouterContext>()({
		component: Outlet,
	});
	const indexRoute = createRoute({
		getParentRoute: () => rootRoute,
		path: "/",
		component: () => ui,
	});
	const router = createRouter({
		routeTree: rootRoute.addChildren([indexRoute]),
		context: { queryClient },
		history: createMemoryHistory({ initialEntries }),
	});

	return {
		user: userEvent.setup(),
		queryClient,
		router,
		...render(
			<QueryClientProvider client={queryClient}>
				<RouterProvider router={router} />
			</QueryClientProvider>,
			options,
		),
	};
}
