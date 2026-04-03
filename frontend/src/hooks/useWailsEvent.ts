import { useEffect, useState } from "react";
import type { Reading } from "@/types";

declare global {
  interface Window {
    runtime: {
      EventsOn(event: string, cb: (...args: any[]) => void): () => void;
    };
  }
}

export function useWailsEvent<T = any>(event: string) {
  const [data, setData] = useState<T | null>(null);
  useEffect(() => {
    const cancel = window.runtime?.EventsOn(event, (d: T) => setData(d));
    return () => cancel?.();
  }, [event]);
  return data;
}

export function useLiveReading() {
  return useWailsEvent<Reading>("measurement:update");
}
