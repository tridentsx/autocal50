import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { api } from "@/api";
import MeasureButton from "../MeasureButton";

type Check = { label: string; value: string; pass: boolean; warn?: boolean };

/** Pre-calibration checks: automated sweep measuring baseline performance. */
export default function PreCalStep({ mode, onComplete }: { step: any; mode: string; standard: string; onComplete: () => void }) {
  const [checks, setChecks] = useState<Check[]>([]);
  const [running, setRunning] = useState(false);

  const runChecks = async () => {
    setRunning(true);
    setChecks([]);
    const results: Check[] = [];

    const addCheck = (c: Check) => {
      results.push(c);
      setChecks([...results]);
    };

    try {
      // Measure black (0% IRE)
      await api.setActivePattern({ id: "gray-win-0", name: "0% Window", group: "Grayscale Windows", type: "window", color: { r: 0, g: 0, b: 0 }, bgColor: { r: 0, g: 0, b: 0 }, windowPc: 18 });
      await sleep(2000);
      const black = await api.measure();
      addCheck({ label: "Black Level", value: `${black.luminance.toFixed(3)} cd/m²`, pass: black.luminance < 0.05, warn: black.luminance >= 0.03 });

      // Measure white (100% window)
      await api.setActivePattern({ id: "gray-win-100", name: "100% Window", group: "Grayscale Windows", type: "window", color: { r: 255, g: 255, b: 255 }, bgColor: { r: 0, g: 0, b: 0 }, windowPc: 18 });
      await sleep(2000);
      const white = await api.measure();
      addCheck({ label: "Peak Luminance (18% win)", value: `${white.luminance.toFixed(0)} cd/m²`, pass: true });

      // Contrast ratio
      const cr = white.luminance / Math.max(black.luminance, 0.001);
      addCheck({ label: "Contrast Ratio", value: cr > 9999 ? "∞:1" : `${cr.toFixed(0)}:1`, pass: cr > 500 });

      // White point
      addCheck({ label: "White Point", value: `x=${white.x.toFixed(4)} y=${white.y.toFixed(4)}`, pass: Math.abs(white.x - 0.3127) < 0.01 && Math.abs(white.y - 0.329) < 0.01 });

      // CCT
      addCheck({ label: "CCT", value: `${white.cct.toFixed(0)} K`, pass: Math.abs(white.cct - 6500) < 500, warn: Math.abs(white.cct - 6500) >= 300 });
    } catch (e: any) {
      addCheck({ label: "Error", value: e?.message || String(e), pass: false });
    }
    setRunning(false);
  };

  // Auto-run in automatic mode
  useEffect(() => {
    if (mode === "automatic") {
      runChecks().then(onComplete);
    }
  }, [mode]);

  return (
    <div>
      {mode === "assisted" && !checks.length && !running && (
        <MeasureButton onMeasure={runChecks} />
      )}
      {running && <div className="text-sm mb-4" style={{ color: "var(--muted)" }}>Measuring baseline…</div>}

      <div className="grid grid-cols-1 gap-2 mt-4">
        {checks.map((c, i) => (
          <motion.div
            key={c.label}
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: i * 0.15, type: "spring", stiffness: 200, damping: 20 }}
            className="flex items-center gap-3 rounded-lg p-3"
            style={{ background: "var(--surface)", border: "1px solid var(--border)" }}
          >
            <motion.span
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              transition={{ delay: i * 0.15 + 0.1, type: "spring", stiffness: 300, damping: 15 }}
              className="text-lg"
              style={{ color: c.pass ? (c.warn ? "#facc15" : "#22c55e") : "#ef4444" }}
            >
              {c.pass ? (c.warn ? "⚠" : "✓") : "✗"}
            </motion.span>
            <div className="flex-1">
              <div className="text-sm font-medium">{c.label}</div>
              <div className="text-xs" style={{ color: "var(--muted)" }}>{c.value}</div>
            </div>
          </motion.div>
        ))}
      </div>
    </div>
  );
}

function sleep(ms: number) { return new Promise((r) => setTimeout(r, ms)); }
