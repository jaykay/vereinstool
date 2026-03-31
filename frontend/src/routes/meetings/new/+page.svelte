<script lang="ts">
	import { api, ApiError } from '$lib/api';
	import { goto } from '$app/navigation';

	let title = $state('');
	let scheduledAt = $state('');
	let location = $state('');
	let error = $state('');
	let loading = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;
		try {
			const meeting = await api.post<{ id: number }>('/meetings', {
				title,
				scheduled_at: new Date(scheduledAt).toISOString(),
				location: location || undefined
			});
			goto(`/meetings/${meeting.id}`);
		} catch (e) {
			if (e instanceof ApiError) {
				error = e.message;
			} else {
				error = 'Verbindungsfehler';
			}
		} finally {
			loading = false;
		}
	}
</script>

<div class="min-h-screen bg-gray-50">
	<header class="bg-white shadow-sm border-b border-gray-200">
		<div class="max-w-4xl mx-auto px-4 py-4">
			<a href="/" class="text-sm text-blue-600 hover:text-blue-800">&larr; Zurück</a>
			<h1 class="text-xl font-bold text-gray-900 mt-1">Neue Sitzung</h1>
		</div>
	</header>

	<main class="max-w-4xl mx-auto px-4 py-8">
		<form
			onsubmit={handleSubmit}
			class="bg-white shadow-sm rounded-lg p-6 space-y-4 border border-gray-200 max-w-lg"
		>
			{#if error}
				<div class="bg-red-50 text-red-700 text-sm px-3 py-2 rounded">{error}</div>
			{/if}

			<div>
				<label for="title" class="block text-sm font-medium text-gray-700 mb-1">Titel</label>
				<input
					id="title"
					type="text"
					bind:value={title}
					required
					placeholder="z.B. Vorstandssitzung April 2026"
					class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
				/>
			</div>

			<div>
				<label for="date" class="block text-sm font-medium text-gray-700 mb-1">Datum & Uhrzeit</label>
				<input
					id="date"
					type="datetime-local"
					bind:value={scheduledAt}
					required
					class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
				/>
			</div>

			<div>
				<label for="location" class="block text-sm font-medium text-gray-700 mb-1">
					Ort <span class="text-gray-400">(optional)</span>
				</label>
				<input
					id="location"
					type="text"
					bind:value={location}
					placeholder="z.B. Vereinsheim"
					class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
				/>
			</div>

			<div class="flex gap-3">
				<button
					type="submit"
					disabled={loading}
					class="px-4 py-2 bg-blue-600 text-white font-medium rounded-md hover:bg-blue-700 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
				>
					{loading ? 'Erstellen...' : 'Sitzung erstellen'}
				</button>
				<a href="/" class="px-4 py-2 text-gray-600 hover:text-gray-800 self-center">Abbrechen</a>
			</div>
		</form>
	</main>
</div>
