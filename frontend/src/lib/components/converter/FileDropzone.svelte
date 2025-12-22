<script lang="ts">
	import { Upload, File, Video, Image, X } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { ScrollArea } from '$lib/components/ui/scroll-area';
	import { converterStore } from '$lib/stores/converter.svelte';
	import { formatStore } from '$lib/stores/formats.svelte';
	import { formatFileSize, type FileInfo } from '$lib/types';
	import { SelectFiles } from '$lib/wailsjs/go/main/App';

	let isDragOver = $state(false);

	// Get supported formats from backend
	let supportedFormatsText = $derived.by(() => {
		const videoFormats = formatStore.videoFormats.map((f) => f.toUpperCase());
		const imageFormats = formatStore.imageFormats.map((f) => f.toUpperCase());
		const allFormats = [...videoFormats, ...imageFormats];
		return allFormats.length > 0 ? allFormats.join(', ') : 'Loading...';
	});

	async function handleSelectFiles() {
		try {
			const files = await SelectFiles();
			if (files && files.length > 0) {
				converterStore.setFiles(files as FileInfo[]);
			}
		} catch (err) {
			console.error('Failed to select files:', err);
			converterStore.setError('Failed to select files');
		}
	}

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		isDragOver = true;
	}

	function handleDragLeave() {
		isDragOver = false;
	}

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		isDragOver = false;
		// Note: In Wails, drag and drop would need special handling
		// For now, we just show a message
		console.log('Drag and drop not supported in desktop mode, use the browse button');
	}

	function removeFile(index: number) {
		converterStore.removeFile(index);
	}

	function clearAll() {
		converterStore.clearFiles();
	}

	function getFileIcon(type: string) {
		switch (type) {
			case 'video':
				return Video;
			case 'image':
				return Image;
			default:
				return File;
		}
	}
</script>

<Card class="w-full">
	<CardContent class="p-6">
		{#if converterStore.files.length === 0}
			<!-- Empty state - Dropzone -->
			<button
				type="button"
				class="w-full rounded-lg border-2 border-dashed p-12 text-center transition-colors {isDragOver
					? 'border-primary bg-primary/5'
					: 'border-muted-foreground/25 hover:border-primary/50'}"
				ondragover={handleDragOver}
				ondragleave={handleDragLeave}
				ondrop={handleDrop}
				onclick={handleSelectFiles}
			>
				<Upload class="mx-auto mb-4 h-12 w-12 text-muted-foreground" />
				<h3 class="mb-2 text-lg font-semibold">Select files to convert</h3>
				<p class="mb-4 text-sm text-muted-foreground">
					Click to browse or drag and drop your files here
				</p>
				<p class="text-xs text-muted-foreground">
					Supported: {supportedFormatsText}
				</p>
			</button>
		{:else}
			<!-- File list -->
			<div class="space-y-4">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-2">
						<h3 class="font-semibold">Selected Files</h3>
						<Badge variant="secondary">{converterStore.files.length} files</Badge>
						<Badge variant="outline">{converterStore.fileType}</Badge>
					</div>
					<div class="flex gap-2">
						<Button variant="outline" size="sm" onclick={handleSelectFiles}>
							Add More
						</Button>
						<Button variant="ghost" size="sm" onclick={clearAll}>
							Clear All
						</Button>
					</div>
				</div>

				<ScrollArea class="h-[200px] rounded-md border">
					<div class="p-4 space-y-2">
						{#each converterStore.files as file, index}
							{@const FileIcon = getFileIcon(file.type)}
							<div
								class="flex items-center justify-between rounded-lg bg-muted/50 p-3 transition-colors hover:bg-muted"
							>
								<div class="flex items-center gap-3 overflow-hidden">
									<FileIcon class="h-5 w-5 shrink-0 text-muted-foreground" />
									<div class="min-w-0">
										<p class="truncate text-sm font-medium">{file.name}</p>
										<p class="text-xs text-muted-foreground">
											{formatFileSize(file.size)} • {file.extension}
										</p>
									</div>
								</div>
								<Button
									variant="ghost"
									size="icon"
									class="h-8 w-8 shrink-0"
									onclick={() => removeFile(index)}
								>
									<X class="h-4 w-4" />
								</Button>
							</div>
						{/each}
					</div>
				</ScrollArea>

				<div class="flex items-center justify-between text-sm text-muted-foreground">
					<span>Total size: {formatFileSize(converterStore.totalSize)}</span>
				</div>
			</div>
		{/if}
	</CardContent>
</Card>
