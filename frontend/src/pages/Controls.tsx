import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { api } from "@/api";
import type { Control } from "@/types";

function ControlWidget({ ctrl, onChange }: { ctrl: Control & { value: any }; onChange: (v: any) => void }) {
  if (ctrl.type === "slider")
    return (
      <div>
        <input type="range" min={ctrl.min} max={ctrl.max} step={ctrl.step}
          value={ctrl.value ?? ctrl.min} onChange={(e) => onChange(+e.target.value)}
          disabled={ctrl.readOnly} className="w-full" />
        <span className="text-sm" style={{ color: "var(--muted)" }}>{ctrl.value}</span>
      </div>
    );
  if (ctrl.type === "enum")
    return (
      <select value={ctrl.value ?? ""} onChange={(e) => onChange(e.target.value)} disabled={ctrl.readOnly}
        className="w-full p-2 rounded" style={{ background: "var(--bg)", color: "var(--text)", border: "1px solid var(--border)" }}>
        {ctrl.options.map((o) => <option key={o}>{o}</option>)}
      </select>
    );
  if (ctrl.type === "toggle")
    return (
      <motion.button onClick={() => onChange(!ctrl.value)} disabled={ctrl.readOnly}
        className="px-4 py-1 rounded text-sm"
        animate={{ background: ctrl.value ? "#22c55e" : "var(--bg)" }}
        transition={{ duration: 0.2 }}
        style={{ color: "var(--text)", border: "1px solid var(--border)" }}>
        {ctrl.value ? "ON" : "OFF"}
      </motion.button>
    );
  return null;
}

export default function Controls() {
  const [controls, setControls] = useState<(Control & { value: any })[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    api.getCapabilities().then(async (caps: any) => {
      const loaded = await Promise.all(
        (caps.controls as Control[]).map(async (c) => {
          const value = await api.getControl(c.id).catch(() => c.type === "toggle" ? false : c.min);
          return { ...c, value };
        })
      );
      setControls(loaded);
    }).catch((e) => setError(String(e)));
  }, []);

  const handleChange = async (id: string, value: any) => {
    try {
      await api.setControl(id, value);
      setControls((prev) => prev.map((c) => (c.id === id ? { ...c, value } : c)));
    } catch (e: any) {
      setError(e?.message || String(e));
    }
  };

  if (!controls.length) return <div style={{ color: "var(--muted)" }}>{error || "Connect a projector first."}</div>;

  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">Projector Controls</h2>
      {error && <div className="mb-4 text-sm" style={{ color: "#ef4444" }}>{error}</div>}
      <div className="grid grid-cols-2 gap-4">
        {controls.map((c, i) => (
          <motion.div key={c.id}
            initial={{ opacity: 0, y: 15 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.05, type: "spring", stiffness: 200, damping: 20 }}
            className="rounded-lg p-4" style={{ background: "var(--surface)", border: "1px solid var(--border)" }}>
            <label className="block text-sm mb-2" style={{ color: "var(--muted)" }}>{c.label}</label>
            <ControlWidget ctrl={c} onChange={(v) => handleChange(c.id, v)} />
          </motion.div>
        ))}
      </div>
    </div>
  );
}
