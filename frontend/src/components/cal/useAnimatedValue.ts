import { useEffect, useRef, useState } from "react";

/** Smoothly interpolates a numeric value over time. */
export function useAnimatedValue(target: number, duration = 400) {
  const [value, setValue] = useState(target);
  const raf = useRef(0);
  const start = useRef({ value: target, time: 0, target });

  useEffect(() => {
    const s = start.current;
    s.value = value;
    s.target = target;
    s.time = performance.now();

    const tick = (now: number) => {
      const t = Math.min((now - s.time) / duration, 1);
      const eased = t < 1 ? t * (2 - t) : 1; // ease-out quad
      setValue(s.value + (s.target - s.value) * eased);
      if (t < 1) raf.current = requestAnimationFrame(tick);
    };
    raf.current = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(raf.current);
  }, [target, duration]);

  return value;
}
