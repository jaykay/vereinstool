<script lang="ts">
	import { user } from '$lib/stores';
	import { goto } from '$app/navigation';

	async function handleLogout() {
		await user.logout();
		goto('/login');
	}
</script>

{#if $user}
	<div class="min-h-screen bg-gray-50">
		<header class="bg-white shadow-sm border-b border-gray-200">
			<div class="max-w-4xl mx-auto px-4 py-4 flex items-center justify-between">
				<h1 class="text-xl font-bold text-gray-900">Vereinstool</h1>
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
			<p class="text-gray-600">Willkommen, {$user.name}! Das Dashboard kommt in Phase 3.</p>
		</main>
	</div>
{:else}
	<div class="min-h-screen flex items-center justify-center bg-gray-50">
		<p class="text-gray-400">Laden...</p>
	</div>
{/if}
