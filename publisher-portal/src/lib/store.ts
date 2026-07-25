// Legacy no-op store kept only for backwards compatibility with any lingering
// imports. The publisher portal reads all data from the REST API via
// TanStack Query — see src/lib/api.ts.
import { useSyncExternalStore } from "react";

interface State {
  balance: number;
}

let state: State = { balance: 0 };
const listeners = new Set<() => void>();

export const store = {
  get: () => state,
  subscribe: (fn: () => void) => {
    listeners.add(fn);
    return () => listeners.delete(fn);
  },
  setBalance(n: number) {
    state = { ...state, balance: n };
    listeners.forEach((l) => l());
  },
};

export function useStore<T>(selector: (s: State) => T): T {
  return useSyncExternalStore(
    store.subscribe,
    () => selector(store.get()),
    () => selector(store.get()),
  );
}
