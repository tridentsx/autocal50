import { useEffect, useRef } from "react";
import { useWailsEvent } from "@/hooks/useWailsEvent";
import type { Pattern, RGB } from "@/types";

function rgb(c: RGB) {
  return `rgb(${c.r},${c.g},${c.b})`;
}

function draw(canvas: HTMLCanvasElement, p: Pattern) {
  const ctx = canvas.getContext("2d")!;
  const w = canvas.width;
  const h = canvas.height;

  // Default black background
  ctx.fillStyle = "#000";
  ctx.fillRect(0, 0, w, h);

  switch (p.type) {
    case "solid":
      ctx.fillStyle = rgb(p.color!);
      ctx.fillRect(0, 0, w, h);
      break;

    case "window": {
      if (p.bgColor) { ctx.fillStyle = rgb(p.bgColor); ctx.fillRect(0, 0, w, h); }
      const pct = (p.windowPc ?? 18) / 100;
      const side = Math.sqrt(pct);
      const ww = w * side, wh = h * side;
      ctx.fillStyle = rgb(p.color!);
      ctx.fillRect((w - ww) / 2, (h - wh) / 2, ww, wh);
      break;
    }

    case "gradient": {
      const steps = p.steps ?? [];
      const sw = w / steps.length;
      for (let i = 0; i < steps.length; i++) {
        ctx.fillStyle = rgb(steps[i]);
        ctx.fillRect(Math.floor(i * sw), 0, Math.ceil(sw) + 1, h);
      }
      break;
    }

    case "grid": {
      if (p.bgColor) { ctx.fillStyle = rgb(p.bgColor); ctx.fillRect(0, 0, w, h); }
      ctx.strokeStyle = rgb(p.color!);
      ctx.lineWidth = 1;
      const cols = p.cols ?? 16, rows = p.rows ?? 9;
      for (let i = 0; i <= cols; i++) {
        const x = Math.round(i * w / cols) + 0.5;
        ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, h); ctx.stroke();
      }
      for (let i = 0; i <= rows; i++) {
        const y = Math.round(i * h / rows) + 0.5;
        ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(w, y); ctx.stroke();
      }
      break;
    }

    case "checker": {
      const cols = p.cols ?? 6, rows = p.rows ?? 4;
      const palette = p.steps ?? [];
      const cw = w / cols, ch = h / rows;
      for (let r = 0; r < rows; r++) {
        for (let c = 0; c < cols; c++) {
          const idx = r * cols + c;
          ctx.fillStyle = idx < palette.length ? rgb(palette[idx]) : "#000";
          ctx.fillRect(Math.floor(c * cw), Math.floor(r * ch), Math.ceil(cw) + 1, Math.ceil(ch) + 1);
        }
      }
      break;
    }
  }
}

export default function PatternRenderer() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const pattern = useWailsEvent<Pattern>("pattern:update");

  useEffect(() => {
    const resize = () => {
      const c = canvasRef.current;
      if (c) { c.width = window.innerWidth; c.height = window.innerHeight; }
    };
    resize();
    window.addEventListener("resize", resize);
    return () => window.removeEventListener("resize", resize);
  }, []);

  useEffect(() => {
    const c = canvasRef.current;
    if (c && pattern) draw(c, pattern);
  }, [pattern]);

  // ESC to exit fullscreen
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") document.exitFullscreen?.();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, []);

  return (
    <canvas
      ref={canvasRef}
      style={{ position: "fixed", inset: 0, background: "#000", cursor: "none" }}
    />
  );
}
