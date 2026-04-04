import { motion } from "framer-motion";
import { useAnimatedValue } from "./useAnimatedValue";

type Result = {
  avgDeltaE: number;
  maxDeltaE: number;
  gammaAvg: number;
  whiteCCT: number;
  gamutCoverage: number;
  standard: string;
};

function stars(avgDE: number): number {
  if (avgDE < 1) return 5;
  if (avgDE < 2) return 4;
  if (avgDE < 3) return 3;
  if (avgDE < 5) return 2;
  return 1;
}

function rating(n: number): string {
  return ["", "Poor", "Fair", "Good", "Excellent", "Reference"][n];
}

/** Animated calibration report scorecard. */
export default function ScoreCard({ result }: { result: Result }) {
  const avgDE = useAnimatedValue(result.avgDeltaE);
  const maxDE = useAnimatedValue(result.maxDeltaE);
  const gamut = useAnimatedValue(result.gamutCoverage);
  const s = stars(result.avgDeltaE);

  const row = (label: string, value: string, pass: boolean) => (
    <div className="flex justify-between py-1">
      <span style={{ color: "var(--muted)" }}>{label}</span>
      <span style={{ color: pass ? "#22c55e" : "#f59e0b" }}>{value}</span>
    </div>
  );

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ type: "spring", stiffness: 100, damping: 15 }}
      className="rounded-xl p-6 max-w-md mx-auto"
      style={{ background: "var(--surface)", border: "1px solid var(--border)" }}
    >
      <h3 className="text-lg font-bold text-center mb-4">Calibration Report</h3>

      <div className="text-sm">
        {row("Average ΔE", avgDE.toFixed(1), result.avgDeltaE < 2)}
        {row("Max ΔE", maxDE.toFixed(1), result.maxDeltaE < 3)}
        {row("Gamma", result.gammaAvg.toFixed(2), Math.abs(result.gammaAvg - 2.2) < 0.1)}
        {row("White Point", `${result.whiteCCT.toFixed(0)} K`, Math.abs(result.whiteCCT - 6500) < 200)}
        {row("Gamut Coverage", `${gamut.toFixed(1)}% ${result.standard}`, result.gamutCoverage > 95)}
      </div>

      {/* Star rating */}
      <div className="flex justify-center gap-1 mt-4">
        {[1, 2, 3, 4, 5].map((i) => (
          <motion.span
            key={i}
            initial={{ scale: 0, rotate: -30 }}
            animate={{ scale: 1, rotate: 0 }}
            transition={{ delay: 0.3 + i * 0.12, type: "spring", stiffness: 300, damping: 12 }}
            className="text-xl"
            style={{ color: i <= s ? "#facc15" : "var(--border)" }}
          >
            ★
          </motion.span>
        ))}
      </div>
      <motion.div
        className="text-center text-sm font-semibold mt-1"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 1 }}
        style={{ color: s >= 4 ? "#22c55e" : s >= 3 ? "#facc15" : "#ef4444" }}
      >
        {rating(s)}
      </motion.div>
    </motion.div>
  );
}
