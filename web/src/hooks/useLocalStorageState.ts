import type { Dispatch, SetStateAction } from "react";
import { useCallback, useState } from "react";

type InitialValue<T> = T | (() => T);

interface LocalStorageStateOptions<T> {
	parse?: (raw: string) => T | undefined;
	serialize?: (value: T) => string;
}

function resolveInitialValue<T>(initialValue: InitialValue<T>): T {
	return typeof initialValue === "function"
		? (initialValue as () => T)()
		: initialValue;
}

function readValue<T>(
	key: string,
	initialValue: InitialValue<T>,
	options: LocalStorageStateOptions<T>,
): T {
	const fallback = resolveInitialValue(initialValue);
	if (typeof window === "undefined") return fallback;

	try {
		const raw = window.localStorage.getItem(key);
		if (raw === null) return fallback;
		return options.parse
			? (options.parse(raw) ?? fallback)
			: (JSON.parse(raw) as T);
	} catch {
		return fallback;
	}
}

/** A typed state setter that persists JSON values without making storage global state. */
export function useLocalStorageState<T>(
	key: string,
	initialValue: InitialValue<T>,
	options: LocalStorageStateOptions<T> = {},
): [T, Dispatch<SetStateAction<T>>] {
	const [value, setValue] = useState<T>(() =>
		readValue(key, initialValue, options),
	);

	const setStoredValue = useCallback<Dispatch<SetStateAction<T>>>(
		(nextValue) => {
			setValue((previousValue) => {
				const resolvedValue =
					typeof nextValue === "function"
						? (nextValue as (value: T) => T)(previousValue)
						: nextValue;

				try {
					if (typeof window === "undefined") return resolvedValue;
					window.localStorage.setItem(
						key,
						options.serialize?.(resolvedValue) ?? JSON.stringify(resolvedValue),
					);
				} catch {
					// Persisted UI preferences are optional when storage is unavailable.
				}

				return resolvedValue;
			});
		},
		[key, options],
	);

	return [value, setStoredValue];
}
