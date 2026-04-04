import { useCallback, useEffect, useRef, useState } from "react";
import { api } from "@/api";
import ConvergenceRing from "../ConvergenceRing";
import LiveCrosshair from "../LiveCrosshair";
import AdjustmentCard from "../AdjustmentCard";
import MeasureButton from "../MeasureButton";

type Point = { x: number; y: number };
type Advice = { deltaE: number; deltaX: number; deltaY: number; passed: boolean; adjustments: any[] };

/** White balance step: interactive gain/bias adjustment with live feedback. */
export default function WhiteBalanceStep({ step, mode, onComplete }: { step: any; mode: string; standard: string; onComplete: () => void }) {
  const [advice, setAdvice] = useState<Advice | null>(null);
  const [measured, setMeasured] = useState<Point>({ x: 0.33, y: 0.33 });
  const [trail, setTrail] = useState<Point[]>([]);
  const [targetIdx, setTargetIdx] = useState(0);
  const auto = mode === "automatic";
  const intervalRef = useRef<ReturnType<typeof setInterval> | undefined>(undefined);

  const target = step.targets?.[targetIdx] ?? { x: 0.3127, y: 0.329, label: "D65" };
  const tolerance = step.tolerance || 1.5;

  const doMeasure = useCallback(async () => {
    const adv = await api.evaluateMeasurement(target, tolerance);
    setAdvice(adv);
    const pt = { x: target.x + adv.deltaX, y: target.y + adv.deltaY };
    setMeasured(pt);
    setTrail((t) => [...t.slice(-20), pt]);

    if (adv.passed) {
      // Move to next target or complete
      if (targetIdx < (step.targets?.length ?? 1) - 1) {
        setTargetIdx((i) => i + 1);
        setTrail([]);
        setAdvice(null);
      } else {
        onComplete();
      }
    }
  }, [target, tolerance, targetIdx, step.targets, onComplete]);

  // Auto-measure loop for automatic mode
  useEffect(() => {
    if (auto) {
      intervalRef.current = setInterval(doMeasure, 2000);
      return () => clearInterval(intervalRef.current);
    }
  }, [auto, doMeasure]);

  return (
    <div>
      <div className="text-sm mb-4 px-1 py-2 rounded" style={{ background: "var(--surface)", border: "1px solid var(--border)" }}>
        Target: <span className="font-mono font-bold">{target.label}</span>
        <span className="ml-3" style={{ color: "var(--muted)" }}>
          x={target.x.toFixed(4)} y={target.y.toFixed(4)}
        </span>
      </div>

      <div className="flex gap-6 items-start">
        {/* Left: ring + crosshair */}
        <div className="flex flex-col items-center gap-4">
          <ConvergenceRing deltaE={advice?.deltaE ?? 10} tolerance={tolerance} />
          <LiveCrosshair
            target={{ x: target.x, y: target.y }}
            measured={measured}
            trail={trail}
          />
        </div>

        {/* Right: advice + button */}
        <div className="flex-1 flex flex-col gap-4">
          {advice && <AdjustmentCard adjustments={advice.adjustments} auto={auto} />}

          {!auto && (
            <MeasureButton onMeasure={doMeasure} />
          )}

          {advice && (
            <div className="text-xs font-mono" style={{ color: "var(--muted)" }}>
              Δx={advice.deltaX > 0 ? "+" : ""}{advice.deltaX.toFixed(4)}{" "}
              Δy={advice.deltaY > 0 ? "+" : ""}{advice.deltaY.toFixed(4)}{" "}
              ΔE={advice.deltaE.toFixed(2)}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
