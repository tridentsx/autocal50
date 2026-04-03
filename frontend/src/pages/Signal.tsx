import { useEffect, useState } from "react";
import { api } from "@/api";
import type { SignalFormat } from "@/types";

const encodings = ["RGB", "YCbCr444", "YCbCr422", "YCbCr420"];
const bitDepths = [8, 10, 12];

export default function Signal() {
  const [formats, setFormats] = useState<SignalFormat[]>([]);
  const [current, setCurrent] = useState<SignalFormat | null>(null);

  useEffect(() => {
    api.getSignalFormats().then(setFormats).catch(() => {});
    api.currentSignalFormat().then(setCurrent).catch(() => {});
  }, []);

  const apply = async (patch: Partial<SignalFormat>) => {
    const next = { ...current!, ...patch };
    await api.applySignalFormat(next);
    setCurrent(next);
  };

  if (!current) return <div style={{ color: "var(--muted)" }}>Connect a transport device first.</div>;

  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">Signal</h2>
      <div className="grid grid-cols-2 gap-4">
        <Field label="Resolution">
          <select value={current.resolution} onChange={(e) => apply({ resolution: e.target.value })}
            className="w-full p-2 rounded" style={{ background: "var(--bg)", color: "var(--text)", border: "1px solid var(--border)" }}>
            {[...new Set(formats.map((f) => f.resolution))].map((r) => <option key={r}>{r}</option>)}
          </select>
        </Field>
        <Field label="Refresh Rate">
          <select value={current.refreshHz} onChange={(e) => apply({ refreshHz: +e.target.value })}
            className="w-full p-2 rounded" style={{ background: "var(--bg)", color: "var(--text)", border: "1px solid var(--border)" }}>
            {[...new Set(formats.map((f) => f.refreshHz))].map((r) => <option key={r} value={r}>{r} Hz</option>)}
          </select>
        </Field>
        <Field label="Encoding">
          <div className="flex gap-2">
            {encodings.map((e) => (
              <button key={e} onClick={() => apply({ encoding: e })}
                className="px-3 py-1 rounded text-sm"
                style={{ background: current.encoding === e ? "var(--accent)" : "var(--bg)", color: "var(--text)", border: "1px solid var(--border)" }}>
                {e}
              </button>
            ))}
          </div>
        </Field>
        <Field label="Bit Depth">
          <div className="flex gap-2">
            {bitDepths.map((b) => (
              <button key={b} onClick={() => apply({ bitDepth: b })}
                className="px-3 py-1 rounded text-sm"
                style={{ background: current.bitDepth === b ? "var(--accent)" : "var(--bg)", color: "var(--text)", border: "1px solid var(--border)" }}>
                {b}-bit
              </button>
            ))}
          </div>
        </Field>
        <Field label="HDR">
          <button onClick={() => apply({ hdr: !current.hdr })}
            className="px-4 py-1 rounded text-sm"
            style={{ background: current.hdr ? "#22c55e" : "var(--bg)", color: "var(--text)", border: "1px solid var(--border)" }}>
            {current.hdr ? "ON" : "OFF"}
          </button>
        </Field>
      </div>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg p-4" style={{ background: "var(--surface)", border: "1px solid var(--border)" }}>
      <label className="block text-sm mb-2" style={{ color: "var(--muted)" }}>{label}</label>
      {children}
    </div>
  );
}
