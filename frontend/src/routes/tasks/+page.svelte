<script lang="ts">
	import { api, ApiError } from '$lib/api';
	import { user } from '$lib/stores';
	import { onMount } from 'svelte';

	interface Task {
		id: number;
		title: string;
		description: string | null;
		assigned_to: number | null;
		due_date: string | null;
		status: string;
		meeting_id?: number;
		created_at: string;
	}

	let tasks = $state<Task[]>([]);
	let loading = $state(true);
	let filter = $state<'mine' | 'open' | 'all'>('mine');
	let error = $state('');

	onMount(() => loadTasks());

	async function loadTasks() {
		loading = true;
		try {
			const params = filter === 'mine' ? '?assigned_to=me' : filter === 'open' ? '?status=open' : '';
			tasks = await api.get<Task[]>(`/tasks${params}`);
		} finally {
			loading = false;
		}
	}

	async function toggleStatus(task: Task) {
		const newStatus = task.status === 'open' ? 'done' : 'open';
		try {
			const updated = await api.patch<Task>(`/tasks/${task.id}`, { status: newStatus });
			tasks = tasks.map((t) => (t.id === task.id ? updated : t));
		} catch (e) {
			if (e instanceof ApiError) error = e.message;
		}
	}

	function switchFilter(f: 'mine' | 'open' | 'all') {
		filter = f;
		loadTasks();
	}

	const openTasks = $derived(tasks.filter((t) => t.status === 'open'));
	const doneTasks = $derived(tasks.filter((t) => t.status === 'done'));

	function isOverdue(dueDate: string | null) {
		if (!dueDate) return false;
		return new Date(dueDate) < new Date(new Date().toISOString().split('T')[0]);
	}
</script>

<div class="min-h-screen bg-gray-50">
	<header class="bg-white shadow-sm border-b border-gray-200">
		<div class="max-w-4xl mx-auto px-4 py-4 flex items-center justify-between">
			<h1 class="text-xl font-bold text-gray-900">Aufgaben</h1>
			<a href="/" class="text-sm text-blue-600 hover:text-blue-800">&larr; Dashboard</a>
		</div>
	</header>

	<main class="max-w-4xl mx-auto px-4 py-8">
		{#if error}
			<div class="bg-red-50 text-red-700 text-sm px-3 py-2 rounded mb-4">{error}</div>
		{/if}

		<div class="flex gap-2 mb-6">
			<button onclick={() => switchFilter('mine')} class="text-sm px-3 py-1.5 rounded-md cursor-pointer {filter === 'mine' ? 'bg-gray-900 text-white' : 'bg-white border border-gray-300 text-gray-600 hover:bg-gray-50'}">Meine</button>
			<button onclick={() => switchFilter('open')} class="text-sm px-3 py-1.5 rounded-md cursor-pointer {filter === 'open' ? 'bg-gray-900 text-white' : 'bg-white border border-gray-300 text-gray-600 hover:bg-gray-50'}">Alle offenen</button>
			<button onclick={() => switchFilter('all')} class="text-sm px-3 py-1.5 rounded-md cursor-pointer {filter === 'all' ? 'bg-gray-900 text-white' : 'bg-white border border-gray-300 text-gray-600 hover:bg-gray-50'}">Alle</button>
		</div>

		{#if loading}
			<p class="text-gray-400">Laden...</p>
		{:else if tasks.length === 0}
			<p class="text-gray-500">Keine Aufgaben gefunden.</p>
		{:else}
			{#if openTasks.length > 0}
				<section class="mb-6">
					<h2 class="text-sm font-medium text-gray-500 uppercase tracking-wide mb-3">Offen ({openTasks.length})</h2>
					<div class="bg-white border border-gray-200 rounded-lg divide-y divide-gray-100">
						{#each openTasks as task}
							<div class="p-4 flex items-center gap-3">
								<button onclick={() => toggleStatus(task)} class="cursor-pointer" title="Als erledigt markieren">
									<svg class="w-5 h-5 text-gray-300 hover:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10" stroke-width="2" /></svg>
								</button>
								<div class="flex-1 min-w-0">
									<span class="text-gray-900">{task.title}</span>
									{#if task.description}<p class="text-sm text-gray-500 mt-0.5">{task.description}</p>{/if}
									<div class="flex gap-2 text-xs text-gray-500 mt-1">
										{#if task.meeting_id}<a href="/meetings/{task.meeting_id}" class="text-blue-600 hover:text-blue-800">Sitzung #{task.meeting_id}</a>{/if}
										{#if task.due_date}
											<span class="{isOverdue(task.due_date) ? 'text-red-600 font-medium' : ''}">
												Fällig: {new Date(task.due_date).toLocaleDateString('de-DE')}
											</span>
										{/if}
									</div>
								</div>
							</div>
						{/each}
					</div>
				</section>
			{/if}

			{#if doneTasks.length > 0}
				<section>
					<h2 class="text-sm font-medium text-gray-500 uppercase tracking-wide mb-3">Erledigt ({doneTasks.length})</h2>
					<div class="bg-white border border-gray-200 rounded-lg divide-y divide-gray-100">
						{#each doneTasks as task}
							<div class="p-4 flex items-center gap-3">
								<button onclick={() => toggleStatus(task)} class="cursor-pointer" title="Wieder öffnen">
									<svg class="w-5 h-5 text-green-500" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" /></svg>
								</button>
								<span class="line-through text-gray-400">{task.title}</span>
							</div>
						{/each}
					</div>
				</section>
			{/if}
		{/if}
	</main>
</div>
