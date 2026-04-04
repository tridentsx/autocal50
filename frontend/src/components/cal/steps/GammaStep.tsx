import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { api } from "@/api";
import MeasureButton from "../MeasureButton";

type GammaPoint = { input: number; target: number; measured: number };

/** Gamma/EOTF step: sweep IRE levels and build animated tracking chart. */
export default function GammaStep({ mode, onComplete }: { step: any; mode: string; standard: string; onComplete: () => void }) {
  const [points, setPoints] = useState<GammaPoint[]>([]);
  const [running, setRunning] = useState(false);
  const steps = 20;

  const runSweep = async () => {
    setRunning(true);
    setPoints([]);
    const results: GammaPoint[] = [];

    for (let i = 0; i <= steps; i++) {
      const pct = i / steps;
      const ire = Math.round(pct * 100);
      const v = Math.round(pct * 255);

      await api.setActivePattern({
        id: `gamma-${ire}`, name: `${ire}% IRE`, group: "Grayscale Windows",
        type: "window", color: { r: v, g: v, b: v }, bgColor: { r: 0, g: 0, b: 0 }, windowPc: 18,
      });
      await sleep(1500);

      try {
        const reading = await api.measure();
        // Normalize measured luminance to 0-1 range (relative to peak)
        const peak = results.length > 0 ? Math.max(...results.map((r) => r.measured), reading.luminance) : reading.luminance;
        const measured = peak > 0 ? reading.luminance / peak : 0;
        const target = Math.pow(pct, 2.2); // default gamma 2.2

        const pt = { input: pct, target, measured };
        results.push(pt);
        setPoints([...results]);
      } catch {
        results.push({ input: pct, target: Math.pow(pct, 2.2), measured: 0 });
        setPoints([...results]);
      }
    }
    setRunning(false);
    if (mode === "automatic") onComplete();
  };

  useEffect(() => {
    if (mode === "automatic") runSweep();
  }, [mode]);

  // Simple SVG chart
  const W = 500, H = 300, PAD = 40;
  const toX = (v: number) => PAD + v * (W - PAD * 2);
  const toY = (v: number) => H - PAD - v * (H - PAD * 2);

  const targetPath = Array.from({ length: 21 }, (_, i) => {
    const inp = i / 20;
    return `${i === 0 ? "M" : "L"}${toX(inp)},${toY(Math.pow(inp, 2.2))}`;
  }).join(" ");

  const measuredPath = points.map((p, i) =>
    `${i === 0 ? "M" : "L"}${toX(p.input)},${toY(p.measured)}`
  ).join(" ");

  return (
    <div>
      {mode === "assisted" && !points.length && !running && (
        <MeasureButton onMeasure={runSweep} />
      )}
      {running && <div className="text-sm mb-2" style={{ color: "var(--muted)" }}>Sweeping {points.length}/{steps + 1} levels…</div>}

      <svg width={W} height={H} className="mt-4">
        {/* Grid */}
        {[0, 0.25, 0.5, 0.75, 1].map((v) => (
          <g key={v}>
            <line x1={toX(v)} y1={toY(0)} x2={toX(v)} y2={toY(1)} stroke="#2a2d3a" strokeWidth={0.5} />
            <line x1={toX(0)} y1={toY(v)} x2={toX(1)} y2={toY(v)} stroke="#2a2d3a" strokeWidth={0.5} />
            <text x={toX(v)} y={H - 10} textAnchor="middle" fill="#8b8fa3" fontSize={10}>{Math.round(v * 100)}%</text>
          </g>
        ))}

        {/* Target curve (dashed) */}
        <path d={targetPath} fill="none" stroke="#555" strokeWidth={1.5} strokeDasharray="4 3" />

        {/* Measured curve (animated) */}
        {points.length > 1 && (
          <motion.path
            d={measuredPath}
            fill="none" stroke="#6366f1" strokeWidth={2}
            initial={{ pathLength: 0 }}
            animate={{ pathLength: 1 }}
            transition={{ duration: 0.5 }}
          />
        )}

        {/* Measured dots */}
        {points.map((p, i) => (
          <motion.circle
            key={i}
            cx={toX(p.input)} cy={toY(p.measured)} r={4}
            fill="#6366f1"
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            transition={{ type: "spring", stiffness: 300, damping: 15 }}
          />
        ))}

        {/* Labels */}
        <text x={W / 2} y={H - 0} textAnchor="middle" fill="#8b8fa3" fontSize={11}>Input (IRE %)</text>
        <text x={12} y={H / 2} textAnchor="middle" fill="#8b8fa3" fontSize={11} transform={`rotate(-90 12 ${H / 2})`}>Output</text>
      </svg>
    </div>
  );
}

function sleep(ms: number) { return new Promise((r) => setTimeout(r, ms)); }
