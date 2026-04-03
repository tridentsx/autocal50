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

