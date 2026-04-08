export namespace dto {
	
	export class DownloadTaskDTO {
	    id: string;
	    url: string;
	    status: string;
	    savePath: string;
	    // Go type: time
	    startTime: any;
	    // Go type: time
	    completeTime: any;
	    // Go type: time
	    updatedAt: any;
	    error: string;
	    name: string;
	    // Go type: struct { Current int "json:\"current\""; Total int "json:\"total\"" }
	    progress: any;
	
	    static createFrom(source: any = {}) {
	        return new DownloadTaskDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.url = source["url"];
	        this.status = source["status"];
	        this.savePath = source["savePath"];
	        this.startTime = this.convertValues(source["startTime"], null);
	        this.completeTime = this.convertValues(source["completeTime"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.error = source["error"];
	        this.name = source["name"];
	        this.progress = this.convertValues(source["progress"], Object);
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

export namespace history {
	
	export class Manager {
	
	
	    static createFrom(source: any = {}) {
	        return new Manager(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

export namespace library {
	
	export class Manga {
	    name: string;
	    path: string;
	    previewImg: string;
	    imagesCount: number;
	    images?: string[];
	
	    static createFrom(source: any = {}) {
	        return new Manga(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.previewImg = source["previewImg"];
	        this.imagesCount = source["imagesCount"];
	        this.images = source["images"];
	    }
	}

}

export namespace logger {
	
	export class LogInfo {
	    dir: string;
	    currentFile: string;
	    sizeBytes: number;
	    backups: string[];
	    maxSizeMB: number;
	    maxBackups: number;
	    maxAgeDays: number;
	    compress: boolean;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new LogInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dir = source["dir"];
	        this.currentFile = source["currentFile"];
	        this.sizeBytes = source["sizeBytes"];
	        this.backups = source["backups"];
	        this.maxSizeMB = source["maxSizeMB"];
	        this.maxBackups = source["maxBackups"];
	        this.maxAgeDays = source["maxAgeDays"];
	        this.compress = source["compress"];
	        this.updatedAt = source["updatedAt"];
	    }
	}

}

export namespace task {
	
	export class DownloadTask {
	    id: string;
	    url: string;
	    status: string;
	    savePath: string;
	    // Go type: time
	    startTime: any;
	    // Go type: time
	    completeTime: any;
	    // Go type: time
	    updatedAt: any;
	    error: string;
	    name: string;
	    // Go type: struct { Current int "json:\"current\""; Total int "json:\"total\"" }
	    progress: any;
	
	    static createFrom(source: any = {}) {
	        return new DownloadTask(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.url = source["url"];
	        this.status = source["status"];
	        this.savePath = source["savePath"];
	        this.startTime = this.convertValues(source["startTime"], null);
	        this.completeTime = this.convertValues(source["completeTime"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.error = source["error"];
	        this.name = source["name"];
	        this.progress = this.convertValues(source["progress"], Object);
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

