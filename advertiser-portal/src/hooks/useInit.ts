import { useEffect } from "react";
import { store } from "@/lib/store";

export function useInit() {
  useEffect(() => {
    store.refresh();
  }, []);
}
