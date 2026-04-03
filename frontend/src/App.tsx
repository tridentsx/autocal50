import { NavLink, Route, Routes, useLocation } from "react-router-dom";
import Devices from "@/pages/Devices";
import Signal from "@/pages/Signal";
import Controls from "@/pages/Controls";
import Patterns from "@/pages/Patterns";
import PatternRenderer from "@/pages/PatternRenderer";
import LiveView from "@/pages/LiveView";
import Measurements from "@/pages/Measurements";

const nav = [
  { to: "/", label: "Devices" },
  { to: "/signal", label: "Signal" },
  { to: "/controls", label: "Controls" },
  { to: "/patterns", label: "Patterns" },
  { to: "/live", label: "Live View" },
  { to: "/measurements", label: "Measurements" },
];

export default function App() {
  const location = useLocation();

  // Pattern renderer is fullscreen, no chrome
  if (location.pathname === "/pattern") return <PatternRenderer />;

  return (
    <div className="flex h-screen">
      <nav className="w-48 p-4 flex flex-col gap-1 shrink-0" style={{ background: "var(--surface)", borderRight: "1px solid var(--border)" }}>
        <h1 className="text-lg font-bold mb-4">AutoCal50</h1>
        {nav.map((n) => (
          <NavLink key={n.to} to={n.to} end
            className="px-3 py-2 rounded text-sm"
            style={({ isActive }) => ({
              background: isActive ? "var(--accent)" : "transparent",
              color: isActive ? "#fff" : "var(--muted)",
            })}>
            {n.label}
          </NavLink>
        ))}
      </nav>
      <main className="flex-1 p-6 overflow-auto">
        <Routes>
          <Route path="/" element={<Devices />} />
          <Route path="/signal" element={<Signal />} />
          <Route path="/controls" element={<Controls />} />
          <Route path="/patterns" element={<Patterns />} />
          <Route path="/live" element={<LiveView />} />
          <Route path="/measurements" element={<Measurements />} />
        </Routes>
      </main>
    </div>
  );
}
