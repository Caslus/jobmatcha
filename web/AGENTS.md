# React Frontend Guidelines & Architecture

## Role & Scope
You are a senior frontend engineer. All work in `/web` must follow modern React 19 patterns, TypeScript, Vite, Tailwind CSS, shadcn/ui components, and TanStack Query.

---

## Folder & Component Organization

1. **Feature-Driven Layout (`src/features/<feature-name>/`):**
   - Co-locate feature-specific components, hooks, and API queries together:
     - `features/jobs/` (Job board feed, match cards, filter sidebar)
     - `features/resumes/` (Resume builder, tailoring preview, diff viewer)
     - `features/letters/` (Cover letter generator, markdown editor)
2. **UI Primitives (`src/components/ui/`):**
   - Reusable components built with Tailwind and `shadcn/ui` (Button, Dialog, Input, Badge, Dropdown, Toast).
3. **Type Imports (`src/types/`):**
   - All backend contracts MUST be imported directly from `@/types/api.gen`. Never declare duplicate DTO interfaces manually.

---

## State, Data Fetching & HTTP Client

- **HTTP Client (`src/lib/api.ts`):** Use `ky` for all HTTP requests with default JSON headers, base URLs, and auth token interceptors.
- **Server State:** Use **TanStack Query (React Query)** for all queries and mutations.
- **Client State:** Keep ephemeral UI state (e.g. active modal, search inputs) in local `useState` or lightweight Zustand stores.
- **Form Handling:** Use TanStack Form (`@tanstack/react-form`) with the Zod validator adapter (@tanstack/zod-form-adapter) for all user input flows. Wrap all form controls with `<form.Field>` and compose standard primitives from `src/components/ui/`.
- **Testing:** Write unit/component tests using **Vitest** + React Testing Library under `__tests__` or co-located `.test.tsx` files.

---

## Styling & UX Rules

- **Tailwind CSS:** Use atomic utility classes with design system tokens.
- **Icons:** Use `lucide-react` exclusively.
- **Async State Handling:** Always render explicit loading skeletons, error alerts, and empty states.
- **Path Aliases:** Use `@/*` pointing to `src/*` for all internal imports.