<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { loginDemo } from '$lib/auth';
	import { setDebugProfile } from '$lib/debug-profiles';

	const providers = [
		{ id: 'discord', label: 'Continue with Discord', copy: 'Use the community identity you already have.', icon: 'D' },
		{ id: 'gmail', label: 'Continue with Gmail', copy: 'Use a familiar account to get started.', icon: 'G' }
	];

	let email = $state('');

	function login() {
		loginDemo();
		const returnTo = $page.url.searchParams.get('returnTo') || '/communities';
		goto(returnTo.startsWith('/') && !returnTo.startsWith('//') && returnTo !== '/' ? returnTo : '/communities');
	}

	function loginWithEmail() {
		const username = email.trim().toLowerCase();
		if (!username) return;
		setDebugProfile(username);
		login();
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
			<p class="muted">Choose a provider or enter an email to use as the temporary username. No authentication is performed yet.</p>
			<div class="provider-list">
				{#each providers as provider}
					<button class="provider provider-{provider.id}" onclick={login}>
						<span class="provider-icon">{provider.icon}</span>
						<span class="provider-copy"><strong>{provider.label}</strong><small>{provider.copy}</small></span>
						<span class="arrow">→</span>
					</button>
				{/each}
			</div>
			<form class="email-login" onsubmit={(event) => { event.preventDefault(); loginWithEmail(); }}>
				<label for="email">Email</label>
				<div class="email-row">
					<input id="email" bind:value={email} type="email" autocomplete="email" placeholder="person@example.com" required />
					<button class="email-submit" type="submit">Continue</button>
				</div>
			</form>
			<a class="back" href="/leaderboard">← Continue browsing as guest</a>
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
	.provider-discord { background:rgba(88,101,242,.18); } .provider-gmail { background:rgba(255,255,255,.92); color:#18181b; }
	.provider-icon { width:50px; height:50px; border-radius:16px; display:grid; place-items:center; flex:none; font-size:22px; font-weight:800; background:var(--theme-accent); color:#fff; } .provider-discord .provider-icon { background:#5865f2; } .provider-gmail .provider-icon { background:#ea4335; }
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