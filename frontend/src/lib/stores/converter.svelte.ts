import type {
	FileInfo,
	FileNamingMode,
	ConversionProgress,
	BatchConversionResult,
	FileType
} from '$lib/types';

// Converter store using Svelte 5 runes
class ConverterStore {
	// File selection state
	files = $state<FileInfo[]>([]);
	outputFormat = $state<string>('');
	outputDirectory = $state<string>('');

	// Options state
	namingMode = $state<FileNamingMode>('original');
	customNames = $state<string[]>([]);
	makeCopies = $state<boolean>(true);

	// Conversion state
	isConverting = $state<boolean>(false);
	currentProgress = $state<ConversionProgress | null>(null);
	overallProgress = $state<number>(0);
	lastResult = $state<BatchConversionResult | null>(null);

	// Error state
	error = $state<string | null>(null);

	// Computed properties
	get hasFiles(): boolean {
		return this.files.length > 0;
	}

	get fileType(): FileType {
		if (this.files.length === 0) return 'unknown';
		return this.files[0].type as FileType;
	}

	get canConvert(): boolean {
		return (
			this.files.length > 0 &&
			this.outputFormat !== '' &&
			this.outputDirectory !== '' &&
			!this.isConverting
		);
	}

	get totalSize(): number {
		return this.files.reduce((sum, file) => sum + file.size, 0);
	}

	// Actions
	setFiles(newFiles: FileInfo[]) {
		this.files = newFiles;
		this.outputFormat = '';
		this.customNames = newFiles.map(() => '');
		this.error = null;
	}

	addFiles(newFiles: FileInfo[]) {
		// Only add files of the same type
		if (this.files.length > 0 && newFiles.length > 0) {
			const existingType = this.files[0].type;
			newFiles = newFiles.filter((f) => f.type === existingType);
		}
		this.files = [...this.files, ...newFiles];
		this.customNames = [...this.customNames, ...newFiles.map(() => '')];
	}

	removeFile(index: number) {
		this.files = this.files.filter((_, i) => i !== index);
		this.customNames = this.customNames.filter((_, i) => i !== index);
	}

	clearFiles() {
		this.files = [];
		this.outputFormat = '';
		this.customNames = [];
		this.error = null;
		this.lastResult = null;
	}

	setOutputFormat(format: string) {
		this.outputFormat = format;
	}

	setOutputDirectory(directory: string) {
		this.outputDirectory = directory;
	}

	setNamingMode(mode: FileNamingMode) {
		this.namingMode = mode;
	}

	setCustomName(index: number, name: string) {
		if (index >= 0 && index < this.customNames.length) {
			this.customNames[index] = name;
		}
	}

	setMakeCopies(value: boolean) {
		this.makeCopies = value;
	}

	startConversion() {
		this.isConverting = true;
		this.overallProgress = 0;
		this.error = null;
	}

	updateProgress(progress: ConversionProgress) {
		this.currentProgress = progress;
	}

	setOverallProgress(progress: number) {
		this.overallProgress = progress;
	}

	finishConversion(result: BatchConversionResult) {
		this.isConverting = false;
		this.lastResult = result;
		this.overallProgress = 100;
		this.currentProgress = null;
	}

	setError(message: string) {
		this.error = message;
		this.isConverting = false;
	}

	clearError() {
		this.error = null;
	}
}

export const converterStore = new ConverterStore();
