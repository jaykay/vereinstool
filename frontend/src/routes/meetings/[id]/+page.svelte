<script lang="ts">
	import { api, ApiError } from '$lib/api';
	import { user } from '$lib/stores';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

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

	let meeting = $state<Meeting | null>(null);
	let attendees = $state<Attendee[]>([]);
	let loading = $state(true);
	let error = $state('');
	let actionLoading = $state(false);

	const id = $derived(page.params.id);

	onMount(() => loadMeeting());

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

	async function startMeeting() {
		actionLoading = true;
		try {
			const updated = await api.post<Meeting>(`/meetings/${id}/start`);
			meeting = updated;
		} catch (e) {
			if (e instanceof ApiError) error = e.message;
		} finally {
			actionLoading = false;
		}
	}

	async function closeMeeting() {
		actionLoading = true;
		try {
			const updated = await api.post<Meeting>(`/meetings/${id}/close`);
			meeting = updated;
		} catch (e) {
			if (e instanceof ApiError) error = e.message;
		} finally {
			actionLoading = false;
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

	const isAdmin = $derived($user?.role === 'admin');
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

			<div class="bg-white border border-gray-200 rounded-lg p-6 mb-6">
				<div class="flex items-start justify-between">
					<div>
						<h1 class="text-2xl font-bold text-gray-900">{meeting.title}</h1>
						<p class="text-gray-600 mt-1">{formatDate(meeting.scheduled_at)}</p>
						{#if meeting.location}
							<p class="text-gray-500 text-sm mt-1">{meeting.location}</p>
						{/if}
					</div>
					<span
						class="text-xs font-medium px-2 py-1 rounded-full {statusColor(meeting.status)}"
					>
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

			<!-- Attendees -->
			<div class="bg-white border border-gray-200 rounded-lg p-6 mb-6">
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

			<!-- Topics placeholder for Phase 4 -->
			{#if meeting.status !== 'closed'}
				<div class="bg-white border border-gray-200 rounded-lg p-6 border-dashed">
					<h2 class="text-lg font-semibold text-gray-400">Themen</h2>
					<p class="text-gray-400 text-sm mt-1">Themen & Voting kommen in Phase 4.</p>
				</div>
			{/if}
		{/if}
	</main>
</div>
