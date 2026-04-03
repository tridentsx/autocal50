import { useEffect, useState } from "react";
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
      <button onClick={() => onChange(!ctrl.value)} disabled={ctrl.readOnly}
        className="px-4 py-1 rounded text-sm"
        style={{ background: ctrl.value ? "#22c55e" : "var(--bg)", color: "var(--text)", border: "1px solid var(--border)" }}>
        {ctrl.value ? "ON" : "OFF"}
      </button>
    );
  return null;
}

export default function Controls() {
  const [controls, setControls] = useState<(Control & { value: any })[]>([]);

  useEffect(() => {
    api.getCapabilities().then(async (caps: any) => {
      const loaded = await Promise.all(
        (caps.controls as Control[]).map(async (c) => {
          const value = await api.getControl(c.id).catch(() => c.type === "toggle" ? false : c.min);
          return { ...c, value };
        })
      );
      setControls(loaded);
    }).catch(() => {});
  }, []);

  const handleChange = async (id: string, value: any) => {
    await api.setControl(id, value);
    setControls((prev) => prev.map((c) => (c.id === id ? { ...c, value } : c)));
  };

  if (!controls.length) return <div style={{ color: "var(--muted)" }}>Connect a projector first.</div>;

  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">Projector Controls</h2>
      <div className="grid grid-cols-2 gap-4">
        {controls.map((c) => (
          <div key={c.id} className="rounded-lg p-4" style={{ background: "var(--surface)", border: "1px solid var(--border)" }}>
            <label className="block text-sm mb-2" style={{ color: "var(--muted)" }}>{c.label}</label>
            <ControlWidget ctrl={c} onChange={(v) => handleChange(c.id, v)} />
          </div>
        ))}
      </div>
    </div>
  );
}
