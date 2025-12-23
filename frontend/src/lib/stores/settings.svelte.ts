import type { UserSettings } from '$lib/types';
import type { main } from '$lib/wailsjs/go/models';

// Settings store using Svelte 5 runes
class SettingsStore {
	settings = $state<UserSettings>({
		lastOutputDirectory: '',
		defaultNamingMode: 'original',
		defaultMakeCopies: true,
		theme: 'system'
	});

	appInfo = $state<main.AppInfoResponse | null>(null);
	ffmpegAvailable = $state<boolean>(false);
	isLoading = $state<boolean>(true);

	setSettings(newSettings: UserSettings) {
		this.settings = newSettings;
	}

	setAppInfo(info: main.AppInfoResponse) {
		this.appInfo = info;
	}

	setFFmpegAvailable(available: boolean) {
		this.ffmpegAvailable = available;
	}

	setLoading(loading: boolean) {
		this.isLoading = loading;
	}

	updateSetting<K extends keyof UserSettings>(key: K, value: UserSettings[K]) {
		this.settings = { ...this.settings, [key]: value };
	}
}

export const settingsStore = new SettingsStore();
