# AutoCal50

A cross-platform projector autocalibration tool built with Go and React. Connect a colorimeter and a projector, run test patterns, take measurements, and calibrate to industry color standards — all from a single desktop app.

![Wails](https://img.shields.io/badge/Wails-v2-blue)
![Go](https://img.shields.io/badge/Go-1.23-00ADD8)
![React](https://img.shields.io/badge/React-19-61DAFB)
![License](https://img.shields.io/badge/license-MIT-green)

## Features

- **Device Management** — Connect projectors (serial/network), colorimeters, and signal transports with a driver-based plugin model
- **Test Pattern Generator** — Full calibration pattern battery: grayscale windows, saturation sweeps, clipping detection, ABL measurement, HDR EOTF steps, ColorChecker, contrast ratio, and more
- **Live Measurements** — Real-time colorimeter readings with CCT gauge, xy chromaticity, and luminance display
- **Projector Controls** — Dynamic capability-based UI that adapts to each projector's available controls (brightness, contrast, RGB gain/bias, gamma presets, etc.)
- **Signal Management** — Resolution, refresh rate, encoding (RGB/YCbCr), bit depth, and HDR mode control
- **Measurement Analysis** — CIE 1931 chromaticity diagram with colored gamut background, gamma/EOTF tracking, delta E bar charts, grayscale luminance, saturation sweep radar, and ColorChecker verification
- **Color Space Standards** — Calibrate against Rec. 709, DCI-P3, Display P3, BT.2020, HDR10 (PQ), or ACES AP1
- **HDR Support** — ST.2084 PQ EOTF curve with configurable peak luminance (1000/4000 nits)
- **Pre-Cal Diagnostics** — Black/white clipping detection, ABL behavior analysis, contrast ratio measurement
- **EDID Reading** — Display identification data including manufacturer, model, bit depth, HDR capability, BT.2020/P3 support flags
- **Cross-Platform** — Runs on Linux, Windows, and macOS

## Screenshots

The app has a sidebar navigation with the following pages:

| Page | Description |
|------|-------------|
| Devices | Connect/disconnect projector, meter, and transport drivers |
| Signal | Configure resolution, refresh rate, encoding, bit depth, HDR |
| Controls | Dynamic projector controls rendered from driver capabilities |
| Patterns | Test pattern browser with display output selection and EDID info |
| Live View | Real-time measurement display with CCT gauge |
| Measurements | Full calibration analysis with 6 chart types and pre-cal checks |

## Getting Started

### Prerequisites

- [Go 1.23+](https://go.dev/dl/)
- [Node.js 22+](https://nodejs.org/)
- [Wails v2](https://wails.io/docs/gettingstarted/installation)

Linux also requires:
```bash
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev
```

### Development

```bash
# Run in live development mode with hot reload
make dev
```

This starts a Vite dev server for the frontend and the Wails backend. Changes to `.tsx` files hot-reload instantly. Go changes trigger a rebuild.

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build for a specific platform
make build-linux
make build-windows
make build-mac
```

Binaries are output to `build/bin/`.

## Architecture

```
autocal50/
├── app.go                  # Wails API surface — binds Go methods to frontend
├── main.go                 # Wails app bootstrap
├── internal/
│   ├── core/               # Central Manager orchestrating all subsystems
│   ├── projector/          # Projector driver interface + implementations
│   ├── meter/              # Colorimeter driver interface + implementations
│   ├── transport/          # Signal/HDMI transport driver interface
│   ├── pattern/            # Test pattern battery generation
│   ├── display/            # Display enumeration + EDID parsing (per-platform)
│   ├── calibration/        # Calibration routines (grayscale, gamma, CMS)
│   ├── store/              # Profile and measurement persistence
│   ├── events/             # Wails event constants
│   └── serialutil/         # Serial port discovery
└── frontend/
    └── src/
        ├── api/            # Typed Wails API wrappers
        ├── hooks/          # React hooks (live measurement events)
        ├── pages/          # UI pages (Devices, Signal, Controls, etc.)
        └── types/          # Shared TypeScript types
```

### Driver Model

Hardware is abstracted behind capability-based driver interfaces. Each driver (projector, meter, transport) exposes:
- A `Connect`/`Disconnect` lifecycle
- Capability discovery (what controls exist, what formats are supported)
- Generic get/set operations

The UI renders controls dynamically from driver metadata, so adding a new projector model doesn't require frontend changes.

### Pattern Window

Test patterns render in a separate fullscreen window on the projector display, while the control UI stays on your primary screen. Patterns are drawn on a canvas at native resolution.

### Test Pattern Battery

| Group | Patterns | Purpose |
|-------|----------|---------|
| Grayscale | 0–100% IRE in 5% steps (solid + 18% window) | White balance and gamma measurement |
| Clipping | 0–5% and 95–100% in 1% steps | Near-black/near-white visibility |
| Primaries & Secondaries | R/G/B/C/M/Y full-field + 18% windows | Gamut measurement |
| Saturation Sweeps | 20/40/60/80/100% × 6 colors | Color tracking accuracy |
| Peak vs Size | 1–100% white windows | ABL (auto brightness limiter) detection |
| HDR EOTF Steps | 1–4000 nits as 18% windows | PQ curve verification |
| Contrast Ratio | Black, white, 4×4 and fine checkerboard | Sequential and simultaneous contrast |
| Gradients | Gray, R, G, B, C, M, Y ramps | Banding and monotonicity |
| ColorChecker | Classic 24-patch grid | Delta E verification |
| Geometry | Crosshatch grid | Convergence and alignment |

## CI/CD

Releases are built automatically via GitHub Actions. Creating a release triggers parallel builds for Linux (amd64), Windows (amd64), and macOS (universal). Binaries are attached to the release as assets.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.23 |
| Desktop Shell | Wails v2 |
| Frontend | React 19 + TypeScript |
| Styling | Tailwind CSS v4 |
| Charts | ECharts (echarts-for-react) |
| 3D (planned) | react-three-fiber + drei |
| Build | Vite |

## License

MIT
