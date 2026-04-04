import { motion } from "framer-motion";
import { useLiveReading } from "@/hooks/useWailsEvent";
import { useAnimatedValue } from "@/components/cal/useAnimatedValue";
import ReactECharts from "echarts-for-react";

function AnimatedStat({ label, value, unit }: { label: string; value: number; unit?: string }) {
  const animated = useAnimatedValue(value);
  const decimals = label === "CCT" ? 0 : label === "Luminance" ? 1 : 4;
  return (
    <motion.div
      initial={{ opacity: 0, y: 15 }}
      animate={{ opacity: 1, y: 0 }}
      className="rounded-lg p-4 text-center"
      style={{ background: "var(--surface)", border: "1px solid var(--border)" }}
    >
      <div className="text-sm mb-1" style={{ color: "var(--muted)" }}>{label}</div>
      <div className="text-xl font-mono">{animated.toFixed(decimals)}{unit ? ` ${unit}` : ""}</div>
    </motion.div>
  );
}

export default function LiveView() {
  const reading = useLiveReading();

  if (!reading) return <div style={{ color: "var(--muted)" }}>Waiting for measurements…</div>;

  const gaugeOption = {
    backgroundColor: "transparent",
    series: [
      {
        type: "gauge",
        startAngle: 200,
        endAngle: -20,
        min: 0,
        max: 10000,
        detail: { formatter: "{value} K", fontSize: 16, color: "#e2e4e9", offsetCenter: [0, "60%"] },
        data: [{ value: Math.round(reading.cct), name: "CCT" }],
        axisLine: { lineStyle: { width: 12, color: [[0.3, "#3b82f6"], [0.65, "#22c55e"], [1, "#ef4444"]] } },
        pointer: { length: "60%" },
        title: { color: "#8b8fa3" },
      },
    ],
  };

  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">Live View</h2>
      <div className="grid grid-cols-4 gap-4 mb-6">
        <AnimatedStat label="x" value={reading.x} />
        <AnimatedStat label="y" value={reading.y} />
        <AnimatedStat label="Luminance" value={reading.luminance} unit="cd/m²" />
        <AnimatedStat label="CCT" value={reading.cct} unit="K" />
      </div>
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ delay: 0.2 }}
        className="rounded-lg p-4"
        style={{ background: "var(--surface)", border: "1px solid var(--border)" }}
      >
        <ReactECharts option={gaugeOption} style={{ height: 300 }} />
      </motion.div>
    </div>
  );
}
