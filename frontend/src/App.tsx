import { useState } from "react";
import { NavLink, Route, Routes, useLocation } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import Devices from "@/pages/Devices";
import Signal from "@/pages/Signal";
import Controls from "@/pages/Controls";
import Patterns from "@/pages/Patterns";
import PatternRenderer from "@/pages/PatternRenderer";
import LiveView from "@/pages/LiveView";
import Measurements from "@/pages/Measurements";

import Calibration from "@/pages/Calibration";

type NavItem = { to: string; label: string; tip?: string };
type NavGroup = { group: string; items: NavItem[] };

const nav: NavGroup[] = [
  { group: "Setup", items: [
    { to: "/", label: "Devices", tip: "Connect projector, colorimeter, and signal transport" },
    { to: "/signal", label: "Signal", tip: "Resolution, refresh rate, encoding, bit depth, HDR mode" },
    { to: "/controls", label: "Controls", tip: "Projector settings — brightness, image mode, input, etc." },
  ]},
  { group: "Calibration", items: [
    { to: "/calibration/assisted", label: "Assisted", tip: "Step-by-step guided calibration — you adjust controls while the app measures" },
    { to: "/calibration/automatic", label: "Automatic", tip: "Fully automated — the app controls the projector and iterates to targets" },
  ]},
  { group: "Patterns", items: [
    { to: "/patterns", label: "All", tip: "Browse the full test pattern battery" },
    { to: "/patterns/grayscale", label: "Grayscale", tip: "IRE steps (0–100%) for white balance and gamma measurement" },
    { to: "/patterns/clipping", label: "Clipping", tip: "Near-black and near-white steps to check shadow/highlight visibility" },
    { to: "/patterns/colors", label: "Colors", tip: "Primary and secondary colors for gamut measurement" },
    { to: "/patterns/saturation", label: "Saturation", tip: "Color sweeps at 20–100% saturation for color tracking accuracy" },
    { to: "/patterns/peak", label: "Peak / ABL", tip: "White windows at different sizes to detect auto brightness limiting" },
    { to: "/patterns/hdr", label: "HDR EOTF", tip: "PQ ST.2084 code values at specific nit levels" },
    { to: "/patterns/contrast", label: "Contrast", tip: "Black/white fields and checkerboards for contrast ratio measurement" },
    { to: "/patterns/gradients", label: "Gradients", tip: "Smooth ramps to check banding and tonal transitions" },
    { to: "/patterns/geometry", label: "Geometry", tip: "Crosshatch grid for convergence and alignment" },
    { to: "/patterns/colorchecker", label: "ColorChecker", tip: "Classic 24-patch reference for ΔE verification" },
  ]},
  { group: "Measurements", items: [
    { to: "/live", label: "Live View", tip: "Real-time colorimeter readings — xy chromaticity, luminance, CCT" },
    { to: "/measurements/precal", label: "Pre-Cal Checks", tip: "Black/white clipping, ABL detection, contrast ratio, peak luminance" },
    { to: "/measurements/cie", label: "Chromaticity", tip: "CIE 1931 xy diagram — measured gamut vs target color space" },
    { to: "/measurements/gamma", label: "Gamma / EOTF", tip: "Electro-optical transfer function — target curve vs measured response" },
    { to: "/measurements/deltae", label: "Delta E", tip: "CIE ΔE2000 color difference for primaries, secondaries, and white" },
    { to: "/measurements/grayscale", label: "Grayscale", tip: "Luminance output per IRE level — linearity and tracking" },
    { to: "/measurements/saturation", label: "Saturation", tip: "Radar chart of color tracking at multiple saturation levels" },
    { to: "/measurements/colorchecker", label: "ColorChecker", tip: "Per-patch ΔE for the 24-patch ColorChecker reference" },
  ]},
];

export default function App() {
  const location = useLocation();
  const [open, setOpen] = useState<Record<string, boolean>>({});

  if (location.pathname === "/pattern") return <PatternRenderer />;

  const toggle = (group: string) => setOpen((prev) => ({ ...prev, [group]: !prev[group] }));

  return (
    <div className="flex h-screen">
      <nav className="w-48 p-4 flex flex-col gap-0.5 shrink-0 overflow-y-auto" style={{ background: "var(--surface)", borderRight: "1px solid var(--border)" }}>
        <h1 className="text-lg font-bold mb-3">AutoCal50</h1>
        {nav.map((section) => (
          <div key={section.group} className="mb-1">
            <button onClick={() => toggle(section.group)}
              className="w-full flex items-center justify-between px-3 py-1.5 text-xs font-semibold uppercase tracking-wider rounded hover:bg-white/5"
              style={{ color: "var(--muted)" }}>
              {section.group}
              <motion.span animate={{ rotate: open[section.group] ? 90 : 0 }} transition={{ duration: 0.15 }}
                className="text-[10px]">▶</motion.span>
            </button>
            <AnimatePresence initial={false}>
              {open[section.group] && (
                <motion.div
                  initial={{ height: 0, opacity: 0 }}
                  animate={{ height: "auto", opacity: 1 }}
                  exit={{ height: 0, opacity: 0 }}
                  transition={{ duration: 0.2, ease: "easeInOut" }}
                  style={{ overflow: "hidden" }}
                >
                  {section.items.map((n) => (
                    <NavLink key={n.to} to={n.to} end title={n.tip}
                      className="px-3 py-1.5 rounded text-sm block ml-1"
                      style={({ isActive }) => ({
                        background: isActive ? "var(--accent)" : "transparent",
                        color: isActive ? "#fff" : "var(--muted)",
                      })}>
                      {n.label}
                    </NavLink>
                  ))}
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        ))}
      </nav>
      <main className="flex-1 p-6 overflow-auto">
        <Routes>
          <Route path="/" element={<Devices />} />
          <Route path="/signal" element={<Signal />} />
          <Route path="/controls" element={<Controls />} />
          <Route path="/calibration/:mode" element={<Calibration />} />
          <Route path="/patterns" element={<Patterns />} />
          <Route path="/patterns/:category" element={<Patterns />} />
          <Route path="/live" element={<LiveView />} />
          <Route path="/measurements/:view" element={<Measurements />} />
        </Routes>
      </main>
    </div>
  );
}
