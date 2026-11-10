export namespace models {
	
	export class AppRule {
	    path: string;
	    name: string;
	    display?: string;
	    category?: string;
	    enabled: boolean;
	    is_enabled?: boolean;
	    is_custom?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.display = source["display"];
	        this.category = source["category"];
	        this.enabled = source["enabled"];
	        this.is_enabled = source["is_enabled"];
	        this.is_custom = source["is_custom"];
	    }
	}
	export class ConnectionItem {
	    id: string;
	    destination: string;
	    rule: string;
	    down_speed: string;
	    up_speed: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.destination = source["destination"];
	        this.rule = source["rule"];
	        this.down_speed = source["down_speed"];
	        this.up_speed = source["up_speed"];
	    }
	}
	export class ConnectionGroup {
	    process: string;
	    count: number;
	    down_speed: string;
	    up_speed: string;
	    children: ConnectionItem[];
	
	    static createFrom(source: any = {}) {
	        return new ConnectionGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.process = source["process"];
	        this.count = source["count"];
	        this.down_speed = source["down_speed"];
	        this.up_speed = source["up_speed"];
	        this.children = this.convertValues(source["children"], ConnectionItem);
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
	
	export class CustomList {
	    id: string;
	    name: string;
	    url?: string;
	    action?: string;
	    rule_type?: string;
	    outbound?: string;
	    enabled: boolean;
	    domains?: string[];
	    count: number;
	    last_updated?: string;
	    updated_at?: string;
	
	    static createFrom(source: any = {}) {
	        return new CustomList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.url = source["url"];
	        this.action = source["action"];
	        this.rule_type = source["rule_type"];
	        this.outbound = source["outbound"];
	        this.enabled = source["enabled"];
	        this.domains = source["domains"];
	        this.count = source["count"];
	        this.last_updated = source["last_updated"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class Node {
	    id: string;
	    name: string;
	    protocol: string;
	    protocol_label?: string;
	    server: string;
	    port: number;
	    uuid?: string;
	    password?: string;
	    flow?: string;
	    security?: string;
	    sni?: string;
	    fp?: string;
	    pbk?: string;
	    sid?: string;
	    spx?: string;
	    alpn?: string[];
	    path?: string;
	    host?: string;
	    service_name?: string;
	    network?: string;
	    subscription_id?: string;
	    subscription_name?: string;
	    ping: number;
	    country?: string;
	    private_key?: string;
	    peer_public_key?: string;
	    local_address?: string;
	    reserved?: number[];
	    raw_outbound?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new Node(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.protocol = source["protocol"];
	        this.protocol_label = source["protocol_label"];
	        this.server = source["server"];
	        this.port = source["port"];
	        this.uuid = source["uuid"];
	        this.password = source["password"];
	        this.flow = source["flow"];
	        this.security = source["security"];
	        this.sni = source["sni"];
	        this.fp = source["fp"];
	        this.pbk = source["pbk"];
	        this.sid = source["sid"];
	        this.spx = source["spx"];
	        this.alpn = source["alpn"];
	        this.path = source["path"];
	        this.host = source["host"];
	        this.service_name = source["service_name"];
	        this.network = source["network"];
	        this.subscription_id = source["subscription_id"];
	        this.subscription_name = source["subscription_name"];
	        this.ping = source["ping"];
	        this.country = source["country"];
	        this.private_key = source["private_key"];
	        this.peer_public_key = source["peer_public_key"];
	        this.local_address = source["local_address"];
	        this.reserved = source["reserved"];
	        this.raw_outbound = source["raw_outbound"];
	    }
	}
	export class ProcessInfo {
	    name: string;
	    path: string;
	    category: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProcessInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.category = source["category"];
	        this.enabled = source["enabled"];
	    }
	}
	export class Settings {
	    split_mode: string;
	    active_node_id?: string;
	    privacy_blur: boolean;
	    auto_start: boolean;
	    minimize_to_tray: boolean;
	    strict_route: boolean;
	    bypass_lan: boolean;
	    bypass_ru: boolean;
	    adblock: boolean;
	    mtu?: number;
	    tun_mtu?: number;
	    remote_dns?: string;
	    direct_dns?: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.split_mode = source["split_mode"];
	        this.active_node_id = source["active_node_id"];
	        this.privacy_blur = source["privacy_blur"];
	        this.auto_start = source["auto_start"];
	        this.minimize_to_tray = source["minimize_to_tray"];
	        this.strict_route = source["strict_route"];
	        this.bypass_lan = source["bypass_lan"];
	        this.bypass_ru = source["bypass_ru"];
	        this.adblock = source["adblock"];
	        this.mtu = source["mtu"];
	        this.tun_mtu = source["tun_mtu"];
	        this.remote_dns = source["remote_dns"];
	        this.direct_dns = source["direct_dns"];
	    }
	}
	export class Subscription {
	    id: string;
	    name: string;
	    url: string;
	    updated_at: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new Subscription(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.url = source["url"];
	        this.updated_at = source["updated_at"];
	        this.count = source["count"];
	    }
	}
	export class TrafficStats {
	    connected: boolean;
	    down_speed: number;
	    up_speed: number;
	    total_down: number;
	    total_up: number;
	    ping: number;
	    ip: string;
	    node_name: string;
	    mode: string;
	
	    static createFrom(source: any = {}) {
	        return new TrafficStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.down_speed = source["down_speed"];
	        this.up_speed = source["up_speed"];
	        this.total_down = source["total_down"];
	        this.total_up = source["total_up"];
	        this.ping = source["ping"];
	        this.ip = source["ip"];
	        this.node_name = source["node_name"];
	        this.mode = source["mode"];
	    }
	}

}

