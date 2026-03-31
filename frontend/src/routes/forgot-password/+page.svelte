<script lang="ts">
	import { api, ApiError } from '$lib/api';

	let email = $state('');
	let sent = $state(false);
	let error = $state('');
	let loading = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;
		try {
			await api.post('/auth/forgot-password', { email });
			sent = true;
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
		<h1 class="text-2xl font-bold text-center text-gray-900 mb-8">Passwort vergessen</h1>

		{#if sent}
			<div class="bg-white shadow-sm rounded-lg p-6 border border-gray-200">
				<p class="text-gray-700 text-sm">
					Falls ein Konto mit dieser E-Mail existiert, wurde ein Reset-Link gesendet. Bitte prüfe dein Postfach.
				</p>
				<div class="mt-4 text-center">
					<a href="/login" class="text-sm text-blue-600 hover:text-blue-800">Zurück zum Login</a>
				</div>
			</div>
		{:else}
			<form onsubmit={handleSubmit} class="bg-white shadow-sm rounded-lg p-6 space-y-4 border border-gray-200">
				{#if error}
					<div class="bg-red-50 text-red-700 text-sm px-3 py-2 rounded">{error}</div>
				{/if}

				<p class="text-sm text-gray-600">
					Gib deine E-Mail-Adresse ein. Du erhältst einen Link zum Zurücksetzen deines Passworts.
				</p>

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

				<button
					type="submit"
					disabled={loading}
					class="w-full py-2 px-4 bg-blue-600 text-white font-medium rounded-md hover:bg-blue-700 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
				>
					{loading ? 'Senden...' : 'Reset-Link senden'}
				</button>

				<div class="text-center">
					<a href="/login" class="text-sm text-blue-600 hover:text-blue-800">Zurück zum Login</a>
				</div>
			</form>
		{/if}
	</div>
</div>
