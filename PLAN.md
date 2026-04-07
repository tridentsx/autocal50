# AutoCal50 — Projector Autocalibration Tool

## Remaining Work — Priority Order

### Critical (can't calibrate without these)

1. ~~**Measurement storage**~~ — ✅ JSON session persistence with auto-save
2. ~~**Automated measure-adjust loop**~~ — ✅ Calibration engine with iterative white balance, gamma sweep, verification
3. ~~**Profile generation from measurements**~~ — ✅ ICC profile from measured primaries + TRC, auto-install per platform
4. ~~**Pre-cal diagnostics automation**~~ — ✅ RunPreCal in calibration engine

### Important (usable without, but limited)

5. ~~**Local transport UI flow**~~ — ✅ Devices page shows connector picker for local-output
6. ~~**Signal page ↔ local transport**~~ — ✅ 10-bit and HDR modes pass through from local transport
7. ~~**Meter calibration**~~ — ✅ Dark cal trigger via ArgyllCMS spotread
8. ~~**Error handling / retry**~~ — ✅ Generic retry with exponential backoff on measurements and projector commands

### Polish

9. ~~**Results export**~~ — ✅ CSV measurement export from session
10. ~~**Settings persistence**~~ — ✅ Save/load last-used devices and preferences
11. **Measurement charts with real data** — Chart components exist but need real data from calibration runs.
12. **Additional projector drivers** — Only XGIMI RS232 exists.

---

## Tech Stack

| Layer | Choice |
|-------|--------|
| Backend / Core | Go |
| Desktop Shell | Wails v2 |
| Frontend | React + TypeScript |
| UI Components | shadcn/ui |
| Styling | Tailwind CSS |
| 2D Charts | ECharts |
| 3D Visualization | react-three-fiber + @react-three/drei |
| Meter (initial) | ArgyllCMS wrapper |
| Driver Model | Capability-based internal registry |
| Plugin Model | Internal registry first, external process plugins later |

---

## Project Structure

```
autocal50/
  main.go
  app.go
  wails.json

  internal/
    core/           # orchestration layer tying everything together
    projector/      # serial/network projector drivers, plugin-style abstraction
    meter/          # colorimeter integration (Argyll wrapper first)
    transport/      # HDMI/output mode handling (4:4:4, 4:2:2, bit depth)
    pattern/        # test pattern generation control
    calibration/    # grayscale, white point, gamma, CMS routines
    store/          # profiles, presets, saved measurements, session state
    events/         # push live updates to frontend via Wails events

  frontend/
    src/
      app/
      components/
      pages/
        Devices/
        Signal/
        ProjectorControls/
        Calibration/
        Report/
      hooks/
      types/
      api/
```

---

## Architecture

### Backend Responsibilities

| Package | Responsibility |
|---------|---------------|
| `internal/core` | Central `Manager` struct. Orchestrates projector, meter, transport, calibration. Exposes methods to `App` struct for Wails binding. |
| `internal/projector` | `Driver` interface. Capability-based control model. Built-in driver registry (Epson, JVC, Sony, etc). |
| `internal/meter` | `Driver` interface. `Reading` struct (XYZ, luminance, CCT). ArgyllCMS wrapper as first implementation. |
| `internal/transport` | `Driver` interface. `SignalFormat` struct (resolution, refresh, encoding, bit depth, HDR). |
| `internal/pattern` | Test pattern generation. Runs separately from main UI window (fullscreen, borderless, precise output on projector display). |
| `internal/calibration` | Grayscale, white point, gamma, CMS calibration routines. Consumes meter readings + projector controls. |
| `internal/store` | Persistence: profiles, presets, measurement sessions. |
| `internal/events` | Wails event emission. Backend emits events like `measurement:update`, frontend subscribes. |

### Wails App Surface

`App` struct exposes UI-facing API. All hardware logic stays in `internal/`. Example methods:

- `ListProjectorDrivers() []string`
- `ConnectProjector(driver string, cfg map[string]any) error`
- `GetProjectorCapabilities() (Capabilities, error)`
- `SetProjectorControl(id string, value any) error`
- `ConnectMeter(driver string, cfg map[string]any) error`
- `Measure() (Reading, error)`
- `GetSignalFormats() ([]SignalFormat, error)`
- `ApplySignalFormat(f SignalFormat) error`
- `StartCalibration(workflow string) error`

Live data uses Wails events (not polling).

### Driver / Plugin Model

**Phase 1: Internal registry (compiled-in drivers)**

```go
var ProjectorDrivers = map[string]projector.DriverFactory{
    "epson-escvp21": NewEpsonDriver,
    "jvc-rs232":     NewJVCDriver,
    "sony-pjtalk":   NewSonyDriver,
}
```

**Later: External process plugins** via gRPC/stdio helper processes. Avoid Go `plugin` package.

### Key Design Decision: Capability-Based Controls

Model projector controls as generic capabilities, not hardcoded fields.

Each projector driver exposes a list of `Control` structs with metadata (type, min, max, options). The UI renders them dynamically. This means different projectors can expose `rgb_gain_r`, `gamma_preset`, `laser_power`, `dynamic_contrast`, etc. without frontend changes.

### Key Design Decision: Pattern Window

Pattern generation must be separate from the main Wails UI window. Calibration patches need fullscreen, borderless, precise output on the projector display. The control UI stays on a separate screen/window.

---

## Core Interfaces

### projector.Driver

```go
type Capabilities struct {
    Controls     []Control
    SupportsModes []string
}

type Control struct {
    ID       string
    Label    string
    Type     string   // "slider", "enum", "toggle"
    Min      float64
    Max      float64
    Step     float64
    Options  []string
    ReadOnly bool
}

type Driver interface {
    Name() string
    Connect(ctx context.Context, cfg map[string]any) error
    Disconnect() error
    Capabilities(ctx context.Context) (Capabilities, error)
    GetControl(ctx context.Context, id string) (any, error)
    SetControl(ctx context.Context, id string, value any) error
}
```

### meter.Driver

```go
type Reading struct {
    X, Y, Z  float64
    Luminance float64
    CCT       float64
    Timestamp int64
}

type Driver interface {
    Name() string
    Connect(ctx context.Context, cfg map[string]any) error
    Disconnect() error
    Measure(ctx context.Context) (Reading, error)
}
```

### transport.Driver

```go
type SignalFormat struct {
    Resolution string
    RefreshHz  float64
    Encoding   string // "RGB", "YCbCr444", "YCbCr422", "YCbCr420"
    BitDepth   int    // 8, 10, 12
    HDR        bool
}

type Driver interface {
    Name() string
    Connect(ctx context.Context, cfg map[string]any) error
    Disconnect() error
    GetFormats(ctx context.Context) ([]SignalFormat, error)
    ApplyFormat(ctx context.Context, f SignalFormat) error
    CurrentFormat(ctx context.Context) (SignalFormat, error)
}
```

---

## Frontend Pages

### 1. Devices
- Projector driver selection + connect/disconnect
- Meter backend selection + connect/disconnect
- Transport backend selection + connect/disconnect
- Connection status indicators

### 2. Signal
- Resolution, refresh rate
- Encoding: RGB / 4:4:4 / 4:2:2 / 4:2:0
- Bit depth: 8 / 10 / 12
- Active mode verification

### 3. Projector Controls
- Dynamically rendered from `Capabilities`
- Picture mode, brightness, contrast, RGB gain/bias, gamma presets
- Dynamic features on/off
- All controls auto-generated from driver metadata

### 4. Calibration
- White point workflow
- Grayscale workflow
- Gamma workflow
- CMS (later)

### 5. Report / Live View

Layout:
```
┌─────────────────────┬─────────────────────┐
│  2D CIE Gamut Chart │  3D Gamut Volume     │
│  (target + before   │  (wireframe target,  │
│   + after triangles,│   translucent before │
│   white point,      │   & after meshes,    │
│   saturation dots)  │   rotate/zoom)       │
├─────────────────────┴─────────────────────┤
│  RGB Balance — before vs after over IRE    │
├────────────────────────────────────────────┤
│  Gamma / EOTF — target, before, after      │
├────────────────────────────────────────────┤
│  DeltaE Bars — grayscale + color patches   │
│  before vs after + summary stats           │
└────────────────────────────────────────────┘

Target selector: [ Rec.709 | DCI-P3 | BT.2020 | HDR ]
```

#### 3D Gamut View Details
- Target gamut: thin wireframe
- Before: transparent warm-colored (red/orange) surface
- After: transparent cool-colored (green/blue) surface
- Toggle each layer on/off
- Free rotation, reset camera
- Switch between SDR and HDR targets
- Inspect points numerically (tooltips)

#### Data Model (frontend types)

```typescript
type ColorPoint = {
  label: string;
  x: number;
  y: number;
  Y: number;
  X?: number;
  Z?: number;
};

type CalibrationDataset = {
  name: string; // "before" | "after"
  whitePoint: ColorPoint;
  primaries: ColorPoint[];
  secondaries: ColorPoint[];
  grayscale: ColorPoint[];
  gamma: { input: number; measured: number; target: number }[];
  deltaE: { label: string; value: number }[];
};
```

Backend sends measurement datasets. Frontend computes visual representation.

---

## Implementation Phases

### Phase 1 — App Shell + Mock Drivers

**Goal:** Wails app opens, all pages render, event flow works end-to-end with fake data.

- [ ] Re-initialize frontend with React + TypeScript + Vite
- [ ] Install Tailwind, shadcn/ui, ECharts, react-three-fiber
- [ ] Create `internal/` package structure with interfaces
- [ ] Implement mock projector driver (exposes fake controls)
- [ ] Implement mock meter driver (returns simulated readings)
- [ ] Implement mock transport driver (returns fake formats)
- [ ] Implement `internal/core.Manager` wiring mock drivers
- [ ] Wire `App` struct methods to `core.Manager`
- [ ] Set up Wails event emission (measurement:update every second)
- [ ] Build Devices page (list drivers, connect/disconnect)
- [ ] Build Signal page (format selector)
- [ ] Build Projector Controls page (dynamic control rendering)
- [ ] Build Live View page (live simulated readings)

**Milestone:** App opens. Devices page shows mock projector, mock meter, mock transport. Controls page renders dynamically. Live View shows simulated measurements updating via events.

### Phase 2 — Real Hardware + Manual Controls

**Goal:** Connect to a real projector and meter. Read and write controls. Take real measurements.

- [ ] Implement first real projector driver (e.g., JVC RS232 or Epson ESC/VP21)
- [ ] Implement ArgyllCMS meter wrapper (`spotread` / `colprof` CLI integration)
- [ ] Serial port discovery and configuration UI
- [ ] Manual projector controls page with real read/write
- [ ] Live measurement display with real meter data
- [ ] Error handling and reconnection logic

**Milestone:** Connect to a real projector over serial. Adjust controls from the app. Take a real color measurement and display it.

### Phase 3 — Calibration + Profiles + Reporting

**Goal:** Automated grayscale calibration. Save/load profiles. Generate reports.

- [ ] Implement grayscale autocal routine (measure → adjust → iterate)
- [ ] White point calibration workflow
- [ ] Gamma calibration workflow
- [ ] Save/load calibration profiles (`internal/store`)
- [ ] Build 2D CIE gamut chart (before/after)
- [ ] Build RGB balance chart (before/after)
- [ ] Build gamma/EOTF chart (before/after)
- [ ] Build DeltaE bar chart
- [ ] Calibration report page with all charts

**Milestone:** Run a grayscale autocal on a real projector. View before/after results in charts. Save the profile.

### Phase 4 — 3D Visualization + Advanced Features

**Goal:** Premium visualization. Transport control. CMS. HDR.

- [ ] Build 3D gamut volume view (react-three-fiber)
  - Target wireframe, before/after translucent meshes
  - Layer toggles, rotation, camera reset
  - Target space selector (Rec.709, DCI-P3, BT.2020)
- [ ] Implement transport driver for HDMI mode testing
- [ ] CMS (color management system) calibration
- [ ] HDR workflows (PQ, HLG)
- [ ] Pattern window (separate fullscreen window for test patches)
- [ ] Export reports (PDF or image)

### Phase 5 — Polish + Additional Drivers

- [ ] Additional projector drivers (Sony PJ Talk, etc.)
- [ ] Native meter support (beyond ArgyllCMS wrapper)
- [ ] External process plugin system (gRPC/stdio)
- [ ] Settings / preferences page
- [ ] Undo/redo for control changes
- [ ] Session history and comparison
- [ ] CGATS (.ti1/.ti3/.cal) import/export for DisplayCAL/ArgyllCMS interoperability
- [ ] YCbCr 4:2:0/4:2:2 output mode to match Blu-ray player signal format (verify projector's chroma processing path)

---

## Immediate Next Step

**Phase 1, Task 1:** Re-initialize the frontend with React + TypeScript, set up the project structure, create the Go interfaces and mock drivers, and get the Devices page rendering with mock data.
