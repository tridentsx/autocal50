export type Control = {
  id: string;
  label: string;
  type: "slider" | "enum" | "toggle";
  min: number;
  max: number;
  step: number;
  options: string[];
  readOnly: boolean;
};

export type Capabilities = {
  controls: Control[];
  supportsModes: string[];
};

export type Reading = {
  x: number;
  y: number;
  z: number;
  luminance: number;
  cct: number;
  timestamp: number;
};

export type SignalFormat = {
  resolution: string;
  refreshHz: number;
  encoding: string;
  bitDepth: number;
  hdr: boolean;
};

export type DeviceStatus = {
  name: string;
  driver: string;
  connected: boolean;
};

export type ColorPoint = {
  label: string;
  x: number;
  y: number;
  Y: number;
  X?: number;
  Z?: number;
};

export type CalibrationDataset = {
  name: string;
  whitePoint: ColorPoint;
  primaries: ColorPoint[];
  secondaries: ColorPoint[];
  grayscale: ColorPoint[];
  gamma: { input: number; measured: number; target: number }[];
  deltaE: { label: string; value: number }[];
};

export type RGB = { r: number; g: number; b: number };

export type Pattern = {
  id: string;
  name: string;
  group: string;
  type: "solid" | "window" | "gradient" | "grid" | "checker";
  color?: RGB;
  bgColor?: RGB;
  windowPc?: number;
  steps?: RGB[];
  cols?: number;
  rows?: number;
};

export type DisplayOutput = {
  name: string;
  connected: boolean;
  primary: boolean;
  width: number;
  height: number;
  x: number;
  y: number;
  refreshHz: number;
};
