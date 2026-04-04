import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";

type State = "idle" | "measuring" | "done" | "error";

/** Animated measure button: idle → spinner → check/error. */
export default function MeasureButton({ onMeasure }: { onMeasure: () => Promise<void> }) {
  const [state, setState] = useState<State>("idle");

  const handleClick = async () => {
    setState("measuring");
    try {
      await onMeasure();
      setState("done");
      setTimeout(() => setState("idle"), 1200);
    } catch {
      setState("error");
      setTimeout(() => setState("idle"), 1500);
    }
  };

  return (
    <motion.button
      onClick={handleClick}
      disabled={state === "measuring"}
      className="relative px-8 py-3 rounded-xl font-semibold text-white overflow-hidden"
      style={{ background: state === "error" ? "#ef4444" : "var(--accent)", minWidth: 160 }}
      whileTap={{ scale: 0.95 }}
      animate={state === "error" ? { x: [0, -6, 6, -4, 4, 0] } : {}}
      transition={{ duration: 0.4 }}
    >
      <AnimatePresence mode="wait">
        {state === "idle" && (
          <motion.span key="idle" initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}>
            Measure
          </motion.span>
        )}
        {state === "measuring" && (
          <motion.span key="spin" initial={{ opacity: 0 }} animate={{ opacity: 1, rotate: 360 }} exit={{ opacity: 0 }}
            transition={{ rotate: { repeat: Infinity, duration: 0.8, ease: "linear" } }}
            className="inline-block">
            ◌
          </motion.span>
        )}
        {state === "done" && (
          <motion.span key="done" initial={{ scale: 0 }} animate={{ scale: 1 }}
            transition={{ type: "spring", stiffness: 300, damping: 15 }}>
            ✓
          </motion.span>
        )}
        {state === "error" && (
          <motion.span key="err" initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
            Error
          </motion.span>
        )}
      </AnimatePresence>
    </motion.button>
  );
}
