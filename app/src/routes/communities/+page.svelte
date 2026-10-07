<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { base } from '$app/paths';
	import type { Community } from '$lib/api';

	let { data } = $props();
	const communities = $derived((data.communities as Community[] | null | undefined) ?? []);

	let joinCode = $state('');
	let joinMsg = $state<{ ok: boolean; text: string } | null>(null);
	let joining = $state(false);

	onMount(() => {
		const shared = $page.url.searchParams.get('code');
		if (shared) joinCode = shared.toUpperCase();
	});

	async function submitJoin() {
		const code = joinCode.trim().toUpperCase();
		if (!code) {
			joinMsg = { ok: false, text: 'Enter a join code to continue.' };
			return;
		}
		joining = true;
		joinMsg = null;
		try {
			const joined = await import('$lib/api').then(({ api }) => api.joinCommunity(code));
			joinCode = '';
			// The destination leaderboard surfaces the confirmation toast.
			goto(`${base}/leaderboard?communityId=${encodeURIComponent(joined.communityId)}&joined=${joined.status === 'pending' ? 'pending' : '1'}`);
		} catch (e) {
			joinMsg = { ok: false, text: (e as Error).message };
		} finally {
			joining = false;
		}
	}

	function short(n: string) {
		return n
			.split(/\s+/)
			.slice(0, 2)
			.map((w) => w[0]?.toUpperCase())
			.join('');
	}
</script>

<div class="page-wrap">
	<div class="page-inner">
		<section class="card header-card">
			<p class="label-caps">Player space</p>
			<h1>My Communities</h1>
			<p>Browse communities you belong to, jump into a community, or use a join code to enter a new one.</p>
			<div class="header-actions">
				<a class="btn btn-primary" href={`${base}/communities/browse`}>Browse Communities</a>
				<a class="btn btn-secondary" href={`${base}/communities/start`}>Create community</a>
			</div>
		</section>

		<section class="grid">
			<div class="card panel">
				<p class="label-caps">Your communities</p>
				{#if communities.length === 0}
					<p class="muted" style="padding:16px 0;">You're not a member of any community yet.</p>
				{:else}
					<div class="communities-grid">
						{#each communities as c}
							<a href={`${base}/leaderboard?communityId=${encodeURIComponent(c.id)}`} class="community-card">
								<div class="community-head">
									<div class="community-title">
										<span class="community-avatar">{short(c.name)}</span>
										<h2>{c.name}</h2>
									</div>
									<span class="mode-chip">{c.mode === 'approval' ? 'Approval' : 'Open'}</span>
								</div>
								<div class="stats">
									<div class="stat-tile">
										<p class="stat-label">Code</p>
										<p class="stat-value">{c.code}</p>
									</div>
									<div class="stat-tile">
										<p class="stat-label">Members</p>
										<p class="stat-value">{c.members}</p>
									</div>
									<div class="stat-tile">
										<p class="stat-label">Your games</p>
										<p class="stat-value">{c.yourGames ?? 0}</p>
									</div>
								</div>
							</a>
						{/each}
					</div>
				{/if}
			</div>

			<div class="card panel join-panel">
				<p class="label-caps">Join path</p>
				<h2 style="font-size:20px; font-weight:700; margin-bottom:6px;">Enter a join code</h2>
				<p class="muted" style="font-size:13px; margin-bottom:16px;">
					Paste the code shared by a community owner to join directly.
				</p>

				<div style="display:flex; gap:10px; align-items:stretch; flex-wrap:wrap;">
					<input
						class="text-input"
						placeholder="Enter code"
						bind:value={joinCode}
						style="flex:1; min-width:200px; letter-spacing:0.1em; font-weight:700; text-transform:uppercase;"
						onkeydown={(e) => e.key === 'Enter' && submitJoin()}
					/>
					<button class="btn btn-primary" disabled={joining} onclick={submitJoin}>
						{joining ? 'Joining…' : 'Join'}
					</button>
				</div>

				{#if joinMsg}
					<p role="status" class:alert-err={!joinMsg.ok} class:alert-ok={joinMsg.ok} style="margin-top:12px;">
						{joinMsg.text}
					</p>
				{/if}
			</div>
		</section>
	</div>
</div>

<style>
	.page-inner {
		display: flex;
		flex-direction: column;
		gap: 20px;
	}
	.header-card {
		padding: 26px 28px;
	}
	.header-card h1 {
		font-size: 34px;
		font-weight: 700;
		line-height: 1.1;
		margin-bottom: 8px;
	}
	.header-card p:last-child {
		color: var(--theme-text-muted);
		line-height: 1.5;
		max-width: 640px;
	}
	.header-actions {
		margin-top: 18px;
	}
	.grid {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(280px, 340px);
		gap: 16px;
		align-items: start;
	}
	.panel {
		padding: 24px;
	}
	.communities-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
		gap: 14px;
		margin-top: 8px;
	}
	.community-card {
		display: flex;
		flex-direction: column;
		gap: 14px;
		padding: 16px;
		border-radius: 16px;
		background: rgba(255, 255, 255, 0.04);
		border: 1px solid var(--theme-border);
		min-height: 148px;
		transition: border-color 0.15s ease;
	}
	.community-card:hover {
		border-color: rgba(240, 90, 26, 0.5);
	}
	.community-head {
		display: flex;
		justify-content: space-between;
		gap: 10px;
		align-items: flex-start;
	}
	.community-title {
		display: flex;
		align-items: center;
		gap: 10px;
		min-width: 0;
	}
	.community-avatar {
		width: 40px;
		height: 40px;
		border-radius: 10px;
		background: linear-gradient(135deg, #b83c08, var(--brand-primary));
		display: inline-flex;
		align-items: center;
		justify-content: center;
		font-size: 14px;
		font-weight: 800;
		color: white;
		flex-shrink: 0;
	}
	.community-title h2 {
		font-size: 15px;
		font-weight: 700;
		color: var(--theme-text);
		line-height: 1.2;
		word-break: break-word;
	}
	.mode-chip {
		padding: 4px 8px;
		border-radius: var(--radius-pill);
		border: 1px solid var(--theme-border);
		background: rgba(255, 255, 255, 0.05);
		color: var(--theme-accent);
		font-size: 10px;
		font-weight: 800;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		white-space: nowrap;
	}
	.stats {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 8px;
		margin-top: auto;
	}
	.stat-tile {
		padding: 10px;
		border-radius: 12px;
		background: rgba(0, 0, 0, 0.2);
		border: 1px solid rgba(255, 255, 255, 0.1);
		text-align: center;
	}
	.stat-label {
		font-size: 10px;
		color: var(--brand-muted);
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		margin-bottom: 4px;
	}
	.stat-value {
		font-size: 14px;
		font-weight: 700;
	}
	.join-panel {
		align-self: stretch;
	}
	.alert-ok {
		color: var(--ok);
		background: var(--ok-bg);
		border: 1px solid var(--ok-border);
		border-radius: 14px;
		padding: 12px 14px;
		font-size: 13px;
		font-weight: 600;
	}
	.alert-err {
		color: var(--danger);
		background: var(--danger-bg);
		border: 1px solid var(--danger-border);
		border-radius: 14px;
		padding: 12px 14px;
		font-size: 13px;
		font-weight: 600;
	}
	@media (max-width: 760px) {
		.grid {
			grid-template-columns: 1fr;
		}
	}
</style>