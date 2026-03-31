<script lang="ts">
	import { user } from '$lib/stores';
	import { api, ApiError } from '$lib/api';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	interface Meeting {
		id: number;
		title: string;
		scheduled_at: string;
		location: string | null;
		status: string;
	}

	interface Topic {
		id: number;
		title: string;
		description: string | null;
		category: string | null;
		vote_count: number;
		voted: boolean;
		estimated_mins: number;
		submitted_by: number;
	}

	let meetings = $state<Meeting[]>([]);
	let poolTopics = $state<Topic[]>([]);
	let loading = $state(true);

	// Pool topic form
	let showPoolForm = $state(false);
	let poolTitle = $state('');
	let poolDescription = $state('');
	let poolCategory = $state('');
	let poolMins = $state(10);
	let poolLoading = $state(false);
	let error = $state('');

	const categories = ['finanzen', 'satzung', 'veranstaltungen', 'sonstiges'];

	onMount(async () => {
		try {
			const [m, t] = await Promise.all([
				api.get<Meeting[]>('/meetings'),
				api.get<Topic[]>('/topics/pool')
			]);
			meetings = m;
			poolTopics = t;
		} finally {
			loading = false;
		}
	});

	async function handleLogout() {
		await user.logout();
		goto('/login');
	}

	async function submitPoolTopic(e: Event) {
		e.preventDefault();
		poolLoading = true;
		error = '';
		try {
			await api.post('/topics', {
				title: poolTitle,
				description: poolDescription || undefined,
				category: poolCategory || undefined,
				estimated_mins: poolMins
			});
			poolTitle = ''; poolDescription = ''; poolCategory = ''; poolMins = 10; showPoolForm = false;
			poolTopics = await api.get<Topic[]>('/topics/pool');
		} catch (e) { if (e instanceof ApiError) error = e.message; }
		finally { poolLoading = false; }
	}

	async function vote(topicId: number) {
		try {
			const updated = await api.post<Topic>(`/topics/${topicId}/vote`);
			poolTopics = poolTopics.map((t) => (t.id === topicId ? updated : t));
		} catch {}
	}

	async function unvote(topicId: number) {
		try {
			const updated = await api.delete<Topic>(`/topics/${topicId}/vote`);
			poolTopics = poolTopics.map((t) => (t.id === topicId ? updated : t));
		} catch {}
	}

	function statusLabel(status: string) {
		return { open: 'Geplant', active: 'Aktiv', closed: 'Abgeschlossen' }[status] ?? status;
	}

	function statusColor(status: string) {
		return {
			open: 'bg-blue-100 text-blue-800',
			active: 'bg-green-100 text-green-800',
			closed: 'bg-gray-100 text-gray-500'
		}[status] ?? 'bg-gray-100 text-gray-500';
	}

	function categoryLabel(c: string | null) { if (!c) return null; return { finanzen: 'Finanzen', satzung: 'Satzung', veranstaltungen: 'Veranstaltungen', sonstiges: 'Sonstiges' }[c] ?? c; }
	function categoryColor(c: string | null) { if (!c) return ''; return { finanzen: 'bg-yellow-100 text-yellow-800', satzung: 'bg-purple-100 text-purple-800', veranstaltungen: 'bg-pink-100 text-pink-800', sonstiges: 'bg-gray-100 text-gray-600' }[c] ?? 'bg-gray-100 text-gray-600'; }

	function formatDate(iso: string) {
		return new Date(iso).toLocaleDateString('de-DE', {
			weekday: 'short', day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit'
		});
	}

	const openMeetings = $derived(meetings.filter((m) => m.status === 'open'));
	const activeMeetings = $derived(meetings.filter((m) => m.status === 'active'));
	const closedMeetings = $derived(meetings.filter((m) => m.status === 'closed'));
</script>

{#if $user}
	<div class="min-h-screen bg-gray-50">
		<header class="bg-white shadow-sm border-b border-gray-200">
			<div class="max-w-4xl mx-auto px-4 py-4 flex items-center justify-between">
				<div class="flex items-center gap-6">
					<h1 class="text-xl font-bold text-gray-900">Vereinstool</h1>
					<nav class="flex gap-4 text-sm">
						<a href="/decisions" class="text-gray-600 hover:text-gray-900">Beschlüsse</a>
						<a href="/tasks" class="text-gray-600 hover:text-gray-900">Aufgaben</a>
					</nav>
				</div>
				<div class="flex items-center gap-4">
					<span class="text-sm text-gray-600">{$user.name}</span>
					<button onclick={handleLogout} class="text-sm text-gray-500 hover:text-gray-700 cursor-pointer">Abmelden</button>
				</div>
			</div>
		</header>

		<main class="max-w-4xl mx-auto px-4 py-8">
			{#if error}
				<div class="bg-red-50 text-red-700 text-sm px-3 py-2 rounded mb-4">{error}</div>
			{/if}

			<!-- Meetings -->
			<div class="flex items-center justify-between mb-6">
				<h2 class="text-lg font-semibold text-gray-900">Sitzungen</h2>
				{#if $user.role === 'admin'}
					<a href="/meetings/new" class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700">Neue Sitzung</a>
				{/if}
			</div>

			{#if loading}
				<p class="text-gray-400">Laden...</p>
			{:else}
				{#if meetings.length === 0}
					<p class="text-gray-500 mb-8">Noch keine Sitzungen vorhanden.</p>
				{:else}
					{#if activeMeetings.length > 0}
						<section class="mb-8">
							<h3 class="text-sm font-medium text-gray-500 uppercase tracking-wide mb-3">Aktiv</h3>
							{#each activeMeetings as meeting}{@render meetingCard(meeting)}{/each}
						</section>
					{/if}
					{#if openMeetings.length > 0}
						<section class="mb-8">
							<h3 class="text-sm font-medium text-gray-500 uppercase tracking-wide mb-3">Geplant</h3>
							{#each openMeetings as meeting}{@render meetingCard(meeting)}{/each}
						</section>
					{/if}
					{#if closedMeetings.length > 0}
						<section class="mb-8">
							<h3 class="text-sm font-medium text-gray-500 uppercase tracking-wide mb-3">Abgeschlossen</h3>
							{#each closedMeetings as meeting}{@render meetingCard(meeting)}{/each}
						</section>
					{/if}
				{/if}

				<!-- Topic Pool -->
				<div class="mt-4">
					<div class="flex items-center justify-between mb-4">
						<h2 class="text-lg font-semibold text-gray-900">Themenpool</h2>
						<button onclick={() => (showPoolForm = !showPoolForm)} class="px-3 py-1.5 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 cursor-pointer">{showPoolForm ? 'Abbrechen' : 'Thema vorschlagen'}</button>
					</div>
					<p class="text-sm text-gray-500 mb-4">Themen hier sammeln und per Upvote priorisieren. Beim Erstellen einer Sitzung werden sie zur Übernahme vorgeschlagen.</p>

					{#if showPoolForm}
						<form onsubmit={submitPoolTopic} class="bg-white border border-gray-200 rounded-lg p-4 mb-4 space-y-3">
							<input type="text" bind:value={poolTitle} required placeholder="Thema" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent" />
							<textarea bind:value={poolDescription} placeholder="Beschreibung (optional)" rows={2} class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"></textarea>
							<div class="flex gap-3">
								<select bind:value={poolCategory} class="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500">
									<option value="">Kategorie...</option>
									{#each categories as cat}<option value={cat}>{categoryLabel(cat)}</option>{/each}
								</select>
								<div class="flex items-center gap-2">
									<input type="number" bind:value={poolMins} min={1} max={120} class="w-20 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500" />
									<span class="text-sm text-gray-500">Min.</span>
								</div>
								<button type="submit" disabled={poolLoading} class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed">{poolLoading ? 'Speichern...' : 'Einreichen'}</button>
							</div>
						</form>
					{/if}

					{#if poolTopics.length === 0}
						<p class="text-gray-500 text-sm">Keine offenen Themen im Pool.</p>
					{:else}
						<div class="bg-white border border-gray-200 rounded-lg divide-y divide-gray-100">
							{#each poolTopics as topic}
								<div class="p-3 flex items-center gap-3">
									<button onclick={() => topic.voted ? unvote(topic.id) : vote(topic.id)} class="flex flex-col items-center cursor-pointer group min-w-[2.5rem]">
										<svg class="w-4 h-4 {topic.voted ? 'text-blue-600' : 'text-gray-300 group-hover:text-blue-400'} transition-colors" fill={topic.voted ? 'currentColor' : 'none'} stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" /></svg>
										<span class="text-sm font-semibold {topic.voted ? 'text-blue-600' : 'text-gray-500'}">{topic.vote_count}</span>
									</button>
									<div class="flex-1 min-w-0">
										<div class="flex items-center gap-2">
											<span class="font-medium text-gray-900">{topic.title}</span>
											{#if topic.category}<span class="text-xs px-1.5 py-0.5 rounded-full {categoryColor(topic.category)}">{categoryLabel(topic.category)}</span>{/if}
											<span class="text-xs text-gray-400">{topic.estimated_mins} Min.</span>
										</div>
										{#if topic.description}<p class="text-sm text-gray-500 mt-0.5">{topic.description}</p>{/if}
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			{/if}
		</main>
	</div>
{:else}
	<div class="min-h-screen flex items-center justify-center bg-gray-50">
		<p class="text-gray-400">Laden...</p>
	</div>
{/if}

{#snippet meetingCard(meeting: Meeting)}
	<a href="/meetings/{meeting.id}" class="block bg-white border border-gray-200 rounded-lg p-4 mb-2 hover:border-gray-300 transition-colors">
		<div class="flex items-center justify-between">
			<div>
				<span class="font-medium text-gray-900">{meeting.title}</span>
				<div class="text-sm text-gray-500 mt-1">
					{formatDate(meeting.scheduled_at)}
					{#if meeting.location}<span class="ml-2">&middot; {meeting.location}</span>{/if}
				</div>
			</div>
			<span class="text-xs font-medium px-2 py-1 rounded-full {statusColor(meeting.status)}">{statusLabel(meeting.status)}</span>
		</div>
	</a>
{/snippet}
