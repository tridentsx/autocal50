import { useCallback, useEffect, useRef, useState } from "react";
import { motion } from "framer-motion";
import { api } from "@/api";
import ConvergenceRing from "../ConvergenceRing";
import AdjustmentCard from "../AdjustmentCard";
import MeasureButton from "../MeasureButton";

const COLORS = [
  { label: "Red", r: 255, g: 0, b: 0 },
  { label: "Green", r: 0, g: 255, b: 0 },
  { label: "Blue", r: 0, g: 0, b: 255 },
  { label: "Cyan", r: 0, g: 255, b: 255 },
  { label: "Magenta", r: 255, g: 0, b: 255 },
  { label: "Yellow", r: 255, g: 255, b: 0 },
];

type Advice = { deltaE: number; passed: boolean; adjustments: any[] };

/** CMS step: per-color adjustment with convergence feedback. */
export default function CMSStep({ step, mode, onComplete }: { step: any; mode: string; standard: string; onComplete: () => void }) {
  const [colorIdx, setColorIdx] = useState(0);
  const [advice, setAdvice] = useState<Advice | null>(null);
  const [results, setResults] = useState<{ label: string; deltaE: number; passed: boolean }[]>([]);
  const auto = mode === "automatic";
  const intervalRef = useRef<ReturnType<typeof setInterval> | undefined>(undefined);

  const color = COLORS[colorIdx];
  const target = step.targets?.[colorIdx] ?? { x: 0.3127, y: 0.329, label: color.label };
  const tolerance = step.tolerance || 2;

  const doMeasure = useCallback(async () => {
    // Set color pattern
    await api.setActivePattern({
      id: `cms-${color.label.toLowerCase()}`, name: color.label, group: "Primaries",
      type: "solid", color: { r: color.r, g: color.g, b: color.b },
    });
    await new Promise((r) => setTimeout(r, 1000));

    const adv = await api.evaluateMeasurement(target, tolerance);
    setAdvice(adv);

    if (adv.passed) {
      setResults((r) => [...r, { label: color.label, deltaE: adv.deltaE, passed: true }]);
      if (colorIdx < COLORS.length - 1) {
        setColorIdx((i) => i + 1);
        setAdvice(null);
      } else {
        onComplete();
      }
    }
  }, [color, target, tolerance, colorIdx, onComplete]);

  useEffect(() => {
    if (auto) {
      intervalRef.current = setInterval(doMeasure, 2500);
      return () => clearInterval(intervalRef.current);
    }
  }, [auto, doMeasure]);

  return (
    <div>
      {/* Color tabs */}
      <div className="flex gap-1 mb-4">
        {COLORS.map((c, i) => {
          const done = results.some((r) => r.label === c.label);
          const active = i === colorIdx;
          return (
            <div key={c.label} className="flex items-center gap-1 px-2 py-1 rounded text-xs"
              style={{
                background: active ? "var(--accent)" : done ? "#22c55e22" : "var(--surface)",
                border: "1px solid var(--border)",
                color: active ? "#fff" : "var(--muted)",
              }}>
              <span style={{ color: `rgb(${c.r},${c.g},${c.b})` }}>●</span>
              {c.label}
              {done && <span style={{ color: "#22c55e" }}>✓</span>}
            </div>
          );
        })}
      </div>

      <div className="flex gap-6 items-start">
        <ConvergenceRing deltaE={advice?.deltaE ?? 10} tolerance={tolerance} />

        <div className="flex-1 flex flex-col gap-4">
          <motion.div
            key={color.label}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-sm"
          >
            Measuring <span className="font-bold" style={{ color: `rgb(${color.r},${color.g},${color.b})` }}>{color.label}</span>
            <span className="ml-2" style={{ color: "var(--muted)" }}>
              target x={target.x.toFixed(4)} y={target.y.toFixed(4)}
            </span>
          </motion.div>

          {advice && <AdjustmentCard adjustments={advice.adjustments} auto={auto} />}
          {!auto && <MeasureButton onMeasure={doMeasure} />}
        </div>
      </div>
    </div>
  );
}
