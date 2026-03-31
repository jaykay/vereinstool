<script lang="ts">
	import { api, ApiError } from '$lib/api';
	import { user } from '$lib/stores';
	import { page } from '$app/state';
	import { onMount } from 'svelte';

	interface Attendee {
		id: number;
		email: string;
		name: string;
		role: string;
		present: boolean;
	}

	interface Meeting {
		id: number;
		title: string;
		scheduled_at: string;
		location: string | null;
		status: string;
		created_by: number;
	}

	interface Topic {
		id: number;
		meeting_id: number;
		title: string;
		description: string | null;
		category: string | null;
		submitted_by: number;
		estimated_mins: number;
		status: string;
		vote_count: number;
		voted: boolean;
		created_at: string;
	}

	let meeting = $state<Meeting | null>(null);
	let attendees = $state<Attendee[]>([]);
	let topics = $state<Topic[]>([]);
	let loading = $state(true);
	let error = $state('');
	let actionLoading = $state(false);

	// New topic form
	let showTopicForm = $state(false);
	let newTitle = $state('');
	let newDescription = $state('');
	let newCategory = $state('');
	let newMins = $state(10);
	let topicLoading = $state(false);

	// Filter
	let categoryFilter = $state('');

	const id = $derived(page.params.id);

	onMount(() => {
		loadMeeting();
		loadTopics();
	});

	async function loadMeeting() {
		loading = true;
		try {
			const data = await api.get<{ meeting: Meeting; attendees: Attendee[] }>(`/meetings/${id}`);
			meeting = data.meeting;
			attendees = data.attendees;
		} catch (e) {
			if (e instanceof ApiError) error = e.message;
			else error = 'Laden fehlgeschlagen';
		} finally {
			loading = false;
		}
	}

	async function loadTopics() {
		try {
			topics = await api.get<Topic[]>(`/meetings/${id}/topics`);
		} catch {
			// silent
		}
	}

	async function startMeeting() {
		actionLoading = true;
		try {
			meeting = await api.post<Meeting>(`/meetings/${id}/start`);
		} catch (e) {
			if (e instanceof ApiError) error = e.message;
		} finally {
			actionLoading = false;
		}
	}

	async function closeMeeting() {
		actionLoading = true;
		try {
			meeting = await api.post<Meeting>(`/meetings/${id}/close`);
		} catch (e) {
			if (e instanceof ApiError) error = e.message;
		} finally {
			actionLoading = false;
		}
	}

	async function submitTopic(e: Event) {
		e.preventDefault();
		topicLoading = true;
		try {
			await api.post('/topics', {
				meeting_id: Number(id),
				title: newTitle,
				description: newDescription || undefined,
				category: newCategory || undefined,
				estimated_mins: newMins
			});
			newTitle = '';
			newDescription = '';
			newCategory = '';
			newMins = 10;
			showTopicForm = false;
			await loadTopics();
		} catch (e) {
			if (e instanceof ApiError) error = e.message;
		} finally {
			topicLoading = false;
		}
	}

	async function vote(topicId: number) {
		try {
			const updated = await api.post<Topic>(`/topics/${topicId}/vote`);
			topics = topics.map((t) => (t.id === topicId ? updated : t));
		} catch (e) {
			if (e instanceof ApiError) error = e.message;
		}
	}

	async function unvote(topicId: number) {
		try {
			const updated = await api.delete<Topic>(`/topics/${topicId}/vote`);
			topics = topics.map((t) => (t.id === topicId ? updated : t));
		} catch (e) {
			if (e instanceof ApiError) error = e.message;
		}
	}

	async function deleteTopic(topicId: number) {
		try {
			await api.delete(`/topics/${topicId}`);
			topics = topics.filter((t) => t.id !== topicId);
		} catch (e) {
			if (e instanceof ApiError) error = e.message;
		}
	}

	function formatDate(iso: string) {
		return new Date(iso).toLocaleDateString('de-DE', {
			weekday: 'long',
			day: '2-digit',
			month: 'long',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
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

	function categoryLabel(cat: string | null) {
		if (!cat) return null;
		return (
			{ finanzen: 'Finanzen', satzung: 'Satzung', veranstaltungen: 'Veranstaltungen', sonstiges: 'Sonstiges' }[cat] ?? cat
		);
	}

	function categoryColor(cat: string | null) {
		if (!cat) return '';
		return (
			{
				finanzen: 'bg-yellow-100 text-yellow-800',
				satzung: 'bg-purple-100 text-purple-800',
				veranstaltungen: 'bg-pink-100 text-pink-800',
				sonstiges: 'bg-gray-100 text-gray-600'
			}[cat] ?? 'bg-gray-100 text-gray-600'
		);
	}

	const categories = ['finanzen', 'satzung', 'veranstaltungen', 'sonstiges'];

	const filteredTopics = $derived(
		categoryFilter ? topics.filter((t) => t.category === categoryFilter) : topics
	);

	const totalMins = $derived(filteredTopics.reduce((sum, t) => sum + t.estimated_mins, 0));

	const isAdmin = $derived($user?.role === 'admin');
	const canSubmitTopics = $derived(meeting?.status === 'open' || meeting?.status === 'active');
</script>

<div class="min-h-screen bg-gray-50">
	<header class="bg-white shadow-sm border-b border-gray-200">
		<div class="max-w-4xl mx-auto px-4 py-4">
			<a href="/" class="text-sm text-blue-600 hover:text-blue-800">&larr; Alle Sitzungen</a>
		</div>
	</header>

	<main class="max-w-4xl mx-auto px-4 py-8">
		{#if loading}
			<p class="text-gray-400">Laden...</p>
		{:else if error && !meeting}
			<p class="text-red-600">{error}</p>
		{:else if meeting}
			{#if error}
				<div class="bg-red-50 text-red-700 text-sm px-3 py-2 rounded mb-4">{error}</div>
			{/if}

			<!-- Meeting header -->
			<div class="bg-white border border-gray-200 rounded-lg p-6 mb-6">
				<div class="flex items-start justify-between">
					<div>
						<h1 class="text-2xl font-bold text-gray-900">{meeting.title}</h1>
						<p class="text-gray-600 mt-1">{formatDate(meeting.scheduled_at)}</p>
						{#if meeting.location}
							<p class="text-gray-500 text-sm mt-1">{meeting.location}</p>
						{/if}
					</div>
					<span class="text-xs font-medium px-2 py-1 rounded-full {statusColor(meeting.status)}">
						{statusLabel(meeting.status)}
					</span>
				</div>

				{#if isAdmin}
					<div class="mt-4 flex gap-2">
						{#if meeting.status === 'open'}
							<button
								onclick={startMeeting}
								disabled={actionLoading}
								class="px-4 py-2 bg-green-600 text-white text-sm font-medium rounded-md hover:bg-green-700 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
							>
								Sitzung starten
							</button>
						{/if}
						{#if meeting.status === 'active'}
							<button
								onclick={closeMeeting}
								disabled={actionLoading}
								class="px-4 py-2 bg-red-600 text-white text-sm font-medium rounded-md hover:bg-red-700 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
							>
								Sitzung beenden
							</button>
						{/if}
					</div>
				{/if}
			</div>

			<!-- Topics / Agenda -->
			<div class="bg-white border border-gray-200 rounded-lg p-6 mb-6">
				<div class="flex items-center justify-between mb-4">
					<div>
						<h2 class="text-lg font-semibold text-gray-900">
							Agenda ({filteredTopics.length} Themen, ~{totalMins} Min.)
						</h2>
					</div>
					{#if canSubmitTopics}
						<button
							onclick={() => (showTopicForm = !showTopicForm)}
							class="px-3 py-1.5 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 cursor-pointer"
						>
							{showTopicForm ? 'Abbrechen' : 'Thema einreichen'}
						</button>
					{/if}
				</div>

				<!-- Category filter -->
				{#if topics.length > 0}
					<div class="flex gap-2 mb-4 flex-wrap">
						<button
							onclick={() => (categoryFilter = '')}
							class="text-xs px-2 py-1 rounded-full cursor-pointer {categoryFilter === '' ? 'bg-gray-900 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}"
						>
							Alle
						</button>
						{#each categories as cat}
							<button
								onclick={() => (categoryFilter = categoryFilter === cat ? '' : cat)}
								class="text-xs px-2 py-1 rounded-full cursor-pointer {categoryFilter === cat ? 'bg-gray-900 text-white' : categoryColor(cat) + ' hover:opacity-80'}"
							>
								{categoryLabel(cat)}
							</button>
						{/each}
					</div>
				{/if}

				<!-- New topic form -->
				{#if showTopicForm}
					<form onsubmit={submitTopic} class="border border-gray-200 rounded-lg p-4 mb-4 space-y-3 bg-gray-50">
						<div>
							<input
								type="text"
								bind:value={newTitle}
								required
								placeholder="Thema"
								class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
							/>
						</div>
						<div>
							<textarea
								bind:value={newDescription}
								placeholder="Beschreibung (optional)"
								rows={2}
								class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
							></textarea>
						</div>
						<div class="flex gap-3">
							<select
								bind:value={newCategory}
								class="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
							>
								<option value="">Kategorie...</option>
								{#each categories as cat}
									<option value={cat}>{categoryLabel(cat)}</option>
								{/each}
							</select>
							<div class="flex items-center gap-2">
								<input
									type="number"
									bind:value={newMins}
									min={1}
									max={120}
									class="w-20 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
								/>
								<span class="text-sm text-gray-500">Min.</span>
							</div>
							<button
								type="submit"
								disabled={topicLoading}
								class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
							>
								{topicLoading ? 'Speichern...' : 'Einreichen'}
							</button>
						</div>
					</form>
				{/if}

				<!-- Topics list -->
				{#if filteredTopics.length === 0}
					<p class="text-gray-500 text-sm">Noch keine Themen eingereicht.</p>
				{:else}
					<ul class="divide-y divide-gray-100">
						{#each filteredTopics as topic, i}
							<li class="py-3 flex items-start gap-3">
								<!-- Vote button -->
								<div class="flex flex-col items-center min-w-[3rem] pt-0.5">
									{#if canSubmitTopics}
										<button
											onclick={() => topic.voted ? unvote(topic.id) : vote(topic.id)}
											class="flex flex-col items-center cursor-pointer group"
										>
											<svg
												class="w-5 h-5 {topic.voted ? 'text-blue-600' : 'text-gray-300 group-hover:text-blue-400'} transition-colors"
												fill={topic.voted ? 'currentColor' : 'none'}
												stroke="currentColor"
												viewBox="0 0 24 24"
											>
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
											</svg>
											<span class="text-sm font-semibold {topic.voted ? 'text-blue-600' : 'text-gray-500'}">
												{topic.vote_count}
											</span>
										</button>
									{:else}
										<span class="text-sm font-semibold text-gray-500">{topic.vote_count}</span>
									{/if}
								</div>

								<!-- Topic content -->
								<div class="flex-1 min-w-0">
									<div class="flex items-center gap-2">
										<span class="text-gray-400 text-sm font-mono">{i + 1}.</span>
										<span class="font-medium text-gray-900">{topic.title}</span>
										{#if topic.category}
											<span class="text-xs px-1.5 py-0.5 rounded-full {categoryColor(topic.category)}">
												{categoryLabel(topic.category)}
											</span>
										{/if}
										<span class="text-xs text-gray-400">{topic.estimated_mins} Min.</span>
									</div>
									{#if topic.description}
										<p class="text-sm text-gray-500 mt-0.5">{topic.description}</p>
									{/if}
								</div>

								<!-- Delete (submitter or admin) -->
								{#if canSubmitTopics && (topic.submitted_by === $user?.id || isAdmin)}
									<button
										onclick={() => deleteTopic(topic.id)}
										class="text-gray-300 hover:text-red-500 cursor-pointer text-sm"
										title="Löschen"
									>
										&times;
									</button>
								{/if}
							</li>
						{/each}
					</ul>
				{/if}
			</div>

			<!-- Attendees -->
			<div class="bg-white border border-gray-200 rounded-lg p-6">
				<h2 class="text-lg font-semibold text-gray-900 mb-3">
					Teilnehmer ({attendees.length})
				</h2>
				{#if attendees.length === 0}
					<p class="text-gray-500 text-sm">Noch keine Teilnehmer.</p>
				{:else}
					<ul class="divide-y divide-gray-100">
						{#each attendees as attendee}
							<li class="py-2 flex items-center justify-between">
								<div>
									<span class="text-gray-900">{attendee.name}</span>
									<span class="text-gray-400 text-sm ml-2">{attendee.email}</span>
								</div>
								{#if attendee.present}
									<span class="text-xs text-green-600 font-medium">Anwesend</span>
								{/if}
							</li>
						{/each}
					</ul>
				{/if}
			</div>
		{/if}
	</main>
</div>
