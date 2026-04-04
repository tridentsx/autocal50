import { motion } from "framer-motion";

type Step = { label: string; kind: string };

/** Animated step progress bar with connected dots. */
export default function StepProgress({ steps, current }: { steps: Step[]; current: number }) {
  return (
    <div className="flex items-center gap-0 w-full mb-6">
      {steps.map((s, i) => {
        const done = i < current;
        const active = i === current;
        const color = done ? "#22c55e" : active ? "var(--accent)" : "var(--border)";

        return (
          <div key={s.kind} className="flex items-center" style={{ flex: i < steps.length - 1 ? 1 : 0 }}>
            {/* Dot */}
            <motion.div
              className="relative flex items-center justify-center rounded-full shrink-0"
              style={{ width: 32, height: 32, border: `2px solid ${color}`, background: done ? "#22c55e" : "transparent" }}
              animate={active ? { scale: [1, 1.1, 1] } : {}}
              transition={active ? { repeat: Infinity, duration: 2 } : {}}
            >
              {done ? (
                <motion.span initial={{ scale: 0 }} animate={{ scale: 1 }}
                  transition={{ type: "spring", stiffness: 300, damping: 15 }}
                  className="text-white text-sm">✓</motion.span>
              ) : (
                <span className="text-xs font-bold" style={{ color }}>{i + 1}</span>
              )}
              {/* Label below */}
              <span className="absolute -bottom-5 text-[10px] whitespace-nowrap"
                style={{ color: active ? "var(--text)" : "var(--muted)" }}>{s.label}</span>
            </motion.div>

            {/* Connector line */}
            {i < steps.length - 1 && (
              <div className="flex-1 h-0.5 mx-1" style={{ background: "var(--border)" }}>
                <motion.div
                  className="h-full"
                  style={{ background: "#22c55e" }}
                  initial={{ width: "0%" }}
                  animate={{ width: done ? "100%" : "0%" }}
                  transition={{ duration: 0.4 }}
                />
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}
