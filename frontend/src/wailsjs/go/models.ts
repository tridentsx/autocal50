export namespace calibration {
	
	export class Adjustment {
	    controlId: string;
	    direction: string;
	    magnitude: string;
	    reason: string;
	
	    static createFrom(source: any = {}) {
	        return new Adjustment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.controlId = source["controlId"];
	        this.direction = source["direction"];
	        this.magnitude = source["magnitude"];
	        this.reason = source["reason"];
	    }
	}
	export class Advice {
	    deltaE: number;
	    deltaX: number;
	    deltaY: number;
	    deltaLum: number;
	    passed: boolean;
	    adjustments: Adjustment[];
	
	    static createFrom(source: any = {}) {
	        return new Advice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deltaE = source["deltaE"];
	        this.deltaX = source["deltaX"];
	        this.deltaY = source["deltaY"];
	        this.deltaLum = source["deltaLum"];
	        this.passed = source["passed"];
	        this.adjustments = this.convertValues(source["adjustments"], Adjustment);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ColorTarget {
	    x: number;
	    y: number;
	    luminance: number;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new ColorTarget(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	        this.luminance = source["luminance"];
	        this.label = source["label"];
	    }
	}
	export class Step {
	    kind: string;
	    label: string;
	    description: string;
	    targets: ColorTarget[];
	    tolerance: number;
	    optional: boolean;
	    requiresControls?: string[];
	
	    static createFrom(source: any = {}) {
	        return new Step(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.label = source["label"];
	        this.description = source["description"];
	        this.targets = this.convertValues(source["targets"], ColorTarget);
	        this.tolerance = source["tolerance"];
	        this.optional = source["optional"];
	        this.requiresControls = source["requiresControls"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace display {
	
	export class EDID {
	    manufacturer: string;
	    productCode: number;
	    serial: number;
	    year: number;
	    week: number;
	    version: string;
	    displayName: string;
	    serialString: string;
	    maxHRes: number;
	    maxVRes: number;
	    bitDepth: number;
	    digitalInput: boolean;
	    hdr: boolean;
	    bt2020: boolean;
	    p3: boolean;
	    rawHex: string;
	
	    static createFrom(source: any = {}) {
	        return new EDID(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.manufacturer = source["manufacturer"];
	        this.productCode = source["productCode"];
	        this.serial = source["serial"];
	        this.year = source["year"];
	        this.week = source["week"];
	        this.version = source["version"];
	        this.displayName = source["displayName"];
	        this.serialString = source["serialString"];
	        this.maxHRes = source["maxHRes"];
	        this.maxVRes = source["maxVRes"];
	        this.bitDepth = source["bitDepth"];
	        this.digitalInput = source["digitalInput"];
	        this.hdr = source["hdr"];
	        this.bt2020 = source["bt2020"];
	        this.p3 = source["p3"];
	        this.rawHex = source["rawHex"];
	    }
	}
	export class Output {
	    name: string;
	    connected: boolean;
	    primary: boolean;
	    width: number;
	    height: number;
	    x: number;
	    y: number;
	    refreshHz: number;
	
	    static createFrom(source: any = {}) {
	        return new Output(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.connected = source["connected"];
	        this.primary = source["primary"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.refreshHz = source["refreshHz"];
	    }
	}

}

export namespace meter {
	
	export class Reading {
	    x: number;
	    y: number;
	    z: number;
	    luminance: number;
	    cct: number;
	    timestamp: number;
	
	    static createFrom(source: any = {}) {
	        return new Reading(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	        this.z = source["z"];
	        this.luminance = source["luminance"];
	        this.cct = source["cct"];
	        this.timestamp = source["timestamp"];
	    }
	}

}

export namespace pattern {
	
	export class RGB {
	    r: number;
	    g: number;
	    b: number;
	
	    static createFrom(source: any = {}) {
	        return new RGB(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.r = source["r"];
	        this.g = source["g"];
	        this.b = source["b"];
	    }
	}
	export class Pattern {
	    id: string;
	    name: string;
	    group: string;
	    type: string;
	    color?: RGB;
	    bgColor?: RGB;
	    windowPc?: number;
	    steps?: RGB[];
	    cols?: number;
	    rows?: number;
	
	    static createFrom(source: any = {}) {
	        return new Pattern(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.group = source["group"];
	        this.type = source["type"];
	        this.color = this.convertValues(source["color"], RGB);
	        this.bgColor = this.convertValues(source["bgColor"], RGB);
	        this.windowPc = source["windowPc"];
	        this.steps = this.convertValues(source["steps"], RGB);
	        this.cols = source["cols"];
	        this.rows = source["rows"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace projector {
	
	export class Control {
	    id: string;
	    label: string;
	    type: string;
	    min: number;
	    max: number;
	    step: number;
	    options: string[];
	    readOnly: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Control(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.type = source["type"];
	        this.min = source["min"];
	        this.max = source["max"];
	        this.step = source["step"];
	        this.options = source["options"];
	        this.readOnly = source["readOnly"];
	    }
	}
	export class Capabilities {
	    controls: Control[];
	    supportsModes: string[];
	
	    static createFrom(source: any = {}) {
	        return new Capabilities(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.controls = this.convertValues(source["controls"], Control);
	        this.supportsModes = source["supportsModes"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace transport {
	
	export class SignalFormat {
	    resolution: string;
	    refreshHz: number;
	    encoding: string;
	    bitDepth: number;
	    hdr: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SignalFormat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.resolution = source["resolution"];
	        this.refreshHz = source["refreshHz"];
	        this.encoding = source["encoding"];
	        this.bitDepth = source["bitDepth"];
	        this.hdr = source["hdr"];
	    }
	}

}

