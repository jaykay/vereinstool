<script lang="ts">
	import { user } from '$lib/stores';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	interface Meeting {
		id: number;
		title: string;
		scheduled_at: string;
		location: string | null;
		status: string;
	}

	let meetings = $state<Meeting[]>([]);
	let loading = $state(true);

	onMount(async () => {
		try {
			meetings = await api.get<Meeting[]>('/meetings');
		} finally {
			loading = false;
		}
	});

	async function handleLogout() {
		await user.logout();
		goto('/login');
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

	function formatDate(iso: string) {
		return new Date(iso).toLocaleDateString('de-DE', {
			weekday: 'short',
			day: '2-digit',
			month: '2-digit',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
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
					<button
						onclick={handleLogout}
						class="text-sm text-gray-500 hover:text-gray-700 cursor-pointer"
					>
						Abmelden
					</button>
				</div>
			</div>
		</header>

		<main class="max-w-4xl mx-auto px-4 py-8">
			<div class="flex items-center justify-between mb-6">
				<h2 class="text-lg font-semibold text-gray-900">Sitzungen</h2>
				{#if $user.role === 'admin'}
					<a
						href="/meetings/new"
						class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700"
					>
						Neue Sitzung
					</a>
				{/if}
			</div>

			{#if loading}
				<p class="text-gray-400">Laden...</p>
			{:else if meetings.length === 0}
				<p class="text-gray-500">Noch keine Sitzungen vorhanden.</p>
			{:else}
				{#if activeMeetings.length > 0}
					<section class="mb-8">
						<h3 class="text-sm font-medium text-gray-500 uppercase tracking-wide mb-3">Aktiv</h3>
						{#each activeMeetings as meeting}
							{@render meetingCard(meeting)}
						{/each}
					</section>
				{/if}

				{#if openMeetings.length > 0}
					<section class="mb-8">
						<h3 class="text-sm font-medium text-gray-500 uppercase tracking-wide mb-3">Geplant</h3>
						{#each openMeetings as meeting}
							{@render meetingCard(meeting)}
						{/each}
					</section>
				{/if}

				{#if closedMeetings.length > 0}
					<section>
						<h3 class="text-sm font-medium text-gray-500 uppercase tracking-wide mb-3">
							Abgeschlossen
						</h3>
						{#each closedMeetings as meeting}
							{@render meetingCard(meeting)}
						{/each}
					</section>
				{/if}
			{/if}
		</main>
	</div>
{:else}
	<div class="min-h-screen flex items-center justify-center bg-gray-50">
		<p class="text-gray-400">Laden...</p>
	</div>
{/if}

{#snippet meetingCard(meeting: Meeting)}
	<a
		href="/meetings/{meeting.id}"
		class="block bg-white border border-gray-200 rounded-lg p-4 mb-2 hover:border-gray-300 transition-colors"
	>
		<div class="flex items-center justify-between">
			<div>
				<span class="font-medium text-gray-900">{meeting.title}</span>
				<div class="text-sm text-gray-500 mt-1">
					{formatDate(meeting.scheduled_at)}
					{#if meeting.location}
						<span class="ml-2">&middot; {meeting.location}</span>
					{/if}
				</div>
			</div>
			<span class="text-xs font-medium px-2 py-1 rounded-full {statusColor(meeting.status)}">
				{statusLabel(meeting.status)}
			</span>
		</div>
	</a>
{/snippet}
