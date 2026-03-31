<script lang="ts">
	import { user } from '$lib/stores';
	import { goto } from '$app/navigation';
	import { ApiError } from '$lib/api';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;
		try {
			await user.login(email, password);
			goto('/');
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

<div class="min-h-screen flex items-center justify-center bg-gray-50 px-4">
	<div class="w-full max-w-sm">
		<h1 class="text-2xl font-bold text-center text-gray-900 mb-8">Vereinstool</h1>

		<form onsubmit={handleSubmit} class="bg-white shadow-sm rounded-lg p-6 space-y-4 border border-gray-200">
			{#if error}
				<div class="bg-red-50 text-red-700 text-sm px-3 py-2 rounded">{error}</div>
			{/if}

			<div>
				<label for="email" class="block text-sm font-medium text-gray-700 mb-1">E-Mail</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					required
					autocomplete="email"
					class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
				/>
			</div>

			<div>
				<label for="password" class="block text-sm font-medium text-gray-700 mb-1">Passwort</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					required
					autocomplete="current-password"
					class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
				/>
			</div>

			<button
				type="submit"
				disabled={loading}
				class="w-full py-2 px-4 bg-blue-600 text-white font-medium rounded-md hover:bg-blue-700 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
			>
				{loading ? 'Anmelden...' : 'Anmelden'}
			</button>

			<div class="text-center">
				<a href="/forgot-password" class="text-sm text-blue-600 hover:text-blue-800">
					Passwort vergessen?
				</a>
			</div>
		</form>
	</div>
</div>
