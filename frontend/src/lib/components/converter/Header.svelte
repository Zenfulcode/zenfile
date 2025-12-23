<script lang="ts">
	import { Sun, Moon, Monitor, Settings, Info } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Badge } from '$lib/components/ui/badge';
	import { settingsStore } from '$lib/stores/settings.svelte';
	import { formatStore } from '$lib/stores/formats.svelte';
	import { toggleMode, setMode, mode } from 'mode-watcher';

	let backendName = $derived(formatStore.isAVFoundation() ? 'AVFoundation' : 'FFmpeg');

	let showAbout = $state(false);

	function handleThemeChange(theme: 'light' | 'dark' | 'system') {
		if (theme === 'system') {
			setMode('system');
		} else {
			setMode(theme);
		}
		settingsStore.updateSetting('theme', theme);
	}
</script>

<header class="border-b bg-card pt-4">
	<div class="container flex h-14 items-center justify-between px-4">
		<div class="flex items-center gap-3">
			<h1 class="text-xl font-bold">Zenfile</h1>
			{#if settingsStore.appInfo}
				<Badge variant="outline" class="text-xs">v{settingsStore.appInfo.version}</Badge>
			{/if}
		</div>

		<div class="flex items-center gap-2">
			<!-- Converter Backend Status -->
			{#if settingsStore.ffmpegAvailable}
				<Badge variant="outline" class="bg-green-500/10 text-green-600 border-green-500/30">
					{backendName} Ready
				</Badge>
			{:else}
				<Badge variant="destructive">{backendName} Not Found</Badge>
			{/if}

			<!-- Theme Toggle -->
			<DropdownMenu.Root>
				<DropdownMenu.Trigger>
					{#snippet child({ props })}
						<Button variant="ghost" size="icon" {...props}>
							<Sun class="h-5 w-5 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
							<Moon
								class="absolute h-5 w-5 rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100"
							/>
							<span class="sr-only">Toggle theme</span>
						</Button>
					{/snippet}
				</DropdownMenu.Trigger>
				<DropdownMenu.Content align="end">
					<DropdownMenu.Item onclick={() => handleThemeChange('light')}>
						<Sun class="mr-2 h-4 w-4" />
						Light
					</DropdownMenu.Item>
					<DropdownMenu.Item onclick={() => handleThemeChange('dark')}>
						<Moon class="mr-2 h-4 w-4" />
						Dark
					</DropdownMenu.Item>
					<DropdownMenu.Item onclick={() => handleThemeChange('system')}>
						<Monitor class="mr-2 h-4 w-4" />
						System
					</DropdownMenu.Item>
				</DropdownMenu.Content>
			</DropdownMenu.Root>

			<!-- About -->
			<Button variant="ghost" size="icon" onclick={() => (showAbout = true)}>
				<Info class="h-5 w-5" />
			</Button>
		</div>
	</div>
</header>

<!-- About Dialog -->
<Dialog.Root bind:open={showAbout}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>About Zenfile</Dialog.Title>
			<Dialog.Description>A cross-platform file converter for video and image files.</Dialog.Description>
		</Dialog.Header>
		<div class="space-y-4 py-4">
			{#if settingsStore.appInfo}
				<div class="grid grid-cols-2 gap-2 text-sm">
					<span class="text-muted-foreground">Version:</span>
					<span>{settingsStore.appInfo.version}</span>

					<span class="text-muted-foreground">Data Directory:</span>
					<span class="truncate text-xs">{settingsStore.appInfo.dataDir}</span>

					<span class="text-muted-foreground">Log File:</span>
					<span class="truncate text-xs">{settingsStore.appInfo.logFile}</span>

					{#if settingsStore.appInfo.ffmpegVersion}
						<span class="text-muted-foreground">FFmpeg:</span>
						<span class="truncate text-xs">{settingsStore.appInfo.ffmpegVersion}</span>
					{/if}
				</div>
			{/if}

			<div class="pt-4 border-t">
				<p class="text-sm text-muted-foreground">
					Supports conversion between video formats (MP4, WebM, AVI, MKV, MOV, GIF) and image
					formats (PNG, JPG, JPEG, GIF, BMP, TIFF).
				</p>
			</div>
		</div>
	</Dialog.Content>
</Dialog.Root>
