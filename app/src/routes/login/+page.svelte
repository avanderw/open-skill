<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { base } from '$app/paths';
	import { AUTH_CHANGE_EVENT, authReady, DEMO_AUTH, isAuthenticated, loginDemo, signIn, type AuthProvider } from '$lib/auth';
	import { setDebugProfile } from '$lib/debug-profiles';

	const providers: { id: AuthProvider; label: string; copy: string; icon: string }[] = [
		{ id: 'google', label: 'Continue with Google', copy: 'Use a familiar account to get started.', icon: 'G' }
	];

	let email = $state('');
	let busy = $state(false);
	let error = $state('');
	let checkingSession = $state(true);

	function continueToApp() {
		const returnTo = $page.url.searchParams.get('returnTo') || `${base}/communities`;
		goto(returnTo.startsWith('/') && !returnTo.startsWith('//') && returnTo !== `${base}/` && returnTo !== '/' ? returnTo : `${base}/communities`);
	}

	onMount(() => {
		const redirectIfSignedIn = () => {
			if (isAuthenticated()) continueToApp();
			else checkingSession = false;
		};
		authReady().then(redirectIfSignedIn);
		window.addEventListener(AUTH_CHANGE_EVENT, redirectIfSignedIn);
		return () => window.removeEventListener(AUTH_CHANGE_EVENT, redirectIfSignedIn);
	});

	async function login(provider: AuthProvider) {
		if (DEMO_AUTH) {
			loginDemo();
			continueToApp();
			return;
		}
		busy = true;
		error = '';
		try {
			await signIn(provider);
			continueToApp();
		} catch (err) {
			const code = (err as { code?: string }).code ?? '';
			if (code !== 'auth/popup-closed-by-user' && code !== 'auth/cancelled-popup-request') {
				error = code === 'auth/account-exists-with-different-credential'
					? 'An account already exists with this email using a different sign-in provider.'
					: 'Sign-in failed. Please try again.';
			}
		} finally {
			busy = false;
		}
	}

	function loginWithEmail() {
		const username = email.trim().toLowerCase();
		if (!username) return;
		setDebugProfile(username);
		loginDemo();
		continueToApp();
	}
</script>

<svelte:head><title>Sign in — Open-Skill</title></svelte:head>

<div class="page-wrap">
	<div class="shell">
		<section class="card hero-card">
			<div>
				<span class="eyebrow">Open-Skill preview</span>
				<p class="label-caps">Competitive ranking</p>
				<h1>Bring your communities into focus.</h1>
				<p class="hero-copy">Track players, report results, and keep every competition moving with one shared ranking space.</p>
			</div>
			<div class="benefits">
				<div><strong>Rankings</strong><span>Modes that update after every match.</span></div>
				<div><strong>Communities</strong><span>Open or approval-based groups.</span></div>
				<div><strong>History</strong><span>See how every result shaped ratings.</span></div>
			</div>
		</section>

		<section class="card login-card">
			<p class="label-caps">Welcome back</p>
			<h2>Sign in to continue</h2>
			{#if checkingSession}
				<p class="muted">Checking your session…</p>
			{:else}
			{#if DEMO_AUTH}
				<p class="muted">Demo mode: choose a provider or enter an email to use as the temporary username. No authentication is performed.</p>
			{:else}
				<p class="muted">Choose a provider to sign in.</p>
			{/if}
			<div class="provider-list">
				{#each providers as provider}
					<button class="provider provider-{provider.id}" disabled={busy} onclick={() => login(provider.id)}>
						<span class="provider-icon">{provider.icon}</span>
						<span class="provider-copy"><strong>{provider.label}</strong><small>{provider.copy}</small></span>
						<span class="arrow">→</span>
					</button>
				{/each}
			</div>
			{#if error}<p class="error" role="alert">{error}</p>{/if}
			{#if DEMO_AUTH}
				<form class="email-login" onsubmit={(event) => { event.preventDefault(); loginWithEmail(); }}>
					<label for="email">Email</label>
					<div class="email-row">
						<input id="email" bind:value={email} type="email" autocomplete="email" placeholder="person@example.com" required />
						<button class="email-submit" type="submit">Continue</button>
					</div>
				</form>
			{/if}
			{/if}
			<a class="back" href={`${base}/leaderboard`}>← Continue browsing as guest</a>
		</section>
	</div>
</div>

<style>
	.page-wrap { min-height: calc(100vh - 65px); display:grid; place-items:center; padding:32px 20px 48px; }
	.shell { width:100%; max-width:1040px; display:grid; grid-template-columns:minmax(0,1.05fr) minmax(360px,.95fr); gap:24px; }
	.card { background:var(--theme-surface); border:1px solid var(--theme-border); border-radius:28px; box-shadow:0 24px 64px rgba(0,0,0,.55); }
	.hero-card { min-height:560px; padding:36px; display:flex; flex-direction:column; justify-content:space-between; background:linear-gradient(145deg,rgba(255,255,255,.04),rgba(255,255,255,.01)),var(--theme-surface); }
	.eyebrow { display:inline-flex; padding:10px 14px; border:1px solid var(--theme-border); border-radius:999px; background:rgba(255,255,255,.03); color:var(--theme-text-muted); font-size:12px; font-weight:700; margin-bottom:26px; }
	.label-caps { color:var(--theme-accent); font-size:11px; font-weight:800; letter-spacing:.12em; text-transform:uppercase; margin-bottom:10px; }
	h1 { max-width:8ch; font-size:clamp(34px,4vw,54px); line-height:.96; letter-spacing:-.04em; margin-bottom:16px; }
	.hero-copy { max-width:42ch; color:var(--theme-text-muted); font-size:15px; line-height:25px; }
	.benefits { display:grid; grid-template-columns:repeat(3,minmax(0,1fr)); gap:12px; }
	.benefits div { padding:16px; border-radius:18px; background:rgba(255,255,255,.04); border:1px solid var(--theme-border); }
	.benefits strong,.benefits span { display:block; } .benefits strong { font-size:13px; margin-bottom:6px; } .benefits span { font-size:12px; line-height:18px; color:var(--theme-text-muted); }
	.login-card { padding:32px; align-self:center; } .login-card h2 { font-size:28px; margin-bottom:8px; } .login-card > p:not(.label-caps) { font-size:14px; line-height:22px; margin-bottom:24px; }
	.provider-list { display:grid; gap:14px; margin-bottom:18px; }
	.provider { width:100%; display:flex; align-items:center; gap:16px; padding:18px; border:1px solid var(--theme-border); border-radius:22px; color:var(--theme-text); text-align:left; cursor:pointer; transition:transform .18s,border-color .18s,background .18s; background:rgba(255,255,255,.03); }
	.provider:hover { transform:translateY(-1px); border-color:var(--theme-accent); background:rgba(240,90,26,.08); }
	.provider-google { background:rgba(255,255,255,.92); color:#18181b; } .provider-github { background:rgba(255,255,255,.06); }
	.provider:disabled { opacity:.6; cursor:progress; transform:none; }
	.error { color:#ff8a80; font-size:13px; margin:-6px 0 14px; }
	.provider-icon { width:50px; height:50px; border-radius:16px; display:grid; place-items:center; flex:none; font-size:22px; font-weight:800; background:var(--theme-accent); color:#fff; } .provider-google .provider-icon { background:#ea4335; } .provider-github .provider-icon { background:#24292f; font-size:18px; }
	.provider-copy { flex:1; display:flex; flex-direction:column; gap:4px; } .provider-copy strong { font-size:16px; } .provider-copy small { font-size:12px; opacity:.72; } .arrow { font-size:20px; color:var(--theme-accent); }
	.email-login { display:grid; gap:8px; margin:4px 0 18px; }
	.email-login label { color:var(--theme-text-muted); font-size:12px; font-weight:800; letter-spacing:.08em; text-transform:uppercase; }
	.email-row { display:flex; gap:10px; }
	.email-row input { min-width:0; flex:1; border:1px solid var(--theme-border); border-radius:16px; padding:0 14px; background:rgba(255,255,255,.04); color:var(--theme-text); font:inherit; font-size:14px; }
	.email-row input:focus { outline:2px solid rgba(240,90,26,.35); border-color:var(--theme-accent); }
	.email-submit { flex:none; border:0; border-radius:16px; padding:0 16px; background:var(--theme-accent); color:#fff; font-weight:800; cursor:pointer; }
	.back { color:var(--theme-text-muted); font-size:13px; } .back:hover { color:var(--theme-accent); }
	@media (max-width:760px) { .shell { grid-template-columns:1fr; } .hero-card { min-height:auto; gap:48px; } .benefits { grid-template-columns:1fr; } }
	@media (max-width:480px) { .email-row { flex-direction:column; } .email-row input,.email-submit { min-height:48px; } }
</style>