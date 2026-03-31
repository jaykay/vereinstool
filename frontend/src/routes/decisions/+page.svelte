<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';

	interface Decision {
		id: number;
		topic_id: number;
		meeting_id: number;
		text: string;
		votes_yes?: number;
		votes_no?: number;
		votes_abstain?: number;
		recorded_by: number;
		created_at: string;
	}

	let decisions = $state<Decision[]>([]);
	let loading = $state(true);
	let search = $state('');

	onMount(async () => {
		try {
			decisions = await api.get<Decision[]>('/decisions');
		} finally {
			loading = false;
		}
	});

	const filtered = $derived(
		search
			? decisions.filter((d) => d.text.toLowerCase().includes(search.toLowerCase()))
			: decisions
	);
</script>

<div class="min-h-screen bg-gray-50">
	<header class="bg-white shadow-sm border-b border-gray-200">
		<div class="max-w-4xl mx-auto px-4 py-4 flex items-center justify-between">
			<h1 class="text-xl font-bold text-gray-900">Beschluss-Register</h1>
			<a href="/" class="text-sm text-blue-600 hover:text-blue-800">&larr; Dashboard</a>
		</div>
	</header>

	<main class="max-w-4xl mx-auto px-4 py-8">
		<div class="mb-6">
			<input
				type="text"
				bind:value={search}
				placeholder="Beschlüsse durchsuchen..."
				class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
			/>
		</div>

		{#if loading}
			<p class="text-gray-400">Laden...</p>
		{:else if filtered.length === 0}
			<p class="text-gray-500">{search ? 'Keine Treffer.' : 'Noch keine Beschlüsse vorhanden.'}</p>
		{:else}
			<div class="space-y-3">
				{#each filtered as dec}
					<div class="bg-white border border-gray-200 rounded-lg p-4">
						<div class="flex items-start justify-between">
							<div>
								<p class="text-gray-900">{dec.text}</p>
								<div class="flex gap-3 mt-2 text-xs text-gray-500">
									<a href="/meetings/{dec.meeting_id}" class="text-blue-600 hover:text-blue-800">Sitzung #{dec.meeting_id}</a>
									<span>{new Date(dec.created_at).toLocaleDateString('de-DE')}</span>
								</div>
							</div>
						</div>
						{#if dec.votes_yes !== undefined || dec.votes_no !== undefined}
							<div class="flex gap-3 mt-2 text-xs">
								{#if dec.votes_yes !== undefined}<span class="text-green-600">Ja: {dec.votes_yes}</span>{/if}
								{#if dec.votes_no !== undefined}<span class="text-red-600">Nein: {dec.votes_no}</span>{/if}
								{#if dec.votes_abstain !== undefined}<span class="text-gray-500">Enthaltung: {dec.votes_abstain}</span>{/if}
							</div>
						{/if}
					</div>
				{/each}
			</div>
		{/if}
	</main>
</div>
