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

	interface Task {
		id: number;
		topic_id?: number;
		meeting_id?: number;
		title: string;
		description: string | null;
		assigned_to: number | null;
		due_date: string | null;
		status: string;
		created_at: string;
	}

	let meeting = $state<Meeting | null>(null);
	let attendees = $state<Attendee[]>([]);
	let topics = $state<Topic[]>([]);
	let decisions = $state<Decision[]>([]);
	let tasks = $state<Task[]>([]);
	let users = $state<{ id: number; name: string }[]>([]);
	let loading = $state(true);
	let error = $state('');
	let actionLoading = $state(false);

	// Topic form
	let showTopicForm = $state(false);
	let newTitle = $state('');
	let newDescription = $state('');
	let newCategory = $state('');
	let newMins = $state(10);
	let topicLoading = $state(false);

	// Decision form
	let showDecisionForm = $state(false);
	let decTopicId = $state(0);
	let decText = $state('');
	let decYes = $state<number | undefined>(undefined);
	let decNo = $state<number | undefined>(undefined);
	let decAbstain = $state<number | undefined>(undefined);
	let decLoading = $state(false);

	// Task form
	let showTaskForm = $state(false);
	let taskTitle = $state('');
	let taskDescription = $state('');
	let taskAssignedTo = $state<number | undefined>(undefined);
	let taskDueDate = $state('');
	let taskLoading = $state(false);

	// Filter
	let categoryFilter = $state('');

	// Pool topics for assignment
	let poolTopics = $state<Topic[]>([]);

	const id = $derived(page.params.id);

	onMount(async () => {
		await Promise.all([loadMeeting(), loadTopics(), loadDecisions(), loadTasks(), loadUsers(), loadPool()]);
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
		try { topics = await api.get<Topic[]>(`/meetings/${id}/topics`); } catch {}
	}

	async function loadDecisions() {
		try { decisions = await api.get<Decision[]>(`/decisions?meeting_id=${id}`); } catch {}
	}

	async function loadTasks() {
		try { tasks = await api.get<Task[]>(`/tasks?meeting_id=${id}`); } catch {}
	}

	async function loadUsers() {
		try { users = await api.get<{ id: number; name: string }[]>('/users'); } catch {}
	}

	async function loadPool() {
		try { poolTopics = await api.get<Topic[]>('/topics/pool'); } catch {}
	}

	async function assignTopic(topicId: number) {
		try {
			await api.post(`/topics/${topicId}/assign`, { meeting_id: Number(id) });
			poolTopics = poolTopics.filter((t) => t.id !== topicId);
			await loadTopics();
		} catch (e) { if (e instanceof ApiError) error = e.message; }
	}

	async function startMeeting() {
		actionLoading = true;
		try { meeting = await api.post<Meeting>(`/meetings/${id}/start`); }
		catch (e) { if (e instanceof ApiError) error = e.message; }
		finally { actionLoading = false; }
	}

	async function closeMeeting() {
		actionLoading = true;
		try { meeting = await api.post<Meeting>(`/meetings/${id}/close`); }
		catch (e) { if (e instanceof ApiError) error = e.message; }
		finally { actionLoading = false; }
	}

	async function submitTopic(e: Event) {
		e.preventDefault();
		topicLoading = true;
		try {
			await api.post('/topics', {
				meeting_id: Number(id), title: newTitle,
				description: newDescription || undefined,
				category: newCategory || undefined, estimated_mins: newMins
			});
			newTitle = ''; newDescription = ''; newCategory = ''; newMins = 10; showTopicForm = false;
			await loadTopics();
		} catch (e) { if (e instanceof ApiError) error = e.message; }
		finally { topicLoading = false; }
	}

	async function vote(topicId: number) {
		try {
			const updated = await api.post<Topic>(`/topics/${topicId}/vote`);
			topics = topics.map((t) => (t.id === topicId ? updated : t));
		} catch (e) { if (e instanceof ApiError) error = e.message; }
	}

	async function unvote(topicId: number) {
		try {
			const updated = await api.delete<Topic>(`/topics/${topicId}/vote`);
			topics = topics.map((t) => (t.id === topicId ? updated : t));
		} catch (e) { if (e instanceof ApiError) error = e.message; }
	}

	async function deleteTopic(topicId: number) {
		try { await api.delete(`/topics/${topicId}`); topics = topics.filter((t) => t.id !== topicId); }
		catch (e) { if (e instanceof ApiError) error = e.message; }
	}

	async function submitDecision(e: Event) {
		e.preventDefault();
		decLoading = true;
		try {
			await api.post('/decisions', {
				topic_id: decTopicId, meeting_id: Number(id), text: decText,
				votes_yes: decYes, votes_no: decNo, votes_abstain: decAbstain
			});
			decTopicId = 0; decText = ''; decYes = undefined; decNo = undefined; decAbstain = undefined;
			showDecisionForm = false;
			await loadDecisions();
		} catch (e) { if (e instanceof ApiError) error = e.message; }
		finally { decLoading = false; }
	}

	async function submitTask(e: Event) {
		e.preventDefault();
		taskLoading = true;
		try {
			await api.post('/tasks', {
				meeting_id: Number(id), title: taskTitle,
				description: taskDescription || undefined,
				assigned_to: taskAssignedTo || undefined,
				due_date: taskDueDate || undefined
			});
			taskTitle = ''; taskDescription = ''; taskAssignedTo = undefined; taskDueDate = '';
			showTaskForm = false;
			await loadTasks();
		} catch (e) { if (e instanceof ApiError) error = e.message; }
		finally { taskLoading = false; }
	}

	async function toggleTaskStatus(task: Task) {
		const newStatus = task.status === 'open' ? 'done' : 'open';
		try {
			const updated = await api.patch<Task>(`/tasks/${task.id}`, { status: newStatus });
			tasks = tasks.map((t) => (t.id === task.id ? updated : t));
		} catch (e) { if (e instanceof ApiError) error = e.message; }
	}

	function formatDate(iso: string) {
		return new Date(iso).toLocaleDateString('de-DE', {
			weekday: 'long', day: '2-digit', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit'
		});
	}

	function statusLabel(s: string) { return { open: 'Geplant', active: 'Aktiv', closed: 'Abgeschlossen' }[s] ?? s; }
	function statusColor(s: string) { return { open: 'bg-blue-100 text-blue-800', active: 'bg-green-100 text-green-800', closed: 'bg-gray-100 text-gray-500' }[s] ?? 'bg-gray-100 text-gray-500'; }
	function categoryLabel(c: string | null) { if (!c) return null; return { finanzen: 'Finanzen', satzung: 'Satzung', veranstaltungen: 'Veranstaltungen', sonstiges: 'Sonstiges' }[c] ?? c; }
	function categoryColor(c: string | null) { if (!c) return ''; return { finanzen: 'bg-yellow-100 text-yellow-800', satzung: 'bg-purple-100 text-purple-800', veranstaltungen: 'bg-pink-100 text-pink-800', sonstiges: 'bg-gray-100 text-gray-600' }[c] ?? 'bg-gray-100 text-gray-600'; }
	function userName(uid: number | null) { if (!uid) return '–'; return users.find((u) => u.id === uid)?.name ?? `#${uid}`; }
	function topicTitle(tid: number) { return topics.find((t) => t.id === tid)?.title ?? `Thema #${tid}`; }

	const categories = ['finanzen', 'satzung', 'veranstaltungen', 'sonstiges'];
	const filteredTopics = $derived(categoryFilter ? topics.filter((t) => t.category === categoryFilter) : topics);
	const totalMins = $derived(filteredTopics.reduce((sum, t) => sum + t.estimated_mins, 0));
	const isAdmin = $derived($user?.role === 'admin');
	const canSubmitTopics = $derived(meeting?.status === 'open' || meeting?.status === 'active');
	const canRecordDecisions = $derived(meeting?.status === 'active' && isAdmin);
	const canCreateTasks = $derived(meeting?.status === 'active' || meeting?.status === 'closed');
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
					<span class="text-xs font-medium px-2 py-1 rounded-full {statusColor(meeting.status)}">{statusLabel(meeting.status)}</span>
				</div>
				{#if isAdmin}
					<div class="mt-4 flex gap-2">
						{#if meeting.status === 'open'}
							<button onclick={startMeeting} disabled={actionLoading} class="px-4 py-2 bg-green-600 text-white text-sm font-medium rounded-md hover:bg-green-700 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed">Sitzung starten</button>
						{/if}
						{#if meeting.status === 'active'}
							<button onclick={closeMeeting} disabled={actionLoading} class="px-4 py-2 bg-red-600 text-white text-sm font-medium rounded-md hover:bg-red-700 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed">Sitzung beenden</button>
						{/if}
					</div>
				{/if}
			</div>

			<!-- Topics / Agenda -->
			<div class="bg-white border border-gray-200 rounded-lg p-6 mb-6">
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-lg font-semibold text-gray-900">Agenda ({filteredTopics.length} Themen, ~{totalMins} Min.)</h2>
					{#if canSubmitTopics}
						<button onclick={() => (showTopicForm = !showTopicForm)} class="px-3 py-1.5 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 cursor-pointer">{showTopicForm ? 'Abbrechen' : 'Thema einreichen'}</button>
					{/if}
				</div>

				{#if topics.length > 0}
					<div class="flex gap-2 mb-4 flex-wrap">
						<button onclick={() => (categoryFilter = '')} class="text-xs px-2 py-1 rounded-full cursor-pointer {categoryFilter === '' ? 'bg-gray-900 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}">Alle</button>
						{#each categories as cat}
							<button onclick={() => (categoryFilter = categoryFilter === cat ? '' : cat)} class="text-xs px-2 py-1 rounded-full cursor-pointer {categoryFilter === cat ? 'bg-gray-900 text-white' : categoryColor(cat) + ' hover:opacity-80'}">{categoryLabel(cat)}</button>
						{/each}
					</div>
				{/if}

				{#if showTopicForm}
					<form onsubmit={submitTopic} class="border border-gray-200 rounded-lg p-4 mb-4 space-y-3 bg-gray-50">
						<input type="text" bind:value={newTitle} required placeholder="Thema" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent" />
						<textarea bind:value={newDescription} placeholder="Beschreibung (optional)" rows={2} class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"></textarea>
						<div class="flex gap-3">
							<select bind:value={newCategory} class="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500">
								<option value="">Kategorie...</option>
								{#each categories as cat}<option value={cat}>{categoryLabel(cat)}</option>{/each}
							</select>
							<div class="flex items-center gap-2">
								<input type="number" bind:value={newMins} min={1} max={120} class="w-20 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500" />
								<span class="text-sm text-gray-500">Min.</span>
							</div>
							<button type="submit" disabled={topicLoading} class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed">{topicLoading ? 'Speichern...' : 'Einreichen'}</button>
						</div>
					</form>
				{/if}

				{#if filteredTopics.length === 0}
					<p class="text-gray-500 text-sm">Noch keine Themen eingereicht.</p>
				{:else}
					<ul class="divide-y divide-gray-100">
						{#each filteredTopics as topic, i}
							<li class="py-3 flex items-start gap-3">
								<div class="flex flex-col items-center min-w-[3rem] pt-0.5">
									{#if canSubmitTopics}
										<button onclick={() => topic.voted ? unvote(topic.id) : vote(topic.id)} class="flex flex-col items-center cursor-pointer group">
											<svg class="w-5 h-5 {topic.voted ? 'text-blue-600' : 'text-gray-300 group-hover:text-blue-400'} transition-colors" fill={topic.voted ? 'currentColor' : 'none'} stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" /></svg>
											<span class="text-sm font-semibold {topic.voted ? 'text-blue-600' : 'text-gray-500'}">{topic.vote_count}</span>
										</button>
									{:else}
										<span class="text-sm font-semibold text-gray-500">{topic.vote_count}</span>
									{/if}
								</div>
								<div class="flex-1 min-w-0">
									<div class="flex items-center gap-2">
										<span class="text-gray-400 text-sm font-mono">{i + 1}.</span>
										<span class="font-medium text-gray-900">{topic.title}</span>
										{#if topic.category}<span class="text-xs px-1.5 py-0.5 rounded-full {categoryColor(topic.category)}">{categoryLabel(topic.category)}</span>{/if}
										<span class="text-xs text-gray-400">{topic.estimated_mins} Min.</span>
									</div>
									{#if topic.description}<p class="text-sm text-gray-500 mt-0.5">{topic.description}</p>{/if}
								</div>
								{#if canSubmitTopics && (topic.submitted_by === $user?.id || isAdmin)}
									<button onclick={() => deleteTopic(topic.id)} class="text-gray-300 hover:text-red-500 cursor-pointer text-sm" title="Löschen">&times;</button>
								{/if}
							</li>
						{/each}
					</ul>
				{/if}

				<!-- Pool topics to assign -->
				{#if (meeting?.status === 'open' || meeting?.status === 'active') && poolTopics.length > 0}
					<div class="mt-4 border-t border-gray-100 pt-4">
						<h3 class="text-sm font-medium text-gray-500 mb-3">Aus Themenpool übernehmen ({poolTopics.length})</h3>
						<div class="space-y-2">
							{#each poolTopics as pt}
								<div class="flex items-center justify-between bg-gray-50 rounded-md px-3 py-2">
									<div class="flex items-center gap-2">
										<span class="text-sm font-semibold text-gray-500">{pt.vote_count}</span>
										<span class="text-sm text-gray-900">{pt.title}</span>
										{#if pt.category}<span class="text-xs px-1.5 py-0.5 rounded-full {categoryColor(pt.category)}">{categoryLabel(pt.category)}</span>{/if}
									</div>
									<button onclick={() => assignTopic(pt.id)} class="text-xs px-2 py-1 bg-blue-600 text-white rounded hover:bg-blue-700 cursor-pointer">Übernehmen</button>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</div>

			<!-- Decisions -->
			<div class="bg-white border border-gray-200 rounded-lg p-6 mb-6">
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-lg font-semibold text-gray-900">Beschlüsse ({decisions.length})</h2>
					{#if canRecordDecisions}
						<button onclick={() => (showDecisionForm = !showDecisionForm)} class="px-3 py-1.5 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 cursor-pointer">{showDecisionForm ? 'Abbrechen' : 'Beschluss erfassen'}</button>
					{/if}
				</div>

				{#if showDecisionForm}
					<form onsubmit={submitDecision} class="border border-gray-200 rounded-lg p-4 mb-4 space-y-3 bg-gray-50">
						<select bind:value={decTopicId} required class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500">
							<option value={0} disabled>Thema wählen...</option>
							{#each topics as t}<option value={t.id}>{t.title}</option>{/each}
						</select>
						<textarea bind:value={decText} required placeholder="Beschlusstext" rows={2} class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"></textarea>
						<div class="flex gap-3 items-center">
							<div class="flex items-center gap-1"><label class="text-xs text-gray-500">Ja:</label><input type="number" bind:value={decYes} min={0} class="w-16 px-2 py-1 border border-gray-300 rounded-md text-sm" /></div>
							<div class="flex items-center gap-1"><label class="text-xs text-gray-500">Nein:</label><input type="number" bind:value={decNo} min={0} class="w-16 px-2 py-1 border border-gray-300 rounded-md text-sm" /></div>
							<div class="flex items-center gap-1"><label class="text-xs text-gray-500">Enthaltung:</label><input type="number" bind:value={decAbstain} min={0} class="w-16 px-2 py-1 border border-gray-300 rounded-md text-sm" /></div>
							<button type="submit" disabled={decLoading} class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed">{decLoading ? 'Speichern...' : 'Speichern'}</button>
						</div>
					</form>
				{/if}

				{#if decisions.length === 0}
					<p class="text-gray-500 text-sm">Noch keine Beschlüsse.</p>
				{:else}
					<ul class="divide-y divide-gray-100">
						{#each decisions as dec}
							<li class="py-3">
								<div class="flex items-start justify-between">
									<div>
										<p class="text-sm text-gray-500">{topicTitle(dec.topic_id)}</p>
										<p class="text-gray-900 mt-0.5">{dec.text}</p>
									</div>
								</div>
								{#if dec.votes_yes !== undefined || dec.votes_no !== undefined}
									<div class="flex gap-3 mt-1 text-xs text-gray-500">
										{#if dec.votes_yes !== undefined}<span class="text-green-600">Ja: {dec.votes_yes}</span>{/if}
										{#if dec.votes_no !== undefined}<span class="text-red-600">Nein: {dec.votes_no}</span>{/if}
										{#if dec.votes_abstain !== undefined}<span>Enthaltung: {dec.votes_abstain}</span>{/if}
									</div>
								{/if}
							</li>
						{/each}
					</ul>
				{/if}
			</div>

			<!-- Tasks -->
			<div class="bg-white border border-gray-200 rounded-lg p-6 mb-6">
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-lg font-semibold text-gray-900">Aufgaben ({tasks.length})</h2>
					{#if canCreateTasks}
						<button onclick={() => (showTaskForm = !showTaskForm)} class="px-3 py-1.5 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 cursor-pointer">{showTaskForm ? 'Abbrechen' : 'Aufgabe erstellen'}</button>
					{/if}
				</div>

				{#if showTaskForm}
					<form onsubmit={submitTask} class="border border-gray-200 rounded-lg p-4 mb-4 space-y-3 bg-gray-50">
						<input type="text" bind:value={taskTitle} required placeholder="Aufgabe" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent" />
						<textarea bind:value={taskDescription} placeholder="Beschreibung (optional)" rows={2} class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"></textarea>
						<div class="flex gap-3">
							<select bind:value={taskAssignedTo} class="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500">
								<option value={undefined}>Zuständig...</option>
								{#each users as u}<option value={u.id}>{u.name}</option>{/each}
							</select>
							<input type="date" bind:value={taskDueDate} class="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500" />
							<button type="submit" disabled={taskLoading} class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed">{taskLoading ? 'Speichern...' : 'Erstellen'}</button>
						</div>
					</form>
				{/if}

				{#if tasks.length === 0}
					<p class="text-gray-500 text-sm">Noch keine Aufgaben.</p>
				{:else}
					<ul class="divide-y divide-gray-100">
						{#each tasks as task}
							<li class="py-3 flex items-center gap-3">
								<button onclick={() => toggleTaskStatus(task)} class="cursor-pointer" title={task.status === 'open' ? 'Als erledigt markieren' : 'Wieder öffnen'}>
									{#if task.status === 'done'}
										<svg class="w-5 h-5 text-green-500" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" /></svg>
									{:else}
										<svg class="w-5 h-5 text-gray-300 hover:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10" stroke-width="2" /></svg>
									{/if}
								</button>
								<div class="flex-1 min-w-0">
									<span class="{task.status === 'done' ? 'line-through text-gray-400' : 'text-gray-900'}">{task.title}</span>
									<div class="flex gap-2 text-xs text-gray-500 mt-0.5">
										{#if task.assigned_to}<span>{userName(task.assigned_to)}</span>{/if}
										{#if task.due_date}<span>Fällig: {new Date(task.due_date).toLocaleDateString('de-DE')}</span>{/if}
									</div>
								</div>
							</li>
						{/each}
					</ul>
				{/if}
			</div>

			<!-- Attendees -->
			<div class="bg-white border border-gray-200 rounded-lg p-6">
				<h2 class="text-lg font-semibold text-gray-900 mb-3">Teilnehmer ({attendees.length})</h2>
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
								{#if attendee.present}<span class="text-xs text-green-600 font-medium">Anwesend</span>{/if}
							</li>
						{/each}
					</ul>
				{/if}
			</div>
		{/if}
	</main>
</div>
