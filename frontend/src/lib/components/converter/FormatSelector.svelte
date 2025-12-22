<script lang="ts">
	import * as Select from '$lib/components/ui/select';
	import { Label } from '$lib/components/ui/label';
	import { converterStore } from '$lib/stores/converter.svelte';
	import { formatStore } from '$lib/stores/formats.svelte';

	// Get formats from the backend-aware format store
	let availableFormats = $derived(formatStore.getFormatsForType(converterStore.fileType));

	// Show a notice if using AVFoundation with limited formats
	let isLimitedBackend = $derived(formatStore.isAVFoundation() && converterStore.fileType === 'video');

	function handleFormatChange(value: string | undefined) {
		if (value) {
			converterStore.setOutputFormat(value);
		}
	}
</script>

<div class="space-y-2">
	<Label for="format-select">Output Format</Label>
	<Select.Root
		type="single"
		value={converterStore.outputFormat}
		onValueChange={handleFormatChange}
		disabled={!converterStore.hasFiles || !formatStore.loaded}
	>
		<Select.Trigger id="format-select" class="w-full">
			{#if formatStore.loading}
				<span class="text-muted-foreground">Loading formats...</span>
			{:else if converterStore.outputFormat}
				<span class="uppercase">{converterStore.outputFormat}</span>
			{:else}
				<span class="text-muted-foreground">Select output format</span>
			{/if}
		</Select.Trigger>
		<Select.Content>
			{#each availableFormats as format}
				<Select.Item value={format}>
					<span class="uppercase">{format}</span>
				</Select.Item>
			{/each}
		</Select.Content>
	</Select.Root>
	{#if converterStore.fileType}
		<p class="text-xs text-muted-foreground">
			Converting {converterStore.files.length}
			{converterStore.fileType} file{converterStore.files.length > 1 ? 's' : ''}
		</p>
	{/if}
	{#if isLimitedBackend}
		<p class="text-xs text-amber-600">
			Using native macOS conversion (limited formats)
		</p>
	{/if}
	{#if formatStore.error}
		<p class="text-xs text-red-500">
			{formatStore.error}
		</p>
	{/if}
</div>
