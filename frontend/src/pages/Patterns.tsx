import { useEffect, useRef, useState } from "react";
import { api } from "@/api";
import type { Pattern, DisplayOutput, RGB } from "@/types";

function rgb(c: RGB) { return `rgb(${c.r},${c.g},${c.b})`; }

function Thumbnail({ pattern: p, active, onClick }: { pattern: Pattern; active: boolean; onClick: () => void }) {
  const ref = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const c = ref.current;
    if (!c) return;
    const ctx = c.getContext("2d")!;
    const w = c.width, h = c.height;
    ctx.fillStyle = "#000";
    ctx.fillRect(0, 0, w, h);

    switch (p.type) {
      case "solid":
        ctx.fillStyle = rgb(p.color!);
        ctx.fillRect(0, 0, w, h);
        break;
      case "window": {
        const pct = (p.windowPc ?? 18) / 100;
        const side = Math.sqrt(pct);
        ctx.fillStyle = rgb(p.color!);
        ctx.fillRect((w - w * side) / 2, (h - h * side) / 2, w * side, h * side);
        break;
      }
      case "gradient": {
        const steps = p.steps ?? [];
        const sw = w / steps.length;
        for (let i = 0; i < steps.length; i++) {
          ctx.fillStyle = rgb(steps[i]);
          ctx.fillRect(i * sw, 0, sw + 1, h);
        }
        break;
      }
      case "grid":
        ctx.strokeStyle = rgb(p.color!);
        ctx.lineWidth = 0.5;
        for (let i = 0; i <= (p.cols ?? 16); i++) { const x = i * w / (p.cols ?? 16); ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, h); ctx.stroke(); }
        for (let i = 0; i <= (p.rows ?? 9); i++) { const y = i * h / (p.rows ?? 9); ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(w, y); ctx.stroke(); }
        break;
      case "checker": {
        const cols = p.cols ?? 6, rows = p.rows ?? 4;
        const pal = p.steps ?? [];
        for (let r = 0; r < rows; r++)
          for (let c = 0; c < cols; c++) {
            ctx.fillStyle = r * cols + c < pal.length ? rgb(pal[r * cols + c]) : "#000";
            ctx.fillRect(c * w / cols, r * h / rows, w / cols + 1, h / rows + 1);
          }
        break;
      }
    }
  }, [p]);

  return (
    <button onClick={onClick} className="flex flex-col items-center gap-1 p-2 rounded"
      style={{ background: active ? "var(--accent)" : "var(--surface)", border: "1px solid var(--border)" }}>
      <canvas ref={ref} width={80} height={45} className="rounded" />
      <span className="text-xs truncate w-20 text-center" style={{ color: active ? "#fff" : "var(--muted)" }}>{p.name}</span>
    </button>
  );
}

export default function Patterns() {
  const [battery, setBattery] = useState<Pattern[]>([]);
  const [displays, setDisplays] = useState<DisplayOutput[]>([]);
  const [selectedDisplay, setSelectedDisplay] = useState("");
  const [activeId, setActiveId] = useState("");
  const [patternWindow, setPatternWindow] = useState<Window | null>(null);
  const [edid, setEdid] = useState<any>(null);

  useEffect(() => {
    api.getPatternBattery().then(setBattery).catch(() => {});
    api.listDisplays().then((d) => {
      setDisplays(d);
      api.getPatternDisplay().then((name) => {
        if (name) { setSelectedDisplay(name); api.getDisplayEDID(name).then(setEdid).catch(() => setEdid(null)); }
        else if (d.length) { setSelectedDisplay(d[0].name); api.setPatternDisplay(d[0].name); api.getDisplayEDID(d[0].name).then(setEdid).catch(() => setEdid(null)); }
      });
    }).catch(() => {});
  }, []);

  const selectPattern = (p: Pattern) => {
    setActiveId(p.id);
    api.setActivePattern(p);
  };

  const openPatternWindow = () => {
    const w = patternWindow && !patternWindow.closed ? patternWindow : window.open("/pattern", "pattern", "popup");
    if (w) {
      setPatternWindow(w);
      w.focus();
      setTimeout(() => w.document.documentElement.requestFullscreen?.(), 300);
    }
  };

  const handleDisplayChange = (name: string) => {
    setSelectedDisplay(name);
    api.setPatternDisplay(name);
    api.getDisplayEDID(name).then(setEdid).catch(() => setEdid(null));
  };

  const groups = [...new Set(battery.map((p) => p.group))];

  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">Test Patterns</h2>

      {/* Controls bar */}
      <div className="flex gap-4 mb-6 items-end">
        <div>
          <label className="block text-sm mb-1" style={{ color: "var(--muted)" }}>Display Output</label>
          <select value={selectedDisplay} onChange={(e) => handleDisplayChange(e.target.value)}
            className="p-2 rounded" style={{ background: "var(--bg)", color: "var(--text)", border: "1px solid var(--border)" }}>
            {displays.filter((d) => d.connected).map((d) => (
              <option key={d.name} value={d.name}>
                {d.name} ({d.width}x{d.height} @ {d.refreshHz}Hz){d.primary ? " ★" : ""}
              </option>
            ))}
          </select>
        </div>
        <button onClick={openPatternWindow}
          className="px-4 py-2 rounded font-medium" style={{ background: "var(--accent)", color: "#fff" }}>
          Open Pattern Window
        </button>
      </div>

      {/* EDID info */}
      {edid && (
        <div className="rounded-lg p-4 mb-6" style={{ background: "var(--surface)", border: "1px solid var(--border)" }}>
          <h3 className="text-sm font-semibold mb-2" style={{ color: "var(--muted)" }}>Display EDID — {edid.displayName || edid.manufacturer}</h3>
          <div className="grid grid-cols-4 gap-x-6 gap-y-1 text-sm">
            <Kv label="Manufacturer" value={edid.manufacturer} />
            <Kv label="Model" value={edid.displayName || `0x${edid.productCode.toString(16)}`} />
            <Kv label="Serial" value={edid.serialString || String(edid.serial)} />
            <Kv label="Year / Week" value={`${edid.year} / W${edid.week}`} />
            <Kv label="EDID Version" value={edid.version} />
            <Kv label="Bit Depth" value={edid.bitDepth ? `${edid.bitDepth}-bit` : "N/A"} />
            <Kv label="Max Resolution" value={edid.maxHRes && edid.maxVRes ? `${edid.maxHRes}×${edid.maxVRes}` : "N/A"} />
            <Kv label="Input" value={edid.digitalInput ? "Digital" : "Analog"} />
            <Kv label="HDR" value={edid.hdr} flag />
            <Kv label="BT.2020" value={edid.bt2020} flag />
            <Kv label="DCI-P3" value={edid.p3} flag />
          </div>
        </div>
      )}

      {/* Pattern groups */}
      {groups.map((group) => (
        <div key={group} className="mb-6">
          <h3 className="text-sm font-semibold mb-2" style={{ color: "var(--muted)" }}>{group}</h3>
          <div className="flex flex-wrap gap-2">
            {battery.filter((p) => p.group === group).map((p) => (
              <Thumbnail key={p.id} pattern={p} active={p.id === activeId} onClick={() => selectPattern(p)} />
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}

function Kv({ label, value, flag }: { label: string; value: any; flag?: boolean }) {
  return (
    <div className="flex justify-between">
      <span style={{ color: "var(--muted)" }}>{label}</span>
      {flag
        ? <span style={{ color: value ? "#22c55e" : "#ef4444" }}>{value ? "✓" : "✗"}</span>
        : <span>{value}</span>}
    </div>
  );
}
