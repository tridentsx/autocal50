import { motion } from "framer-motion";
import { useAnimatedValue } from "./useAnimatedValue";

const SIZE = 140;
const STROKE = 8;
const R = (SIZE - STROKE) / 2;
const CIRC = 2 * Math.PI * R;

function deColor(de: number, tolerance: number): string {
  if (de <= tolerance) return "#22c55e";
  if (de <= tolerance * 2) return "#f59e0b";
  return "#ef4444";
}

/** Animated ring showing ΔE convergence toward target. */
export default function ConvergenceRing({ deltaE, tolerance }: { deltaE: number; tolerance: number }) {
  const animDE = useAnimatedValue(deltaE);
  // Map ΔE to progress: 0 at 5× tolerance, 1 at 0
  const progress = Math.max(0, Math.min(1, 1 - animDE / (tolerance * 5)));
  const color = deColor(animDE, tolerance);
  const passed = animDE <= tolerance;

  return (
    <div className="relative flex items-center justify-center" style={{ width: SIZE, height: SIZE }}>
      <svg width={SIZE} height={SIZE} className="absolute">
        {/* Background track */}
        <circle cx={SIZE / 2} cy={SIZE / 2} r={R} fill="none" stroke="var(--border)" strokeWidth={STROKE} />
        {/* Animated arc */}
        <motion.circle
          cx={SIZE / 2} cy={SIZE / 2} r={R} fill="none"
          stroke={color} strokeWidth={STROKE} strokeLinecap="round"
          strokeDasharray={CIRC}
          strokeDashoffset={CIRC * (1 - progress)}
          transform={`rotate(-90 ${SIZE / 2} ${SIZE / 2})`}
          animate={{ strokeDashoffset: CIRC * (1 - progress) }}
          transition={{ type: "spring", stiffness: 60, damping: 15 }}
        />
      </svg>
      <motion.div
        className="text-center z-10"
        animate={passed ? { scale: [1, 1.1, 1] } : {}}
        transition={{ duration: 0.4 }}
      >
        <div className="text-2xl font-bold font-mono" style={{ color }}>{animDE.toFixed(1)}</div>
        <div className="text-xs" style={{ color: "var(--muted)" }}>ΔE · target ≤{tolerance}</div>
      </motion.div>
    </div>
  );
}
