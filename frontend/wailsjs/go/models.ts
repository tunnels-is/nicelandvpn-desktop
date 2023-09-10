export namespace core {
	
	export class CONTROLLER_SESSION_REQUEST {
	    GROUP: number;
	    ROUTERID: number;
	    SESSIONID: number;
	    XGROUP: number;
	    XROUTERID: number;
	    DEVICEID: number;
	
	    static createFrom(source: any = {}) {
	        return new CONTROLLER_SESSION_REQUEST(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.GROUP = source["GROUP"];
	        this.ROUTERID = source["ROUTERID"];
	        this.SESSIONID = source["SESSIONID"];
	        this.XGROUP = source["XGROUP"];
	        this.XROUTERID = source["XROUTERID"];
	        this.DEVICEID = source["DEVICEID"];
	    }
	}
	export class CONFIG_FORM {
	    DNS1: string;
	    DNS2: string;
	    ManualRouter: boolean;
	    Region: string;
	    Version: string;
	    RouterFilePath: string;
	    DebugLogging: boolean;
	    AutoReconnect: boolean;
	    KillSwitch: boolean;
	    PrevSlot?: CONTROLLER_SESSION_REQUEST;
	    DisableIPv6OnConnect: boolean;
	    CloseConnectionsOnConnect: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CONFIG_FORM(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.DNS1 = source["DNS1"];
	        this.DNS2 = source["DNS2"];
	        this.ManualRouter = source["ManualRouter"];
	        this.Region = source["Region"];
	        this.Version = source["Version"];
	        this.RouterFilePath = source["RouterFilePath"];
	        this.DebugLogging = source["DebugLogging"];
	        this.AutoReconnect = source["AutoReconnect"];
	        this.KillSwitch = source["KillSwitch"];
	        this.PrevSlot = this.convertValues(source["PrevSlot"], CONTROLLER_SESSION_REQUEST);
	        this.DisableIPv6OnConnect = source["DisableIPv6OnConnect"];
	        this.CloseConnectionsOnConnect = source["CloseConnectionsOnConnect"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
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

