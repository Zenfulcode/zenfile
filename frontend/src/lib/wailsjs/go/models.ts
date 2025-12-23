export namespace main {
	
	export class AppInfoResponse {
	    name: string;
	    version: string;
	    dataDir: string;
	    logFile: string;
	    converterBackend: string;
	    ffmpegVersion?: string;
	
	    static createFrom(source: any = {}) {
	        return new AppInfoResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.dataDir = source["dataDir"];
	        this.logFile = source["logFile"];
	        this.converterBackend = source["converterBackend"];
	        this.ffmpegVersion = source["ffmpegVersion"];
	    }
	}
	export class SupportedFormatsResponse {
	    videoFormats: string[];
	    imageFormats: string[];
	    backend: string;
	
	    static createFrom(source: any = {}) {
	        return new SupportedFormatsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.videoFormats = source["videoFormats"];
	        this.imageFormats = source["imageFormats"];
	        this.backend = source["backend"];
	    }
	}

}

export namespace models {
	
	export class BatchConversionRequest {
	    files: string[];
	    outputFormat: string;
	    outputDirectory: string;
	    namingMode: string;
	    customNames?: string[];
	    makeCopies: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BatchConversionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.files = source["files"];
	        this.outputFormat = source["outputFormat"];
	        this.outputDirectory = source["outputDirectory"];
	        this.namingMode = source["namingMode"];
	        this.customNames = source["customNames"];
	        this.makeCopies = source["makeCopies"];
	    }
	}
	export class ConversionResult {
	    success: boolean;
	    inputPath: string;
	    outputPath: string;
	    outputSize: number;
	    errorMessage?: string;
	    duration: number;
	
	    static createFrom(source: any = {}) {
	        return new ConversionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.inputPath = source["inputPath"];
	        this.outputPath = source["outputPath"];
	        this.outputSize = source["outputSize"];
	        this.errorMessage = source["errorMessage"];
	        this.duration = source["duration"];
	    }
	}
	export class BatchConversionResult {
	    totalFiles: number;
	    successCount: number;
	    failCount: number;
	    results: ConversionResult[];
	    totalDuration: number;
	
	    static createFrom(source: any = {}) {
	        return new BatchConversionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalFiles = source["totalFiles"];
	        this.successCount = source["successCount"];
	        this.failCount = source["failCount"];
	        this.results = this.convertValues(source["results"], ConversionResult);
	        this.totalDuration = source["totalDuration"];
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
	export class Conversion {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    inputPath: string;
	    outputPath: string;
	    inputFormat: string;
	    outputFormat: string;
	    fileType: string;
	    fileSize: number;
	    outputSize: number;
	    status: string;
	    errorMessage?: string;
	    progress: number;
	    // Go type: time
	    startedAt?: any;
	    // Go type: time
	    completedAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new Conversion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.inputPath = source["inputPath"];
	        this.outputPath = source["outputPath"];
	        this.inputFormat = source["inputFormat"];
	        this.outputFormat = source["outputFormat"];
	        this.fileType = source["fileType"];
	        this.fileSize = source["fileSize"];
	        this.outputSize = source["outputSize"];
	        this.status = source["status"];
	        this.errorMessage = source["errorMessage"];
	        this.progress = source["progress"];
	        this.startedAt = this.convertValues(source["startedAt"], null);
	        this.completedAt = this.convertValues(source["completedAt"], null);
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
	
	export class FileInfo {
	    path: string;
	    name: string;
	    extension: string;
	    size: number;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.extension = source["extension"];
	        this.size = source["size"];
	        this.type = source["type"];
	    }
	}
	export class UserSettings {
	    lastOutputDirectory: string;
	    defaultNamingMode: string;
	    defaultMakeCopies: boolean;
	    theme: string;
	
	    static createFrom(source: any = {}) {
	        return new UserSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lastOutputDirectory = source["lastOutputDirectory"];
	        this.defaultNamingMode = source["defaultNamingMode"];
	        this.defaultMakeCopies = source["defaultMakeCopies"];
	        this.theme = source["theme"];
	    }
	}

}

