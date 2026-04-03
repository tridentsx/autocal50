import { useLiveReading } from "@/hooks/useWailsEvent";
import ReactECharts from "echarts-for-react";

export default function LiveView() {
  const reading = useLiveReading();

  if (!reading) return <div style={{ color: "var(--muted)" }}>Waiting for measurements…</div>;

  const stats = [
    { label: "x", value: reading.x.toFixed(4) },
    { label: "y", value: reading.y.toFixed(4) },
    { label: "Luminance", value: `${reading.luminance.toFixed(1)} cd/m²` },
    { label: "CCT", value: `${reading.cct.toFixed(0)} K` },
  ];

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
        {stats.map((s) => (
          <div key={s.label} className="rounded-lg p-4 text-center" style={{ background: "var(--surface)", border: "1px solid var(--border)" }}>
            <div className="text-sm mb-1" style={{ color: "var(--muted)" }}>{s.label}</div>
            <div className="text-xl font-mono">{s.value}</div>
          </div>
        ))}
      </div>
      <div className="rounded-lg p-4" style={{ background: "var(--surface)", border: "1px solid var(--border)" }}>
        <ReactECharts option={gaugeOption} style={{ height: 300 }} />
      </div>
    </div>
  );
}
