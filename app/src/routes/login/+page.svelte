<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { base } from '$app/paths';
	import { AUTH_CHANGE_EVENT, authReady, DEMO_AUTH, hasSignedInBefore, isAuthenticated, loginDemo, signIn, type AuthProvider } from '$lib/auth';
	import { setDebugProfile } from '$lib/debug-profiles';

	const providers: { id: AuthProvider; label: string; copy: string; icon: string }[] = [
		{ id: 'google', label: 'Continue with Google', copy: 'Use a familiar account to get started.', icon: 'G' }
	];

	let email = $state('');
	let busy = $state(false);
	let error = $state('');
	let checkingSession = $state(true);
	let slowSession = $state(false);
	let sessionError = $state('');
	let returningUser = $state(false);
	let redirecting = $state(false);

	function continueToApp() {
		if (redirecting) return;
		redirecting = true;
		performance.mark('open-skill:redirect-start');
		const returnTo = $page.url.searchParams.get('returnTo') || `${base}/communities`;
		goto(returnTo.startsWith('/') && !returnTo.startsWith('//') && returnTo !== `${base}/` && returnTo !== '/' ? returnTo : `${base}/communities`).then(() => {
			performance.measure('open-skill:login-redirect', 'open-skill:redirect-start');
		}).catch((err: unknown) => {
			console.error('Unable to open the app after sign-in.', err);
			redirecting = false;
			busy = false;
			checkingSession = false;
			sessionError = 'You are signed in, but we could not open the app. Please try again.';
		});
	}

	onMount(() => {
		performance.mark('open-skill:login-mounted');
		returningUser = hasSignedInBefore();
		let active = true;
		const slowTimer = window.setTimeout(() => { slowSession = true; }, 4000);
		const redirectIfSignedIn = () => {
			if (!active) return;
			window.clearTimeout(slowTimer);
			if (isAuthenticated()) continueToApp();
			else checkingSession = false;
		};
		authReady().then(() => {
			if (!active) return;
			performance.measure('open-skill:login-session-wait', 'open-skill:login-mounted');
			redirectIfSignedIn();
		}).catch((err: unknown) => {
			if (!active) return;
			console.error('Unable to check the saved session.', err);
			window.clearTimeout(slowTimer);
			checkingSession = false;
			sessionError = 'We could not check your saved session. Check your connection and try again.';
		});
		window.addEventListener(AUTH_CHANGE_EVENT, redirectIfSignedIn);
		return () => {
			active = false;
			window.clearTimeout(slowTimer);
			window.removeEventListener(AUTH_CHANGE_EVENT, redirectIfSignedIn);
		};
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
				console.error('Sign-in failed.', err);
				error = code === 'auth/account-exists-with-different-credential'
					? 'An account already exists with this email using a different sign-in provider.'
					: 'Sign-in failed. Please try again.';
			}
		} finally {
			if (!redirecting) busy = false;
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
		<section class="card login-card" aria-labelledby="login-heading">
			<p class="label-caps">{returningUser ? 'Welcome back' : 'Get started'}</p>
			<h1 id="login-heading">Sign in to continue</h1>
			{#if sessionError}
				<p class="error session-error" role="alert">{sessionError}</p>
				<button class="retry" onclick={() => window.location.reload()}>Try again</button>
			{:else if checkingSession || redirecting}
				<div class="session-status" role="status" aria-live="polite">
					<span class="spinner" aria-hidden="true"></span>
					<div>
						<strong>{redirecting ? 'Taking you to your communities…' : 'Checking your saved session…'}</strong>
						<p>{redirecting ? 'You are signed in. Opening the app.' : slowSession ? 'This is taking longer than usual. Check your connection, or try again.' : 'You will continue automatically if you are already signed in.'}</p>
					</div>
				</div>
				{#if slowSession && !redirecting}<button class="retry" onclick={() => window.location.reload()}>Try again</button>{/if}
			{:else}
			{#if DEMO_AUTH}
				<p class="muted">Demo mode: choose a provider or enter an email to use as the temporary username. No authentication is performed.</p>
			{:else}
				<p class="muted">Choose a provider to sign in.</p>
			{/if}
			<div class="provider-list">
				{#each providers as provider}
					<button class="provider provider-{provider.id}" disabled={busy} onclick={() => login(provider.id)}>
						<span class="provider-icon" aria-hidden="true">{#if busy}<span class="spinner"></span>{:else}{provider.icon}{/if}</span>
						<span class="provider-copy"><strong>{busy ? 'Signing in with Google…' : provider.label}</strong><small>{busy ? 'Complete sign-in in the Google window.' : provider.copy}</small></span>
						{#if !busy}<span class="arrow" aria-hidden="true">→</span>{/if}
					</button>
				{/each}
			</div>
			{#if busy}<p class="sr-only" role="status">Signing in with Google. Complete sign-in in the Google window.</p>{/if}
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

		<details class="card hero-card" open={!returningUser}>
			<summary>Learn about Open-Skill</summary>
			<div class="hero-content">
				<div>
					<span class="eyebrow">Open-Skill preview</span>
					<p class="label-caps">Competitive ranking</p>
					<h2>Bring your communities into focus.</h2>
					<p class="hero-copy">Track players, report results, and keep every competition moving with one shared ranking space.</p>
				</div>
				<div class="benefits">
					<div><strong>Rankings</strong><span>Modes that update after every match.</span></div>
					<div><strong>Communities</strong><span>Open or approval-based groups.</span></div>
					<div><strong>History</strong><span>See how every result shaped ratings.</span></div>
				</div>
			</div>
		</details>
	</div>
</div>

<style>
	.page-wrap { min-height: calc(100vh - 65px); display:grid; justify-items:center; align-items:start; padding:20px 16px 40px; }
	.shell { width:100%; max-width:1040px; min-width:0; display:grid; grid-template-columns:minmax(0,1fr); gap:24px; }
	.card { background:var(--theme-surface); border:1px solid var(--theme-border); border-radius:28px; box-shadow:0 24px 64px rgba(0,0,0,.55); }
	.hero-card { min-width:0; padding:24px; align-self:center; background:linear-gradient(145deg,rgba(255,255,255,.04),rgba(255,255,255,.01)),var(--theme-surface); }
	.hero-card summary { cursor:pointer; font-size:14px; font-weight:700; color:var(--theme-text-muted); }
	.hero-card summary:focus-visible,.retry:focus-visible { outline:2px solid var(--theme-accent); outline-offset:5px; }
	.hero-content { display:flex; flex-direction:column; gap:32px; padding-top:24px; }
	.eyebrow { display:inline-flex; padding:10px 14px; border:1px solid var(--theme-border); border-radius:999px; background:rgba(255,255,255,.03); color:var(--theme-text-muted); font-size:12px; font-weight:700; margin-bottom:26px; }
	.label-caps { color:var(--theme-accent); font-size:11px; font-weight:800; letter-spacing:.12em; text-transform:uppercase; margin-bottom:10px; }
	.hero-card h2 { max-width:8ch; font-size:clamp(34px,4vw,54px); line-height:.96; letter-spacing:-.04em; margin-bottom:16px; }
	.hero-copy { max-width:42ch; color:var(--theme-text-muted); font-size:15px; line-height:25px; }
	.benefits { display:grid; grid-template-columns:1fr; gap:12px; }
	.benefits div { padding:16px; border-radius:18px; background:rgba(255,255,255,.04); border:1px solid var(--theme-border); }
	.benefits strong,.benefits span { display:block; } .benefits strong { font-size:13px; margin-bottom:6px; } .benefits span { font-size:12px; line-height:18px; color:var(--theme-text-muted); }
	.login-card { min-width:0; padding:24px; align-self:center; } .login-card h1 { font-size:28px; margin-bottom:8px; } .login-card > p:not(.label-caps) { font-size:14px; line-height:22px; margin-bottom:24px; }
	.session-status { display:flex; align-items:flex-start; gap:14px; padding:20px; margin:20px 0; border:1px solid var(--theme-accent); border-radius:18px; background:rgba(240,90,26,.08); }
	.session-status strong { font-size:16px; } .session-status p { margin:8px 0 0; font-size:14px; line-height:22px; color:var(--theme-text-muted); }
	.spinner { display:inline-block; flex:none; width:22px; height:22px; border:3px solid currentColor; border-right-color:transparent; border-radius:50%; animation:spin .8s linear infinite; }
	.session-status > .spinner { color:var(--theme-accent); margin-top:2px; }
	@keyframes spin { to { transform:rotate(360deg); } }
	@media (prefers-reduced-motion:reduce) { .spinner { animation:none; } .provider { transition:none; } }
	.retry { min-height:44px; padding:10px 18px; margin-bottom:20px; border:1px solid var(--theme-accent); border-radius:12px; background:transparent; color:var(--theme-text); font:inherit; cursor:pointer; }
	.login-card > .session-error { margin-top:20px; }
	.sr-only { position:absolute; width:1px; height:1px; padding:0; margin:-1px; overflow:hidden; clip-path:inset(50%); white-space:nowrap; }
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
	@media (min-width:761px) {
		.page-wrap { align-items:center; padding:32px 20px 48px; }
		.shell { grid-template-columns:minmax(340px,.95fr) minmax(0,1.05fr); }
		.login-card { padding:32px; } .hero-card { padding:36px; }
		.hero-content { min-height:480px; justify-content:space-between; }
		.benefits { grid-template-columns:repeat(3,minmax(0,1fr)); }
	}
	@media (max-width:480px) { .email-row { flex-direction:column; } .email-row input,.email-submit { min-height:48px; } }
</style>