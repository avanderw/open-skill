<script lang="ts">
	import '../lib/theme/tokens.css';
	import '../lib/theme/index.css';
	import favicon from '$lib/assets/favicon.svg';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { base } from '$app/paths';
	import { AUTH_CHANGE_EVENT, authReady, DEMO_AUTH, isAuthenticated, isProtectedPath, loginPath, signOut as authSignOut } from '$lib/auth';
	import { selectedDebugProfile } from '$lib/debug-profiles';
	import { api } from '$lib/api';
	import { onMount } from 'svelte';

	let { children } = $props();
	let authenticated = $state(false);
	let hydrated = $state(false);
	let menuOpen = $state(false);
	let handle = $state(DEMO_AUTH ? selectedDebugProfile().handle : '');
	let playerId = $state('');

	onMount(() => {
		const sync = () => {
			authenticated = isAuthenticated();
			if (DEMO_AUTH) handle = selectedDebugProfile().handle;
			if (!authenticated) {
				playerId = '';
				return;
			}
			// The database starts empty, so the player id is resolved server-side.
			api.me().then((me) => { playerId = me.playerId; handle = me.handle; }).catch(() => { playerId = ''; });
		};
		authReady().then(() => {
			sync();
			hydrated = true;
		});
		window.addEventListener(AUTH_CHANGE_EVENT, sync);
		return () => window.removeEventListener(AUTH_CHANGE_EVENT, sync);
	});

	$effect(() => {
		if (hydrated && authenticated === false && isProtectedPath($page.url.pathname) && $page.url.pathname !== '/login') {
			goto(loginPath($page.url.pathname + $page.url.search));
		}
	});

	async function signOut() {
		menuOpen = false;
		await authSignOut();
		goto(`${base}/login`);
	}
</script>

<svelte:head><link rel="icon" href={favicon} /><meta name="description" content="Community rankings and match reporting" /><title>Open-Skill</title></svelte:head>

<a class="skip-link" href="#main">Skip to content</a>
<nav class="site-nav" aria-label="Primary">
	<div class="site-nav-inner">
		<a class="brand" href={authenticated ? `${base}/communities` : `${base}/`}>Open&#8209;Skill</a>
		{#if authenticated}
			<div class="account-menu">
				<button class="account-toggle" aria-expanded={menuOpen} aria-haspopup="menu" onclick={() => (menuOpen = !menuOpen)}>
					<span class="account-avatar">{(handle || '?').slice(0, 2).toUpperCase()}</span><span>{handle}</span><span aria-hidden="true">⌄</span>
				</button>
				{#if menuOpen}
					<div class="account-dropdown" role="menu"><a role="menuitem" href={`${base}/profile/${playerId}`} onclick={() => (menuOpen = false)}>Profile</a><a role="menuitem" href={`${base}/communities`} onclick={() => (menuOpen = false)}>My communities</a><button role="menuitem" onclick={signOut}>Sign out</button></div>
				{/if}
			</div>
		{:else if $page.url.pathname !== '/login'}
			<a class="btn btn-primary sign-in" href={loginPath($page.url.pathname + $page.url.search)}>Sign in</a>
		{/if}
	</div>
</nav>

<main id="main">{@render children()}</main>

<style>
	.site-nav { border-bottom:1px solid var(--theme-border); background:rgba(11,11,11,.78); backdrop-filter:blur(6px); position:sticky; top:0; z-index:40; }
	.site-nav-inner { max-width:1120px; margin:0 auto; padding:14px 20px; display:flex; align-items:center; justify-content:space-between; gap:20px; }
	.brand { font-weight:800; font-size:18px; color:var(--theme-text); letter-spacing:.02em; }
	.sign-in { padding:8px 14px; font-size:13px; }
	.account-menu { position:relative; } .account-toggle { display:flex; align-items:center; gap:8px; border:1px solid var(--theme-border); border-radius:999px; padding:5px 10px 5px 5px; background:rgba(255,255,255,.03); color:var(--theme-text); cursor:pointer; font-size:13px; font-weight:700; }
	.account-avatar { display:grid; place-items:center; width:28px; height:28px; border-radius:50%; background:linear-gradient(135deg,#b83c08,var(--brand-primary)); color:#fff; font-size:10px; }
	.account-dropdown { position:absolute; right:0; top:calc(100% + 8px); min-width:170px; padding:6px; border:1px solid var(--theme-border); border-radius:14px; background:#121212; box-shadow:0 16px 35px rgba(0,0,0,.45); }
	.account-dropdown a,.account-dropdown button { display:block; width:100%; padding:10px; border:0; border-radius:8px; background:none; color:var(--theme-text); text-align:left; font-size:13px; cursor:pointer; } .account-dropdown a:hover,.account-dropdown button:hover { background:rgba(240,90,26,.14); color:var(--theme-accent); } 
</style>