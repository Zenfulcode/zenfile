<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { ModeWatcher } from 'mode-watcher';
	import { Toaster } from '$lib/components/ui/sonner';
	import { formatStore } from '$lib/stores/formats.svelte';
	import { onMount } from 'svelte';

	let { children } = $props();

	// Load supported formats from the backend on app startup
	onMount(() => {
		formatStore.loadFormats();
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>Zenfile</title>
</svelte:head>

<ModeWatcher />
<Toaster richColors position="top-right" />

<div class="min-h-screen bg-background">
	{@render children()}
</div>
