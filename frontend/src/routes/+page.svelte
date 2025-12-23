<script lang="ts">
	import { onMount } from 'svelte';
	import { Play, AlertCircle } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Alert, AlertDescription, AlertTitle } from '$lib/components/ui/alert';
	import { Separator } from '$lib/components/ui/separator';
	import {
		Header,
		FileDropzone,
		FormatSelector,
		OutputOptions,
		ConversionProgress
	} from '$lib/components/converter';
	import { converterStore } from '$lib/stores/converter.svelte';
	import { settingsStore } from '$lib/stores/settings.svelte';
	import { formatStore } from '$lib/stores/formats.svelte';
	import { GetSettings, GetAppInfo, CheckFFmpeg, ConvertFiles } from '$lib/wailsjs/go/main/App';
	import { EventsOn } from '$lib/wailsjs/runtime/runtime';
	import type {
		BatchConversionRequest,
		ConversionProgress as ConversionProgressType,
		FileNamingMode,
		UserSettings
	} from '$lib/types';

	onMount(async () => {
		try {
			// Load app info
			const appInfo = await GetAppInfo();
			settingsStore.setAppInfo(appInfo);

			// Check FFmpeg
			const ffmpegAvailable = await CheckFFmpeg();
			settingsStore.setFFmpegAvailable(ffmpegAvailable);

			// Load settings
			const settings = await GetSettings();
			if (settings) {
				settingsStore.setSettings(settings as UserSettings);
				if (settings.lastOutputDirectory) {
					converterStore.setOutputDirectory(settings.lastOutputDirectory);
				}
				converterStore.setNamingMode(settings.defaultNamingMode as FileNamingMode);
				converterStore.setMakeCopies(settings.defaultMakeCopies);
			}

			// Listen for conversion progress events
			EventsOn('conversion:progress', (progress: ConversionProgressType) => {
				converterStore.updateProgress(progress);
				converterStore.setOverallProgress(progress.progress);
			});

			// Load supported formats from backend
			await formatStore.loadFormats();

			settingsStore.setLoading(false);
		} catch (err) {
			console.error('Failed to initialize:', err);
			toast.error('Failed to initialize application');
			settingsStore.setLoading(false);
		}
	});

	async function handleConvert() {
		if (!converterStore.canConvert) return;

		converterStore.startConversion();
		toast.info('Starting conversion...');

		try {
			const request: BatchConversionRequest = {
				files: converterStore.files.map((f) => f.path),
				outputFormat: converterStore.outputFormat,
				outputDirectory: converterStore.outputDirectory,
				namingMode: converterStore.namingMode,
				customNames: converterStore.customNames.filter((n) => n !== ''),
				makeCopies: converterStore.makeCopies
			};

			const result = await ConvertFiles(request);

			converterStore.finishConversion(result);

			if (result.failCount === 0) {
				toast.success(`Successfully converted ${result.successCount} file(s)`);
			} else if (result.successCount > 0) {
				toast.warning(`Converted ${result.successCount} file(s), ${result.failCount} failed`);
			} else {
				toast.error('All conversions failed');
			}
		} catch (err) {
			console.error('Conversion error:', err);
			converterStore.setError(String(err));
			toast.error('Conversion failed: ' + String(err));
		}
	}
</script>

<Header />

<main class="container mx-auto max-w-4xl px-4 py-6">
	{#if !settingsStore.ffmpegAvailable && !settingsStore.isLoading}
		<Alert variant="destructive" class="mb-6">
			<AlertCircle class="h-4 w-4" />
			<AlertTitle>FFmpeg Not Found</AlertTitle>
			<AlertDescription>
				FFmpeg is required for video conversion. Please install FFmpeg and restart the application.
				Image conversion will still work without FFmpeg.
			</AlertDescription>
		</Alert>
	{/if}

	{#if converterStore.error}
		<Alert variant="destructive" class="mb-6">
			<AlertCircle class="h-4 w-4" />
			<AlertTitle>Error</AlertTitle>
			<AlertDescription>{converterStore.error}</AlertDescription>
		</Alert>
	{/if}

	<div class="grid gap-6 lg:grid-cols-2">
		<!-- Left Column - File Selection -->
		<div class="space-y-6">
			<FileDropzone />

			{#if converterStore.isConverting || converterStore.lastResult}
				<ConversionProgress />
			{/if}
		</div>

		<!-- Right Column - Options -->
		<div class="space-y-6">
			<Card>
				<CardHeader>
					<CardTitle>Conversion Options</CardTitle>
				</CardHeader>
				<CardContent class="space-y-6">
					<FormatSelector />

					<Separator />

					<OutputOptions />

					<Separator />

					<Button
						class="w-full"
						size="lg"
						disabled={!converterStore.canConvert}
						onclick={handleConvert}
					>
						<Play class="mr-2 h-5 w-5" />
						{#if converterStore.isConverting}
							Converting...
						{:else}
							Convert {converterStore.files.length || ''} File{converterStore.files.length !== 1
								? 's'
								: ''}
						{/if}
					</Button>
				</CardContent>
			</Card>
		</div>
	</div>
</main>
