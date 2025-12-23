// File types matching Go backend models

export type FileType = 'video' | 'image' | 'unknown';

export interface FileInfo {
	path: string;
	name: string;
	extension: string;
	size: number;
	type: FileType;
}

export type ConversionStatus = 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled';

export type FileNamingMode = 'original' | 'custom';

export interface ConversionProgress {
	id: number;
	inputPath: string;
	progress: number;
	status: string;
}

export interface ConversionResult {
	success: boolean;
	inputPath: string;
	outputPath: string;
	outputSize: number;
	errorMessage?: string;
	duration: number;
}

export interface BatchConversionRequest {
	files: string[];
	outputFormat: string;
	outputDirectory: string;
	namingMode: FileNamingMode;
	customNames?: string[];
	makeCopies: boolean;
}

export interface BatchConversionResult {
	totalFiles: number;
	successCount: number;
	failCount: number;
	results: ConversionResult[];
	totalDuration: number;
}

export interface Conversion {
	ID: number;
	CreatedAt: string;
	UpdatedAt: string;
	inputPath: string;
	outputPath: string;
	inputFormat: string;
	outputFormat: string;
	fileType: FileType;
	fileSize: number;
	outputSize: number;
	status: ConversionStatus;
	errorMessage?: string;
	progress: number;
	startedAt?: string;
	completedAt?: string;
}

export interface UserSettings {
	lastOutputDirectory: string;
	defaultNamingMode: FileNamingMode;
	defaultMakeCopies: boolean;
	theme: string;
}

export interface AppInfo {
	name: string;
	version: string;
	dataDir: string;
	logFile: string;
	converterBackend: string;
	ffmpegVersion?: string;
}

// Format options
export const VIDEO_OUTPUT_FORMATS = ['mp4', 'webm', 'avi', 'mkv', 'mov', 'gif'] as const;
export const IMAGE_OUTPUT_FORMATS = ['png', 'jpg', 'jpeg', 'gif', 'bmp', 'tiff'] as const;

export type VideoOutputFormat = (typeof VIDEO_OUTPUT_FORMATS)[number];
export type ImageOutputFormat = (typeof IMAGE_OUTPUT_FORMATS)[number];

// Helper to format file size
export function formatFileSize(bytes: number): string {
	if (bytes === 0) return '0 B';
	const k = 1024;
	const sizes = ['B', 'KB', 'MB', 'GB'];
	const i = Math.floor(Math.log(bytes) / Math.log(k));
	return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

// Helper to format duration
export function formatDuration(ms: number): string {
	if (ms < 1000) return `${ms}ms`;
	if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
	const minutes = Math.floor(ms / 60000);
	const seconds = Math.floor((ms % 60000) / 1000);
	return `${minutes}m ${seconds}s`;
}
