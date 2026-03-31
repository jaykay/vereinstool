<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import { user } from '$lib/stores';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';

	let { children } = $props();

	const publicRoutes = ['/login', '/forgot-password', '/reset-password'];

	$effect(() => {
		user.load().then((u) => {
			const path = page.url.pathname;
			if (!u && !publicRoutes.some((r) => path.startsWith(r))) {
				goto('/login');
			} else if (u && path === '/login') {
				goto('/');
			}
		});
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>Vereinstool</title>
</svelte:head>

{@render children()}
