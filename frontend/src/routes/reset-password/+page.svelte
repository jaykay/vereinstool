<script lang="ts">
	import { api, ApiError } from '$lib/api';
	import { page } from '$app/state';

	let password = $state('');
	let confirmPassword = $state('');
	let success = $state(false);
	let error = $state('');
	let loading = $state(false);

	const token = $derived(page.url.searchParams.get('token') ?? '');

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';

		if (password !== confirmPassword) {
			error = 'Passwörter stimmen nicht überein';
			return;
		}

		if (password.length < 8) {
			error = 'Passwort muss mindestens 8 Zeichen lang sein';
			return;
		}

		loading = true;
		try {
			await api.post('/auth/reset-password', { token, password });
			success = true;
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
		<h1 class="text-2xl font-bold text-center text-gray-900 mb-8">Neues Passwort setzen</h1>

		{#if !token}
			<div class="bg-white shadow-sm rounded-lg p-6 border border-gray-200">
				<p class="text-red-600 text-sm">Kein gültiger Reset-Link. Bitte fordere einen neuen an.</p>
				<div class="mt-4 text-center">
					<a href="/forgot-password" class="text-sm text-blue-600 hover:text-blue-800">
						Neuen Link anfordern
					</a>
				</div>
			</div>
		{:else if success}
			<div class="bg-white shadow-sm rounded-lg p-6 border border-gray-200">
				<p class="text-green-700 text-sm">Passwort wurde erfolgreich geändert!</p>
				<div class="mt-4 text-center">
					<a href="/login" class="text-sm text-blue-600 hover:text-blue-800">Jetzt anmelden</a>
				</div>
			</div>
		{:else}
			<form onsubmit={handleSubmit} class="bg-white shadow-sm rounded-lg p-6 space-y-4 border border-gray-200">
				{#if error}
					<div class="bg-red-50 text-red-700 text-sm px-3 py-2 rounded">{error}</div>
				{/if}

				<div>
					<label for="password" class="block text-sm font-medium text-gray-700 mb-1">
						Neues Passwort
					</label>
					<input
						id="password"
						type="password"
						bind:value={password}
						required
						minlength={8}
						autocomplete="new-password"
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
					/>
				</div>

				<div>
					<label for="confirm" class="block text-sm font-medium text-gray-700 mb-1">
						Passwort bestätigen
					</label>
					<input
						id="confirm"
						type="password"
						bind:value={confirmPassword}
						required
						minlength={8}
						autocomplete="new-password"
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
					/>
				</div>

				<button
					type="submit"
					disabled={loading}
					class="w-full py-2 px-4 bg-blue-600 text-white font-medium rounded-md hover:bg-blue-700 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
				>
					{loading ? 'Speichern...' : 'Passwort speichern'}
				</button>
			</form>
		{/if}
	</div>
</div>
