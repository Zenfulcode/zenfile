<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { ModeWatcher } from 'mode-watcher';
	import { Toaster } from '$lib/components/ui/sonner';

	let { children } = $props();
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
