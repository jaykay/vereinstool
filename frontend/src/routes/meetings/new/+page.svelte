<script lang="ts">
	import { api, ApiError } from '$lib/api';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	interface PoolTopic {
		id: number;
		title: string;
		category: string | null;
		vote_count: number;
		estimated_mins: number;
	}

	// Step 1: create meeting
	let title = $state('');
	let scheduledDate = $state('');
	let scheduledTime = $state('19:00');
	let durationMins = $state(90);
	let location = $state('');
	let error = $state('');
	let loading = $state(false);

	// Step 2: assign pool topics
	let step = $state<1 | 2>(1);
	let meetingId = $state(0);
	let poolTopics = $state<PoolTopic[]>([]);
	let assigning = $state<Set<number>>(new Set());

	const timeOptions: string[] = [];
	for (let h = 0; h < 24; h++) {
		for (const m of ['00', '30']) {
			timeOptions.push(`${String(h).padStart(2, '0')}:${m}`);
		}
	}

	onMount(async () => {
		try { poolTopics = await api.get<PoolTopic[]>('/topics/pool'); } catch {}
	});

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;
		try {
			const scheduledAt = new Date(`${scheduledDate}T${scheduledTime}`).toISOString();
			const meeting = await api.post<{ id: number }>('/meetings', {
				title, scheduled_at: scheduledAt, duration_mins: durationMins,
				location: location || undefined
			});
			meetingId = meeting.id;
			if (poolTopics.length > 0) {
				step = 2;
			} else {
				goto(`/meetings/${meeting.id}`);
			}
		} catch (e) {
			if (e instanceof ApiError) error = e.message;
			else error = 'Verbindungsfehler';
		} finally {
			loading = false;
		}
	}

	async function assignTopic(topicId: number) {
		assigning = new Set([...assigning, topicId]);
		try {
			await api.post(`/topics/${topicId}/assign`, { meeting_id: meetingId });
			poolTopics = poolTopics.filter((t) => t.id !== topicId);
		} catch (e) { if (e instanceof ApiError) error = e.message; }
		finally {
			const next = new Set(assigning);
			next.delete(topicId);
			assigning = next;
		}
	}

	function categoryLabel(c: string | null) { if (!c) return null; return { finanzen: 'Finanzen', satzung: 'Satzung', veranstaltungen: 'Veranstaltungen', sonstiges: 'Sonstiges' }[c] ?? c; }
	function categoryColor(c: string | null) { if (!c) return ''; return { finanzen: 'bg-yellow-100 text-yellow-800', satzung: 'bg-purple-100 text-purple-800', veranstaltungen: 'bg-pink-100 text-pink-800', sonstiges: 'bg-gray-100 text-gray-600' }[c] ?? 'bg-gray-100 text-gray-600'; }
</script>

<div class="min-h-screen bg-gray-50">
	<header class="bg-white shadow-sm border-b border-gray-200">
		<div class="max-w-4xl mx-auto px-4 py-4">
			<a href="/" class="text-sm text-blue-600 hover:text-blue-800">&larr; Zurück</a>
			<h1 class="text-xl font-bold text-gray-900 mt-1">Neue Sitzung</h1>
		</div>
	</header>

	<main class="max-w-4xl mx-auto px-4 py-8">
		{#if error}
			<div class="bg-red-50 text-red-700 text-sm px-3 py-2 rounded mb-4">{error}</div>
		{/if}

		{#if step === 1}
			<form onsubmit={handleSubmit} class="bg-white shadow-sm rounded-lg p-6 space-y-4 border border-gray-200 max-w-lg">
				<div>
					<label for="title" class="block text-sm font-medium text-gray-700 mb-1">Titel</label>
					<input id="title" type="text" bind:value={title} required placeholder="z.B. Vorstandssitzung April 2026" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent" />
				</div>

				<div class="flex gap-3">
					<div class="flex-1">
						<label for="date" class="block text-sm font-medium text-gray-700 mb-1">Datum</label>
						<input id="date" type="date" bind:value={scheduledDate} required class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent" />
					</div>
					<div class="w-28">
						<label for="time" class="block text-sm font-medium text-gray-700 mb-1">Uhrzeit</label>
						<select id="time" bind:value={scheduledTime} class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500">
							{#each timeOptions as t}<option value={t}>{t}</option>{/each}
						</select>
					</div>
					<div class="w-28">
						<label for="duration" class="block text-sm font-medium text-gray-700 mb-1">Dauer</label>
						<select id="duration" bind:value={durationMins} class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500">
							<option value={30}>30 Min.</option>
							<option value={60}>1 Std.</option>
							<option value={90}>1,5 Std.</option>
							<option value={120}>2 Std.</option>
							<option value={150}>2,5 Std.</option>
							<option value={180}>3 Std.</option>
						</select>
					</div>
				</div>

				<div>
					<label for="location" class="block text-sm font-medium text-gray-700 mb-1">Ort <span class="text-gray-400">(optional)</span></label>
					<input id="location" type="text" bind:value={location} placeholder="z.B. Vereinsheim" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent" />
				</div>

				<div class="flex gap-3">
					<button type="submit" disabled={loading} class="px-4 py-2 bg-blue-600 text-white font-medium rounded-md hover:bg-blue-700 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed">
						{loading ? 'Erstellen...' : poolTopics.length > 0 ? 'Weiter: Themen zuweisen' : 'Sitzung erstellen'}
					</button>
					<a href="/" class="px-4 py-2 text-gray-600 hover:text-gray-800 self-center">Abbrechen</a>
				</div>
			</form>
		{:else}
			<!-- Step 2: Assign pool topics -->
			<div class="max-w-lg">
				<div class="bg-white shadow-sm rounded-lg p-6 border border-gray-200">
					<h2 class="text-lg font-semibold text-gray-900 mb-2">Themen aus dem Pool übernehmen</h2>
					<p class="text-sm text-gray-500 mb-4">Sortiert nach Votes. Klicke auf "Übernehmen" um ein Thema in diese Sitzung aufzunehmen.</p>

					{#if poolTopics.length === 0}
						<p class="text-gray-500 text-sm mb-4">Alle Themen wurden übernommen oder der Pool ist leer.</p>
					{:else}
						<div class="space-y-2 mb-4">
							{#each poolTopics as topic}
								<div class="flex items-center justify-between bg-gray-50 rounded-md px-3 py-2">
									<div class="flex items-center gap-2">
										<span class="text-sm font-semibold text-gray-500 w-6 text-center">{topic.vote_count}</span>
										<span class="text-sm text-gray-900">{topic.title}</span>
										{#if topic.category}<span class="text-xs px-1.5 py-0.5 rounded-full {categoryColor(topic.category)}">{categoryLabel(topic.category)}</span>{/if}
										<span class="text-xs text-gray-400">{topic.estimated_mins} Min.</span>
									</div>
									<button onclick={() => assignTopic(topic.id)} disabled={assigning.has(topic.id)} class="text-xs px-2 py-1 bg-blue-600 text-white rounded hover:bg-blue-700 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed">
										{assigning.has(topic.id) ? '...' : 'Übernehmen'}
									</button>
								</div>
							{/each}
						</div>
					{/if}

					<a href="/meetings/{meetingId}" class="inline-block px-4 py-2 bg-blue-600 text-white font-medium rounded-md hover:bg-blue-700">
						Zur Sitzung &rarr;
					</a>
				</div>
			</div>
		{/if}
	</main>
</div>
