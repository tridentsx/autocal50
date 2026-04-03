import { useMemo, useState } from "react";
import ReactECharts from "echarts-for-react";

// --- ST.2084 PQ EOTF ---
function pqEOTF(v: number): number {
  if (v <= 0) return 0;
  const m1 = 0.1593017578125, m2 = 78.84375;
  const c1 = 0.8359375, c2 = 18.8515625, c3 = 18.6875;
  const vp = Math.pow(v, 1 / m2);
  const num = Math.max(vp - c1, 0);
  return Math.pow(num / (c2 - c3 * vp), 1 / m1); // 0-1 normalized (×10000 = cd/m²)
}

// --- Color space standards ---
type Standard = {
  label: string;
  r: [number, number]; g: [number, number]; b: [number, number];
  w: [number, number];
  gamma: number;
  hdr?: boolean;
  peakNits?: number;
};

const standards: Record<string, Standard> = {
  "rec709":    { label: "Rec. 709 / sRGB",    r: [0.64, 0.33], g: [0.30, 0.60], b: [0.15, 0.06], w: [0.3127, 0.329], gamma: 2.2 },
  "dcip3":     { label: "DCI-P3 (D65)",        r: [0.68, 0.32], g: [0.265, 0.69], b: [0.15, 0.06], w: [0.3127, 0.329], gamma: 2.6 },
  "p3d65":     { label: "Display P3",          r: [0.68, 0.32], g: [0.265, 0.69], b: [0.15, 0.06], w: [0.3127, 0.329], gamma: 2.2 },
  "bt2020":    { label: "BT.2020 (SDR)",       r: [0.708, 0.292], g: [0.17, 0.797], b: [0.131, 0.046], w: [0.3127, 0.329], gamma: 2.2 },
  "hdr10":     { label: "HDR10 (PQ)",          r: [0.708, 0.292], g: [0.17, 0.797], b: [0.131, 0.046], w: [0.3127, 0.329], gamma: 0, hdr: true, peakNits: 1000 },
  "hdr10_4k":  { label: "HDR10 (PQ 4000)",    r: [0.708, 0.292], g: [0.17, 0.797], b: [0.131, 0.046], w: [0.3127, 0.329], gamma: 0, hdr: true, peakNits: 4000 },
  "p3hdr":     { label: "P3-D65 HDR (PQ)",    r: [0.68, 0.32], g: [0.265, 0.69], b: [0.15, 0.06], w: [0.3127, 0.329], gamma: 0, hdr: true, peakNits: 1000 },
  "aces":      { label: "ACES AP1",            r: [0.713, 0.293], g: [0.165, 0.830], b: [0.128, 0.044], w: [0.32168, 0.33767], gamma: 2.2 },
};

// --- ColorChecker classic 24 patches (sRGB approximations + names) ---
const colorChecker = [
  { name: "Dark Skin",    rgb: "#735244" }, { name: "Light Skin",   rgb: "#c29682" },
  { name: "Blue Sky",     rgb: "#627a9d" }, { name: "Foliage",      rgb: "#576c43" },
  { name: "Blue Flower",  rgb: "#8580b1" }, { name: "Bluish Green", rgb: "#67bdaa" },
  { name: "Orange",       rgb: "#d67e2c" }, { name: "Purplish Blue",rgb: "#505ba6" },
  { name: "Moderate Red", rgb: "#c15a63" }, { name: "Purple",       rgb: "#5e3c6c" },
  { name: "Yellow Green", rgb: "#9dbc40" }, { name: "Orange Yellow",rgb: "#e0a32e" },
  { name: "Blue",         rgb: "#383d96" }, { name: "Green",        rgb: "#469449" },
  { name: "Red",          rgb: "#af363c" }, { name: "Yellow",       rgb: "#e7c71f" },
  { name: "Magenta",      rgb: "#bb5695" }, { name: "Cyan",         rgb: "#0885a1" },
  { name: "White",        rgb: "#f3f3f2" }, { name: "Neutral 8",   rgb: "#c8c8c8" },
  { name: "Neutral 6.5",  rgb: "#a0a0a0" }, { name: "Neutral 5",   rgb: "#7a7a79" },
  { name: "Neutral 3.5",  rgb: "#555555" }, { name: "Black",        rgb: "#343434" },
];

// --- Simulated measured data ---
function jitter(v: number, amt: number) { return v + (Math.random() - 0.5) * amt; }

const satColors = ["Red", "Green", "Blue", "Cyan", "Magenta", "Yellow"] as const;
const satLevels = [20, 40, 60, 80, 100];

function simForStandard(std: Standard) {
  const isHDR = !!std.hdr;
  return {
    primaries: [
      { label: "Red",   tx: std.r[0], ty: std.r[1], mx: jitter(std.r[0], 0.012), my: jitter(std.r[1], 0.010) },
      { label: "Green", tx: std.g[0], ty: std.g[1], mx: jitter(std.g[0], 0.010), my: jitter(std.g[1], 0.012) },
      { label: "Blue",  tx: std.b[0], ty: std.b[1], mx: jitter(std.b[0], 0.008), my: jitter(std.b[1], 0.006) },
    ],
    white: { tx: std.w[0], ty: std.w[1], mx: jitter(std.w[0], 0.004), my: jitter(std.w[1], 0.004) },
    grayscale: Array.from({ length: 11 }, (_, i) => ({
      label: `${i * 10}%`,
      x: jitter(std.w[0], 0.006), y: jitter(std.w[1], 0.006),
      Y: i * 10 * (isHDR ? (std.peakNits! / 1000) : 0.48),
    })),
    gamma: Array.from({ length: 21 }, (_, i) => {
      const input = i / 20;
      const target = isHDR ? pqEOTF(input) : Math.pow(input, std.gamma);
      return { input, target, measured: Math.max(0, target + (Math.random() - 0.5) * 0.025) };
    }),
    deltaE: [
      { label: "White", value: +(Math.random() * 1.5 + 0.3).toFixed(1) },
      { label: "Red", value: +(Math.random() * 2.5 + 0.5).toFixed(1) },
      { label: "Green", value: +(Math.random() * 2 + 0.4).toFixed(1) },
      { label: "Blue", value: +(Math.random() * 3 + 0.5).toFixed(1) },
      { label: "Cyan", value: +(Math.random() * 2 + 0.3).toFixed(1) },
      { label: "Magenta", value: +(Math.random() * 2.5 + 0.4).toFixed(1) },
      { label: "Yellow", value: +(Math.random() * 1.5 + 0.2).toFixed(1) },
    ],
    // Saturation sweeps: for each color at each level, target=level, measured=jittered
    saturation: satColors.map((color) => ({
      color,
      levels: satLevels.map((pct) => ({ pct, target: pct, measured: Math.round(jitter(pct, pct * 0.12)) })),
    })),
    // ColorChecker deltaE per patch
    colorChecker: colorChecker.map((p) => ({ ...p, deltaE: +(Math.random() * 3.5 + 0.2).toFixed(1) })),
    // Pre-cal checks
    preCal: {
      blackClip: { level: 1, visible: Math.random() > 0.3 },
      whiteClip: { level: 254, visible: Math.random() > 0.2 },
      peakLum: { full: jitter(isHDR ? (std.peakNits! * 0.6) : 42, 5), window10: jitter(isHDR ? std.peakNits! : 48, 8) },
      blackLevel: +(Math.random() * 0.08 + 0.01).toFixed(3),
      ablDrop: Math.round(jitter(isHDR ? 35 : 12, 8)),
    },
  };
}

// --- CIE 1931 spectral locus ---
const locus: [number, number][] = [
  [.1741,.005],[.174,.005],[.1738,.0049],[.1736,.0049],[.1733,.0048],
  [.173,.0048],[.1726,.0048],[.1721,.0048],[.1714,.0051],[.1703,.0058],
  [.1689,.0069],[.1669,.0086],[.1644,.0109],[.1611,.0138],[.1566,.0177],
  [.151,.0227],[.144,.0297],[.1355,.0399],[.1241,.0578],[.1096,.0868],
  [.0913,.1327],[.0687,.2007],[.0454,.295],[.0235,.4127],[.0082,.5384],
  [.0039,.6548],[.0139,.7502],[.0389,.812],[.0743,.8338],[.1142,.8262],
  [.1547,.8059],[.1929,.7816],[.2296,.7543],[.2658,.7243],[.3016,.6923],
  [.3373,.6589],[.3731,.6245],[.4087,.5896],[.4441,.5547],[.4788,.5202],
  [.5125,.4866],[.5448,.4544],[.5752,.4242],[.6029,.3965],[.627,.3725],
  [.6482,.3514],[.6658,.334],[.6801,.3197],[.6915,.3083],[.7006,.2993],
  [.7079,.292],[.714,.2859],[.719,.2809],[.723,.277],[.726,.274],
  [.7283,.2717],[.73,.27],[.7311,.2689],[.732,.268],[.7327,.2673],
  [.7334,.2666],[.734,.266],[.7344,.2656],[.7346,.2654],[.7347,.2653],
];

function pointInPoly(x: number, y: number, poly: [number, number][]) {
  let inside = false;
  for (let i = 0, j = poly.length - 1; i < poly.length; j = i++) {
    const [xi, yi] = poly[i], [xj, yj] = poly[j];
    if ((yi > y) !== (yj > y) && x < ((xj - xi) * (y - yi)) / (yj - yi) + xi) inside = !inside;
  }
  return inside;
}

function xyToRgb(cx: number, cy: number): [number, number, number] {
  const Y = 1, X = (Y / cy) * cx, Z = (Y / cy) * (1 - cx - cy);
  let r = 3.2406 * X - 1.5372 * Y - 0.4986 * Z;
  let g = -0.9689 * X + 1.8758 * Y + 0.0415 * Z;
  let b = 0.0557 * X - 0.204 * Y + 1.057 * Z;
  const m = Math.max(r, g, b, 1);
  r /= m; g /= m; b /= m;
  const srgb = (v: number) => Math.round(Math.max(0, Math.min(1, v <= 0.0031308 ? 12.92 * v : 1.055 * Math.pow(v, 1 / 2.4) - 0.055)) * 255);
  return [srgb(r), srgb(g), srgb(b)];
}

function generateCIEImage(w: number, h: number, xR: [number, number], yR: [number, number]) {
  const canvas = document.createElement("canvas");
  canvas.width = w; canvas.height = h;
  const ctx = canvas.getContext("2d")!;
  const img = ctx.createImageData(w, h);
  for (let py = 0; py < h; py++) {
    for (let px = 0; px < w; px++) {
      const cx = xR[0] + (px / w) * (xR[1] - xR[0]);
      const cy = yR[1] - (py / h) * (yR[1] - yR[0]);
      const idx = (py * w + px) * 4;
      if (cy > 0.01 && pointInPoly(cx, cy, locus)) {
        const [r, g, b] = xyToRgb(cx, cy);
        img.data[idx] = r; img.data[idx + 1] = g; img.data[idx + 2] = b; img.data[idx + 3] = 180;
      }
    }
  }
  ctx.putImageData(img, 0, 0);
  return canvas.toDataURL();
}

// --- Component ---
const card = { background: "var(--surface)", border: "1px solid var(--border)" };
const ax = { axisLine: { lineStyle: { color: "#2a2d3a" } }, axisLabel: { color: "#8b8fa3" }, splitLine: { lineStyle: { color: "#2a2d3a" } } };

export default function Measurements() {
  const [stdKey, setStdKey] = useState("rec709");
  const std = standards[stdKey];
  const sim = useMemo(() => simForStandard(std), [stdKey]);
  const cieImage = useMemo(() => generateCIEImage(320, 280, [0, 0.8], [0, 0.7]), []);
  const isHDR = !!std.hdr;

  const chromaticityOpt = {
    backgroundColor: "transparent",
    tooltip: { trigger: "item" as const },
    legend: { data: ["Target", "Measured", "White (target)", "White (measured)", "Grayscale"], textStyle: { color: "#8b8fa3", fontSize: 10 }, top: 0 },
    xAxis: { min: 0, max: 0.8, name: "x", ...ax },
    yAxis: { min: 0, max: 0.7, name: "y", ...ax },
    graphic: [{ type: "image", left: "center", top: "center", z: -1, style: { image: cieImage, width: 320, height: 280 } }],
    series: [
      { type: "line", name: "Target", data: [[...std.r], [...std.g], [...std.b], [...std.r]], lineStyle: { color: "#fff", width: 2 }, symbol: "none", silent: true },
      { type: "line", name: "Measured", data: sim.primaries.map((p) => [p.mx, p.my]).concat([[ sim.primaries[0].mx, sim.primaries[0].my ]]),
        lineStyle: { color: "#6366f1", width: 2 }, symbol: "none", silent: true },
      { type: "scatter", name: "Target", symbolSize: 10, symbol: "circle",
        data: sim.primaries.map((p) => ({ value: [p.tx, p.ty], name: `${p.label} target` })),
        itemStyle: { color: "transparent", borderColor: "#fff", borderWidth: 2 } },
      { type: "scatter", name: "Measured", symbolSize: 10,
        data: sim.primaries.map((p) => ({ value: [p.mx, p.my], name: `${p.label} measured` })),
        itemStyle: { color: "#6366f1", borderColor: "#fff", borderWidth: 1 } },
      { type: "scatter", name: "White (target)", symbolSize: 12, symbol: "diamond",
        data: [{ value: [std.w[0], std.w[1]], name: "D65 target" }],
        itemStyle: { color: "transparent", borderColor: "#facc15", borderWidth: 2 } },
      { type: "scatter", name: "White (measured)", symbolSize: 12, symbol: "diamond",
        data: [{ value: [sim.white.mx, sim.white.my], name: "D65 measured" }],
        itemStyle: { color: "#facc15", borderColor: "#fff", borderWidth: 1 } },
      { type: "scatter", name: "Grayscale", symbolSize: 5, data: sim.grayscale.map((g) => [g.x, g.y]), itemStyle: { color: "#fff" } },
    ],
  };

  const eotfLabel = isHDR ? "PQ ST.2084" : `γ${std.gamma}`;
  const gammaOpt = {
    backgroundColor: "transparent",
    tooltip: { trigger: "axis" as const },
    legend: { data: [`Target ${eotfLabel}`, "Measured"], textStyle: { color: "#8b8fa3" }, top: 0 },
    xAxis: { type: "category" as const, data: sim.gamma.map((g) => `${(g.input * 100).toFixed(0)}%`), ...ax },
    yAxis: { min: 0, max: isHDR ? undefined : 1, name: isHDR ? "Normalized luminance" : undefined, ...ax },
    series: [
      { name: `Target ${eotfLabel}`, type: "line", data: sim.gamma.map((g) => +g.target.toFixed(4)), lineStyle: { color: "#555", type: "dashed" as const }, symbol: "none" },
      { name: "Measured", type: "line", data: sim.gamma.map((g) => +g.measured.toFixed(4)), lineStyle: { color: "#6366f1" }, symbol: "circle", symbolSize: 4, itemStyle: { color: "#6366f1" } },
    ],
  };

  const deltaEOpt = {
    backgroundColor: "transparent",
    tooltip: { trigger: "axis" as const },
    xAxis: { type: "category" as const, data: sim.deltaE.map((d) => d.label), ...ax },
    yAxis: { min: 0, ...ax },
    series: [
      { type: "bar", data: sim.deltaE.map((d) => ({ value: d.value, itemStyle: { color: d.value < 1 ? "#22c55e" : d.value < 2 ? "#facc15" : "#ef4444" } })) },
      { type: "line", markLine: { silent: true, data: [{ yAxis: 2, lineStyle: { color: "#ef4444", type: "dashed" as const } }] }, data: [] },
    ],
  };

  const grayscaleOpt = {
    backgroundColor: "transparent",
    tooltip: { trigger: "axis" as const },
    legend: { data: ["Measured Y", "Ideal"], textStyle: { color: "#8b8fa3" }, top: 0 },
    xAxis: { type: "category" as const, data: sim.grayscale.map((g) => g.label), ...ax },
    yAxis: { name: "Luminance (cd/m²)", ...ax },
    series: [
      { name: "Measured Y", type: "line", data: sim.grayscale.map((g) => g.Y.toFixed(1)), lineStyle: { color: "#6366f1" }, symbol: "circle", symbolSize: 6, itemStyle: { color: "#6366f1" } },
      { name: "Ideal", type: "line", data: sim.grayscale.map((_, i) => {
        const peak = isHDR ? (std.peakNits! / 1000) * 48.2 : 48.2;
        return (( isHDR ? pqEOTF(i / 10) : Math.pow(i / 10, std.gamma)) * peak).toFixed(1);
      }), lineStyle: { color: "#555", type: "dashed" as const }, symbol: "none" },
    ],
  };

  // Saturation sweep radar
  const satColors6 = ["#ef4444", "#22c55e", "#3b82f6", "#06b6d4", "#d946ef", "#eab308"];
  const satOpt = {
    backgroundColor: "transparent",
    tooltip: {},
    legend: { data: satLevels.map((l) => `${l}% target`).concat(satLevels.map((l) => `${l}% measured`)), textStyle: { color: "#8b8fa3", fontSize: 9 }, top: 0, itemWidth: 12 },
    radar: { indicator: satColors.map((c) => ({ name: c, max: 110 })), shape: "circle" as const,
      axisName: { color: "#8b8fa3" }, splitArea: { areaStyle: { color: "transparent" } },
      splitLine: { lineStyle: { color: "#2a2d3a" } }, axisLine: { lineStyle: { color: "#2a2d3a" } } },
    series: [{ type: "radar",
      data: satLevels.flatMap((lvl, li) => {
        const tgt = sim.saturation.map((s) => s.levels[li].target);
        const meas = sim.saturation.map((s) => s.levels[li].measured);
        const alpha = 0.4 + li * 0.15;
        return [
          { name: `${lvl}% target`, value: tgt, lineStyle: { color: `rgba(255,255,255,${alpha})`, type: "dashed" as const }, symbol: "none" },
          { name: `${lvl}% measured`, value: meas, lineStyle: { color: `rgba(99,102,241,${alpha})` }, symbol: "circle", symbolSize: 4, itemStyle: { color: `rgba(99,102,241,${alpha})` } },
        ];
      }),
    }],
  };

  // ColorChecker grid
  const ccOpt = {
    backgroundColor: "transparent",
    tooltip: { formatter: (p: any) => `${p.name}: ΔE ${p.value[2]}` },
    grid: { left: 10, right: 10, top: 10, bottom: 10 },
    xAxis: { show: false, min: 0, max: 6 },
    yAxis: { show: false, min: 0, max: 4, inverse: true },
    series: [{
      type: "scatter", symbolSize: 48, symbol: "roundRect",
      data: sim.colorChecker.map((p, i) => ({
        name: p.name,
        value: [i % 6 + 0.5, Math.floor(i / 6) + 0.5, p.deltaE],
        itemStyle: { color: p.rgb, borderColor: p.deltaE < 1 ? "#22c55e" : p.deltaE < 2 ? "#facc15" : p.deltaE < 3 ? "#ef4444" : "#ff0000", borderWidth: 2 },
        label: { show: true, formatter: `${p.deltaE}`, fontSize: 9, color: "#fff", textBorderColor: "#000", textBorderWidth: 2 },
      })),
    }],
  };

  // Pre-cal checks
  const pc = sim.preCal;
  const cr = pc.peakLum.window10 / Math.max(pc.blackLevel, 0.001);

  const avgDe = (sim.deltaE.reduce((s, d) => s + d.value, 0) / sim.deltaE.length).toFixed(1);
  const maxDe = Math.max(...sim.deltaE.map((d) => d.value)).toFixed(1);
  const ccAvg = (sim.colorChecker.reduce((s, p) => s + p.deltaE, 0) / sim.colorChecker.length).toFixed(1);

  return (
    <div>
      <div className="flex items-center gap-4 mb-6">
        <h2 className="text-2xl font-bold">Measurements</h2>
        <select value={stdKey} onChange={(e) => setStdKey(e.target.value)}
          className="p-2 rounded text-sm" style={{ background: "var(--bg)", color: "var(--text)", border: "1px solid var(--border)" }}>
          {Object.entries(standards).map(([k, s]) => <option key={k} value={k}>{s.label}</option>)}
        </select>
        <span className="text-sm" style={{ color: "var(--muted)" }}>
          Avg ΔE: <span style={{ color: +avgDe < 2 ? "#22c55e" : "#facc15" }}>{avgDe}</span>
          {" · "}Max ΔE: <span style={{ color: +maxDe < 2 ? "#22c55e" : +maxDe < 3 ? "#facc15" : "#ef4444" }}>{maxDe}</span>
          {" · "}CC Avg: <span style={{ color: +ccAvg < 2 ? "#22c55e" : +ccAvg < 3 ? "#facc15" : "#ef4444" }}>{ccAvg}</span>
        </span>
      </div>

      {/* Pre-Cal Checks */}
      <div className="rounded-lg p-4 mb-4" style={card}>
        <h3 className="text-sm font-semibold mb-3" style={{ color: "var(--muted)" }}>Pre-Calibration Checks</h3>
        <div className="grid grid-cols-5 gap-4">
          <Check label="Black Clipping" pass={pc.blackClip.visible} detail={`Level ${pc.blackClip.level} ${pc.blackClip.visible ? "visible" : "crushed"}`} />
          <Check label="White Clipping" pass={pc.whiteClip.visible} detail={`Level ${pc.whiteClip.level} ${pc.whiteClip.visible ? "visible" : "clipped"}`} />
          <Check label="Black Level" pass={pc.blackLevel < 0.05} detail={`${pc.blackLevel} cd/m²`} />
          <Check label="Peak Luminance" pass={true} detail={`${pc.peakLum.window10.toFixed(0)} cd/m² (10% win)`} />
          <Check label="ABL Drop" pass={pc.ablDrop < 20} detail={`${pc.ablDrop}% (full vs 10% window)`} warn={pc.ablDrop >= 15} />
        </div>
        <div className="mt-2 text-xs" style={{ color: "var(--muted)" }}>
          Contrast ratio: {cr > 9999 ? "∞" : cr.toFixed(0)}:1 · Dynamic range: {(Math.log10(cr) * 20).toFixed(0)} dB
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="rounded-lg p-4" style={card}>
          <h3 className="text-sm font-semibold mb-2" style={{ color: "var(--muted)" }}>CIE 1931 — {std.label} vs Measured</h3>
          <ReactECharts option={chromaticityOpt} style={{ height: 360 }} />
        </div>
        <div className="rounded-lg p-4" style={card}>
          <h3 className="text-sm font-semibold mb-2" style={{ color: "var(--muted)" }}>EOTF — {isHDR ? "PQ ST.2084" : `Gamma ${std.gamma}`}</h3>
          <ReactECharts option={gammaOpt} style={{ height: 360 }} />
        </div>
        <div className="rounded-lg p-4" style={card}>
          <h3 className="text-sm font-semibold mb-2" style={{ color: "var(--muted)" }}>Delta E (CIE2000) vs {std.label}</h3>
          <ReactECharts option={deltaEOpt} style={{ height: 360 }} />
        </div>
        <div className="rounded-lg p-4" style={card}>
          <h3 className="text-sm font-semibold mb-2" style={{ color: "var(--muted)" }}>Grayscale Luminance</h3>
          <ReactECharts option={grayscaleOpt} style={{ height: 360 }} />
        </div>
        <div className="rounded-lg p-4" style={card}>
          <h3 className="text-sm font-semibold mb-2" style={{ color: "var(--muted)" }}>Saturation Sweeps — Target vs Measured</h3>
          <ReactECharts option={satOpt} style={{ height: 360 }} />
        </div>
        <div className="rounded-lg p-4" style={card}>
          <h3 className="text-sm font-semibold mb-2" style={{ color: "var(--muted)" }}>ColorChecker — ΔE per Patch (avg {ccAvg})</h3>
          <ReactECharts option={ccOpt} style={{ height: 360 }} />
        </div>
      </div>
    </div>
  );
}

function Check({ label, pass, detail, warn }: { label: string; pass: boolean; detail: string; warn?: boolean }) {
  const color = pass ? (warn ? "#facc15" : "#22c55e") : "#ef4444";
  const icon = pass ? (warn ? "⚠" : "✓") : "✗";
  return (
    <div className="text-center">
      <div className="text-lg" style={{ color }}>{icon}</div>
      <div className="text-xs font-medium" style={{ color: "var(--text)" }}>{label}</div>
      <div className="text-xs" style={{ color: "var(--muted)" }}>{detail}</div>
    </div>
  );
}
