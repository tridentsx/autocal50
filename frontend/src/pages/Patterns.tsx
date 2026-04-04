import { useEffect, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import { motion } from "framer-motion";
import { api } from "@/api";
import type { Pattern, DisplayOutput, RGB } from "@/types";

// Map sidebar category slugs to pattern group name prefixes
const categoryGroups: Record<string, string[]> = {
  grayscale: ["Grayscale", "Grayscale Windows"],
  clipping: ["Clipping"],
  colors: ["Primaries", "Secondaries", "Color Windows"],
  saturation: ["Saturation"],
  peak: ["Peak vs Size"],
  hdr: ["HDR EOTF Steps"],
  contrast: ["Contrast Ratio", "Windows"],
  gradients: ["Gradients"],
  geometry: ["Geometry"],
  colorchecker: ["ColorChecker"],
};

// Purpose descriptions for pattern groups (shown in group headers)
const groupPurpose: Record<string, string> = {
  "Grayscale": "Full-field IRE steps for white balance and gamma measurement",
  "Grayscale Windows": "18% windows on black for accurate meter readings without ABL influence",
  "Clipping – Black": "Near-black steps (0–5%) to check shadow detail visibility",
  "Clipping – White": "Near-white steps (95–100%) to check highlight clipping",
  "Primaries": "Full-field R/G/B for gamut boundary measurement",
  "Secondaries": "Full-field C/M/Y for secondary color accuracy",
  "Color Windows": "18% color windows for accurate primary/secondary metering",
  "Saturation 20%": "20% saturation sweep across all colors for color tracking",
  "Saturation 40%": "40% saturation sweep across all colors for color tracking",
  "Saturation 60%": "60% saturation sweep across all colors for color tracking",
  "Saturation 80%": "80% saturation sweep across all colors for color tracking",
  "Saturation 100%": "100% saturation sweep across all colors for color tracking",
  "Peak vs Size": "White windows at different sizes to detect auto brightness limiting (ABL)",
  "HDR EOTF Steps": "PQ ST.2084 code values at specific nit levels for HDR EOTF verification",
  "Contrast Ratio": "Black/white fields and checkerboards for sequential and ANSI contrast",
  "Windows": "Standard measurement windows at common sizes (10%, 18%, 50%)",
  "Gradients": "Smooth ramps to check banding, quantization, and monotonicity",
  "Geometry": "Crosshatch grid for convergence, alignment, and keystone verification",
  "ColorChecker": "Classic 24-patch reference for overall color accuracy (ΔE verification)",
};

// Purpose for individual pattern types
function patternPurpose(p: Pattern): string {
  if (p.group.startsWith("Saturation")) return `${p.name} — measures color tracking accuracy at partial saturation`;
  switch (p.type) {
    case "solid": return `${p.name} — full-field stimulus for direct measurement`;
    case "window": return `${p.name} — ${p.windowPc}% window on black, avoids ABL for accurate metering`;
    case "gradient": return `${p.name} — smooth ramp to check banding and tonal transitions`;
    case "grid": return `${p.name} — line grid for geometric alignment verification`;
    case "checker": return p.group === "ColorChecker"
      ? `${p.name} — 24-patch reference target for ΔE verification`
      : `${p.name} — alternating pattern for contrast measurement`;
    default: return p.name;
  }
}

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
    <button onClick={onClick} title={patternPurpose(p)} className="flex flex-col items-center gap-1 p-2 rounded"
      style={{ background: active ? "var(--accent)" : "var(--surface)", border: "1px solid var(--border)" }}>
      <canvas ref={ref} width={80} height={45} className="rounded" />
      <span className="text-xs truncate w-20 text-center" style={{ color: active ? "#fff" : "var(--muted)" }}>{p.name}</span>
    </button>
  );
}

export default function Patterns() {
  const { category } = useParams<{ category?: string }>();
  const [battery, setBattery] = useState<Pattern[]>([]);
  const [displays, setDisplays] = useState<DisplayOutput[]>([]);
  const [selectedDisplay, setSelectedDisplay] = useState("");
  const [activeId, setActiveId] = useState("");
  const [patternWindow, setPatternWindow] = useState<Window | null>(null);
  const [edid, setEdid] = useState<any>(null);

  const [error, setError] = useState("");

  useEffect(() => {
    api.getPatternBattery().then(setBattery).catch((e) => setError(String(e)));
    api.listDisplays().then((d) => {
      setDisplays(d);
      api.getPatternDisplay().then((name) => {
        if (name) { setSelectedDisplay(name); api.getDisplayEDID(name).then(setEdid).catch(() => setEdid(null)); }
        else if (d.length) { setSelectedDisplay(d[0].name); api.setPatternDisplay(d[0].name); api.getDisplayEDID(d[0].name).then(setEdid).catch(() => setEdid(null)); }
      });
    }).catch((e) => setError(String(e)));
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
    }
  };

  const handleDisplayChange = (name: string) => {
    setSelectedDisplay(name);
    api.setPatternDisplay(name);
    api.getDisplayEDID(name).then(setEdid).catch(() => setEdid(null));
  };

  const filtered = category && categoryGroups[category]
    ? battery.filter((p) => categoryGroups[category].some((g) => p.group.startsWith(g)))
    : battery;
  const groups = [...new Set(filtered.map((p) => p.group))];

  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">Test Patterns</h2>
      {error && <div className="mb-4 text-sm" style={{ color: "#ef4444" }}>{error}</div>}

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
          <h3 className="text-sm font-semibold mb-2" title={groupPurpose[group] ?? ""} style={{ color: "var(--muted)" }}>{group}</h3>
          <div className="flex flex-wrap gap-2">
            {filtered.filter((p) => p.group === group).map((p, i) => (
              <motion.div key={p.id}
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: i * 0.02, type: "spring", stiffness: 300, damping: 20 }}>
                <Thumbnail pattern={p} active={p.id === activeId} onClick={() => selectPattern(p)} />
              </motion.div>
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
