import { motion, AnimatePresence } from "framer-motion";

type Adj = { controlId: string; direction: string; magnitude: string; reason: string };

/** Animated adjustment advice cards with stagger-in. */
export default function AdjustmentCard({ adjustments, auto }: { adjustments: Adj[]; auto?: boolean }) {
  if (!adjustments.length) return null;

  return (
    <div className="flex flex-col gap-2">
      <AnimatePresence mode="popLayout">
        {adjustments.map((a, i) => (
          <motion.div
            key={a.controlId + a.direction}
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: 20 }}
            transition={{ delay: i * 0.08, type: "spring", stiffness: 200, damping: 20 }}
            className="rounded-lg p-3"
            style={{ background: "var(--surface)", border: "1px solid var(--border)" }}
          >
            <div className="flex items-center gap-2">
              <motion.span
                className="text-lg"
                animate={{ y: a.direction === "increase" ? [0, -3, 0] : [0, 3, 0] }}
                transition={{ repeat: Infinity, duration: 1.5 }}
              >
                {a.direction === "increase" ? "↑" : "↓"}
              </motion.span>
              <div>
                <div className="text-sm font-medium">
                  {auto ? (a.direction === "increase" ? "Increasing" : "Decreasing") : (a.direction === "increase" ? "Increase" : "Decrease")}{" "}
                  <span style={{ color: "var(--accent)" }}>{a.controlId.replace(/_/g, " ")}</span>
                  {auto && <span className="ml-1 text-xs" style={{ color: "var(--muted)" }}>…</span>}
                </div>
                <div className="text-xs" style={{ color: "var(--muted)" }}>{a.magnitude} — {a.reason}</div>
              </div>
            </div>
          </motion.div>
        ))}
      </AnimatePresence>
    </div>
  );
}
