// Typed wrappers around Wails runtime calls.
// Wails injects window.go at runtime with bound methods.

declare global {
  interface Window {
    go: {
      main: {
        App: Record<string, (...args: any[]) => Promise<any>>;
      };
    };
  }
}

function call<T>(method: string, ...args: any[]): Promise<T> {
  return window.go.main.App[method](...args);
}

export const api = {
  listProjectorDrivers: () => call<string[]>("ListProjectorDrivers"),
  connectProjector: (driver: string, cfg: Record<string, any>) =>
    call<void>("ConnectProjector", driver, cfg),
  disconnectProjector: () => call<void>("DisconnectProjector"),
  getCapabilities: () => call<any>("GetProjectorCapabilities"),
  getControl: (id: string) => call<any>("GetProjectorControl", id),
  setControl: (id: string, value: any) =>
    call<void>("SetProjectorControl", id, value),

  listMeterDrivers: () => call<string[]>("ListMeterDrivers"),
  connectMeter: (driver: string, cfg: Record<string, any>) =>
    call<void>("ConnectMeter", driver, cfg),
  disconnectMeter: () => call<void>("DisconnectMeter"),
  measure: () => call<any>("Measure"),

  listTransportDrivers: () => call<string[]>("ListTransportDrivers"),
  connectTransport: (driver: string, cfg: Record<string, any>) =>
    call<void>("ConnectTransport", driver, cfg),
  disconnectTransport: () => call<void>("DisconnectTransport"),
  getSignalFormats: () => call<any[]>("GetSignalFormats"),
  applySignalFormat: (f: any) => call<void>("ApplySignalFormat", f),
  currentSignalFormat: () => call<any>("CurrentSignalFormat"),

  // Patterns
  getPatternBattery: () => call<any[]>("GetPatternBattery"),
  setActivePattern: (p: any) => call<void>("SetActivePattern", p),
  getActivePattern: () => call<any>("GetActivePattern"),

  // Display
  listDisplays: () => call<any[]>("ListDisplays"),
  setPatternDisplay: (name: string) => call<void>("SetPatternDisplay", name),
  getPatternDisplay: () => call<string>("GetPatternDisplay"),

  // Serial
  listSerialPorts: () => call<string[]>("ListSerialPorts"),

  // EDID
  getDisplayEDID: (name: string) => call<any>("GetDisplayEDID", name),
};
