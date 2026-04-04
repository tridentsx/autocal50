import { useState } from "react";
import { useParams } from "react-router-dom";
import CalibrationWizard from "@/components/cal/CalibrationWizard";

export default function Calibration() {
  const { mode } = useParams<{ mode: string }>();
  const [standard, setStandard] = useState("rec709");
  const [started, setStarted] = useState(false);
  const calMode = mode === "automatic" ? "automatic" : "assisted";

  if (started) {
    return <CalibrationWizard mode={calMode} standard={standard} />;
  }

  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">
        {calMode === "automatic" ? "Automatic" : "Assisted"} Calibration
      </h2>
      <div className="rounded-lg p-6 max-w-lg" style={{ background: "var(--surface)", border: "1px solid var(--border)" }}>
        <p className="text-sm mb-4" style={{ color: "var(--muted)" }}>
          {calMode === "assisted"
            ? "Step-by-step guided calibration. You adjust projector controls while the app measures and advises."
            : "Fully automated. The app controls the projector and iterates until targets are met."}
        </p>

        <label className="block text-sm mb-2" style={{ color: "var(--muted)" }}>Target Standard</label>
        <select value={standard} onChange={(e) => setStandard(e.target.value)}
          className="w-full p-2 rounded mb-6"
          style={{ background: "var(--bg)", color: "var(--text)", border: "1px solid var(--border)" }}>
          <option value="rec709">Rec. 709 / sRGB</option>
          <option value="dcip3">DCI-P3</option>
          <option value="p3d65">Display P3</option>
          <option value="bt2020">BT.2020</option>
          <option value="hdr10">HDR10 (PQ)</option>
        </select>

        <button onClick={() => setStarted(true)}
          className="px-6 py-3 rounded-xl font-semibold text-white w-full"
          style={{ background: "var(--accent)" }}>
          Start Calibration
        </button>
      </div>
    </div>
  );
}
