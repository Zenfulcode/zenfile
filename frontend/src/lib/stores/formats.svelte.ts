/**
 * Format Store - Manages supported formats from the backend
 *
 * This store follows the Single Responsibility Principle by only handling
 * format-related state and operations.
 */

import { GetSupportedFormats } from '$lib/wailsjs/go/main/App';

interface SupportedFormats {
	videoFormats: string[];
	imageFormats: string[];
	backend: string;
	loaded: boolean;
	loading: boolean;
	error: string | null;
}

function createFormatStore() {
	let state = $state<SupportedFormats>({
		videoFormats: [],
		imageFormats: [],
		backend: '',
		loaded: false,
		loading: false,
		error: null
	});

	return {
		get videoFormats() {
			return state.videoFormats;
		},
		get imageFormats() {
			return state.imageFormats;
		},
		get backend() {
			return state.backend;
		},
		get loaded() {
			return state.loaded;
		},
		get loading() {
			return state.loading;
		},
		get error() {
			return state.error;
		},

		/**
		 * Load supported formats from the backend
		 * Should be called once when the app initializes
		 */
		async loadFormats() {
			if (state.loading || state.loaded) return;

			state.loading = true;
			state.error = null;

			try {
				const formats = await GetSupportedFormats();
				state.videoFormats = formats.videoFormats || [];
				state.imageFormats = formats.imageFormats || [];
				state.backend = formats.backend || 'unknown';
				state.loaded = true;
			} catch (err) {
				state.error = err instanceof Error ? err.message : 'Failed to load formats';
				console.error('Failed to load supported formats:', err);
				// Fallback to empty arrays - UI should handle this gracefully
				state.videoFormats = [];
				state.imageFormats = [];
			} finally {
				state.loading = false;
			}
		},

		/**
		 * Get formats for a specific file type
		 */
		getFormatsForType(fileType: 'video' | 'image' | 'unknown'): string[] {
			switch (fileType) {
				case 'video':
					return state.videoFormats;
				case 'image':
					return state.imageFormats;
				default:
					return [];
			}
		},

		/**
		 * Check if a format is supported for a file type
		 */
		isFormatSupported(fileType: 'video' | 'image', format: string): boolean {
			const formats = this.getFormatsForType(fileType);
			return formats.includes(format.toLowerCase());
		},

		/**
		 * Check if using AVFoundation (limited format support)
		 */
		isAVFoundation(): boolean {
			return state.backend === 'avfoundation';
		},

		/**
		 * Reset the store (useful for testing)
		 */
		reset() {
			state = {
				videoFormats: [],
				imageFormats: [],
				backend: '',
				loaded: false,
				loading: false,
				error: null
			};
		}
	};
}

export const formatStore = createFormatStore();
