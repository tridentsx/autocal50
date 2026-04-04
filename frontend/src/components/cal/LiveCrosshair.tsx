import { motion } from "framer-motion";

type Point = { x: number; y: number };

const W = 200, H = 180;
const XR: [number, number] = [0, 0.8];
const YR: [number, number] = [0, 0.7];

function toPixel(p: Point): { px: number; py: number } {
  return {
    px: ((p.x - XR[0]) / (XR[1] - XR[0])) * W,
    py: H - ((p.y - YR[0]) / (YR[1] - YR[0])) * H,
  };
}

/** Mini CIE 1931 diagram with animated measured dot converging on target. */
export default function LiveCrosshair({ target, measured, trail }: {
  target: Point;
  measured: Point;
  trail?: Point[];
}) {
  const t = toPixel(target);
  const m = toPixel(measured);

  return (
    <div className="relative rounded-lg overflow-hidden" style={{ width: W, height: H, background: "#111" }}>
      {/* Axis labels */}
      <span className="absolute bottom-0 right-1 text-[9px]" style={{ color: "var(--muted)" }}>x</span>
      <span className="absolute top-0 left-1 text-[9px]" style={{ color: "var(--muted)" }}>y</span>

      {/* Grid lines */}
      <svg width={W} height={H} className="absolute inset-0">
        {[0.2, 0.4, 0.6].map((v) => {
          const x = (v / 0.8) * W;
          const y = H - (v / 0.7) * H;
          return <g key={v}>
            <line x1={x} y1={0} x2={x} y2={H} stroke="#222" strokeWidth={0.5} />
            <line x1={0} y1={y} x2={W} y2={y} stroke="#222" strokeWidth={0.5} />
          </g>;
        })}

        {/* Trail dots */}
        {trail?.map((p, i) => {
          const pp = toPixel(p);
          return <circle key={i} cx={pp.px} cy={pp.py} r={2}
            fill="#6366f1" opacity={0.15 + (i / (trail.length || 1)) * 0.3} />;
        })}

        {/* Target crosshair */}
        <line x1={t.px - 8} y1={t.py} x2={t.px + 8} y2={t.py} stroke="#fff" strokeWidth={1} />
        <line x1={t.px} y1={t.py - 8} x2={t.px} y2={t.py + 8} stroke="#fff" strokeWidth={1} />
        <circle cx={t.px} cy={t.py} r={4} fill="none" stroke="#fff" strokeWidth={1} />
      </svg>

      {/* Animated measured dot */}
      <motion.div
        className="absolute rounded-full"
        style={{ width: 10, height: 10, background: "#6366f1", boxShadow: "0 0 8px #6366f1", marginLeft: -5, marginTop: -5 }}
        animate={{ left: m.px, top: m.py }}
        transition={{ type: "spring", stiffness: 80, damping: 12 }}
      />
    </div>
  );
}
