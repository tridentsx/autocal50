import { useEffect, useState } from "react";
import { api } from "@/api";

type DeviceSection = {
  label: string;
  listDrivers: () => Promise<string[]>;
  connect: (d: string, cfg: Record<string, any>) => Promise<void>;
  disconnect: () => Promise<void>;
  needsSerial?: boolean;
};

const sections: DeviceSection[] = [
  { label: "Projector", listDrivers: api.listProjectorDrivers, connect: (d, cfg) => api.connectProjector(d, cfg), disconnect: api.disconnectProjector, needsSerial: true },
  { label: "Meter", listDrivers: api.listMeterDrivers, connect: (d, cfg) => api.connectMeter(d, cfg), disconnect: api.disconnectMeter },
  { label: "Transport", listDrivers: api.listTransportDrivers, connect: (d, cfg) => api.connectTransport(d, cfg), disconnect: api.disconnectTransport },
];

const serialDrivers = new Set(["xgimi-rs232"]);

function DeviceCard({ section }: { section: DeviceSection }) {
  const [drivers, setDrivers] = useState<string[]>([]);
  const [selected, setSelected] = useState("");
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState("");
  const [ports, setPorts] = useState<string[]>([]);
  const [selectedPort, setSelectedPort] = useState("");

  useEffect(() => {
    section.listDrivers().then((d) => { setDrivers(d); if (d.length) setSelected(d[0]); }).catch(() => {});
  }, []);

  useEffect(() => {
    if (section.needsSerial && serialDrivers.has(selected)) {
      api.listSerialPorts().then((p) => { setPorts(p); if (p.length) setSelectedPort(p[0]); }).catch(() => setPorts([]));
    }
  }, [selected]);

  const showSerial = section.needsSerial && serialDrivers.has(selected);

  const toggle = async () => {
    setError("");
    try {
      if (connected) { await section.disconnect(); setConnected(false); }
      else {
        const cfg: Record<string, any> = {};
        if (showSerial && selectedPort) cfg.device = selectedPort;
        await section.connect(selected, cfg);
        setConnected(true);
      }
    } catch (e: any) {
      setError(e?.message || String(e));
    }
  };

  return (
    <div className="rounded-lg p-4" style={{ background: "var(--surface)", border: "1px solid var(--border)" }}>
      <h3 className="text-lg font-semibold mb-3">{section.label}</h3>
      <select value={selected} onChange={(e) => setSelected(e.target.value)}
        className="w-full p-2 rounded mb-3" style={{ background: "var(--bg)", color: "var(--text)", border: "1px solid var(--border)" }}>
        {drivers.map((d) => <option key={d} value={d}>{d}</option>)}
      </select>

      {showSerial && (
        <select value={selectedPort} onChange={(e) => setSelectedPort(e.target.value)}
          className="w-full p-2 rounded mb-3" style={{ background: "var(--bg)", color: "var(--text)", border: "1px solid var(--border)" }}>
          {ports.length === 0 && <option value="">No serial ports found</option>}
          {ports.map((p) => <option key={p} value={p}>{p}</option>)}
        </select>
      )}

      <button onClick={toggle}
        className="px-4 py-2 rounded font-medium w-full"
        style={{ background: connected ? "#ef4444" : "var(--accent)", color: "#fff" }}>
        {connected ? "Disconnect" : "Connect"}
      </button>
      <div className="mt-2 text-sm" style={{ color: connected ? "#22c55e" : "var(--muted)" }}>
        {connected ? "● Connected" : "○ Disconnected"}
      </div>
      {error && <div className="mt-2 text-sm" style={{ color: "#ef4444" }}>{error}</div>}
    </div>
  );
}

export default function Devices() {
  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">Devices</h2>
      <div className="grid grid-cols-3 gap-4">
        {sections.map((s) => <DeviceCard key={s.label} section={s} />)}
      </div>
    </div>
  );
}
