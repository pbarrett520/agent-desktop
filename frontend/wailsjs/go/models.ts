export namespace audit {
	
	export class Event {
	    timestamp: string;
	    session_id: string;
	    subscription_id?: string;
	    tier: string;
	    command: string;
	    proposed_by: string;
	    decision: string;
	    exit_code: number;
	    duration_ms: number;
	    output_hash: string;
	
	    static createFrom(source: any = {}) {
	        return new Event(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.session_id = source["session_id"];
	        this.subscription_id = source["subscription_id"];
	        this.tier = source["tier"];
	        this.command = source["command"];
	        this.proposed_by = source["proposed_by"];
	        this.decision = source["decision"];
	        this.exit_code = source["exit_code"];
	        this.duration_ms = source["duration_ms"];
	        this.output_hash = source["output_hash"];
	    }
	}

}

export namespace azure {
	
	export class Context {
	    installed: boolean;
	    logged_in: boolean;
	    tenant_id?: string;
	    subscription_id?: string;
	    subscription_name?: string;
	    user?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new Context(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.installed = source["installed"];
	        this.logged_in = source["logged_in"];
	        this.tenant_id = source["tenant_id"];
	        this.subscription_id = source["subscription_id"];
	        this.subscription_name = source["subscription_name"];
	        this.user = source["user"];
	        this.error = source["error"];
	    }
	}
	export class CostSummary {
	    currency: string;
	    amount_to_date: number;
	
	    static createFrom(source: any = {}) {
	        return new CostSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currency = source["currency"];
	        this.amount_to_date = source["amount_to_date"];
	    }
	}
	export class ResourceGroupSummary {
	    name: string;
	    location: string;
	    resource_count: number;
	
	    static createFrom(source: any = {}) {
	        return new ResourceGroupSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.location = source["location"];
	        this.resource_count = source["resource_count"];
	    }
	}
	export class Subscription {
	    id: string;
	    name: string;
	    tenant_id: string;
	    state: string;
	    is_default: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Subscription(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.tenant_id = source["tenant_id"];
	        this.state = source["state"];
	        this.is_default = source["is_default"];
	    }
	}
	export class VMPowerState {
	    name: string;
	    resource_group: string;
	    power_state: string;
	
	    static createFrom(source: any = {}) {
	        return new VMPowerState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.resource_group = source["resource_group"];
	        this.power_state = source["power_state"];
	    }
	}

}

export namespace config {
	
	export class Config {
	    api_key: string;
	    endpoint: string;
	    model: string;
	    execution_timeout: number;
	    mode: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.api_key = source["api_key"];
	        this.endpoint = source["endpoint"];
	        this.model = source["model"];
	        this.execution_timeout = source["execution_timeout"];
	        this.mode = source["mode"];
	    }
	}

}

export namespace conversation {
	
	export class Conversation {
	    id: string;
	    title: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	    messages: llm.Message[];
	
	    static createFrom(source: any = {}) {
	        return new Conversation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.messages = this.convertValues(source["messages"], llm.Message);
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
	export class Summary {
	    id: string;
	    title: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	    turn_count: number;
	
	    static createFrom(source: any = {}) {
	        return new Summary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.turn_count = source["turn_count"];
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

export namespace llm {
	
	export class ToolCall {
	    id: string;
	    name: string;
	    arguments: string;
	
	    static createFrom(source: any = {}) {
	        return new ToolCall(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.arguments = source["arguments"];
	    }
	}
	export class Message {
	    role: string;
	    content: string;
	    tool_calls?: ToolCall[];
	    tool_call_id?: string;
	
	    static createFrom(source: any = {}) {
	        return new Message(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role = source["role"];
	        this.content = source["content"];
	        this.tool_calls = this.convertValues(source["tool_calls"], ToolCall);
	        this.tool_call_id = source["tool_call_id"];
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
	export class ModelInfo {
	    id: string;
	    object: string;
	    owned_by: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.object = source["object"];
	        this.owned_by = source["owned_by"];
	    }
	}

}

