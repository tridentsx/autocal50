import { useCallback, useEffect, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { api } from "@/api";
import StepProgress from "./StepProgress";
import PreCalStep from "./steps/PreCalStep";
import WhiteBalanceStep from "./steps/WhiteBalanceStep";
import GammaStep from "./steps/GammaStep";
import CMSStep from "./steps/CMSStep";
import VerifyStep from "./steps/VerifyStep";

type CalStep = { kind: string; label: string; description: string; targets: any[]; tolerance: number };

const stepComponents: Record<string, React.FC<any>> = {
  precal: PreCalStep,
  whitebalance: WhiteBalanceStep,
  gamma: GammaStep,
  cms: CMSStep,
  verify: VerifyStep,
};

const slideVariants = {
  enter: (dir: number) => ({ x: dir > 0 ? 300 : -300, opacity: 0 }),
  center: { x: 0, opacity: 1 },
  exit: (dir: number) => ({ x: dir > 0 ? -300 : 300, opacity: 0 }),
};

export default function CalibrationWizard({ mode, standard }: { mode: "assisted" | "automatic"; standard: string }) {
  const [steps, setSteps] = useState<CalStep[]>([]);
  const [current, setCurrent] = useState(0);
  const [dir, setDir] = useState(1);
  const [error, setError] = useState("");

  useEffect(() => {
    api.getCalibrationSteps(standard).then(setSteps).catch((e) => setError(String(e)));
  }, [standard]);

  const next = useCallback(() => {
    setDir(1);
    setCurrent((c) => Math.min(c + 1, steps.length - 1));
  }, [steps.length]);

  const back = useCallback(() => {
    setDir(-1);
    setCurrent((c) => Math.max(c - 1, 0));
  }, []);

  if (error) return <div style={{ color: "#ef4444" }}>{error}</div>;
  if (!steps.length) return <div style={{ color: "var(--muted)" }}>Loading calibration steps…</div>;

  const step = steps[current];
  const StepComponent = stepComponents[step.kind] ?? PreCalStep;
  const isFirst = current === 0;
  const isLast = current === steps.length - 1;
  const auto = mode === "automatic";

  return (
    <div className="flex flex-col h-full">
      {/* Progress bar */}
      <div className="pt-2 px-4">
        <StepProgress steps={steps} current={current} />
      </div>

      {/* Step content */}
      <div className="flex-1 overflow-hidden relative mt-6">
        <AnimatePresence mode="wait" custom={dir}>
          <motion.div
            key={step.kind + current}
            custom={dir}
            variants={slideVariants}
            initial="enter"
            animate="center"
            exit="exit"
            transition={{ type: "spring", stiffness: 200, damping: 25 }}
            className="absolute inset-0 px-4 overflow-y-auto"
          >
            {/* Step header */}
            <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.1 }}>
              <h3 className="text-xl font-bold mb-1">{step.label}</h3>
              <p className="text-sm mb-4" style={{ color: "var(--muted)" }}>{step.description}</p>
            </motion.div>

            <StepComponent
              step={step}
              mode={mode}
              standard={standard}
              onComplete={next}
            />
          </motion.div>
        </AnimatePresence>
      </div>

      {/* Navigation */}
      {!auto && (
        <div className="flex items-center justify-between px-4 py-3" style={{ borderTop: "1px solid var(--border)" }}>
          <button onClick={back} disabled={isFirst}
            className="px-4 py-2 rounded text-sm"
            style={{ color: isFirst ? "var(--border)" : "var(--muted)", background: "transparent" }}>
            ← Back
          </button>
          <span className="text-xs" style={{ color: "var(--muted)" }}>
            Step {current + 1} of {steps.length}
          </span>
          <button onClick={next} disabled={isLast}
            className="px-4 py-2 rounded text-sm font-medium"
            style={{ background: isLast ? "var(--border)" : "var(--accent)", color: "#fff" }}>
            {isLast ? "Done" : "Next →"}
          </button>
        </div>
      )}
    </div>
  );
}
