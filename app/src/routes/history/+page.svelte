<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { base } from '$app/paths';
	import { api, type Community, type MatchResult, type Mode } from '$lib/api';

	let { data } = $props();
	let history = $state(data.history as MatchResult[]);
	let mode = $state<Mode | ''>('');
	let loading = $state(false);
	let error = $state('');
	let pendingDelete = $state<string | null>(null);
	let busyMatch = $state<string | null>(null);
	const playerId = data.playerId as string;
	const communityId = data.communityId as string;
	const community = data.community as Community | null;
	const canManage = data.canManage as boolean;
	const modeLabel: Record<string, string> = { '1v1': '1v1', team: 'Team', ffa: 'FFA' };

	async function filter(next: Mode | '') {
		mode = next;
		loading = true;
		error = '';
		try {
			history = await api.history(playerId, { communityId: communityId || undefined, mode: mode || undefined, limit: 50 });
		} catch (e) {
			error = (e as Error).message;
		} finally {
			loading = false;
		}
	}

	async function removeMatch(matchId: string) {
		busyMatch = matchId;
		error = '';
		try {
			await api.deleteMatch(matchId);
			history = history.filter((match) => match.matchId !== matchId);
			pendingDelete = null;
			await invalidateAll();
		} catch (e) {
			error = (e as Error).message;
		} finally {
			busyMatch = null;
		}
	}

	// Matches are stored as scores/finishes; group them back into ranked sides.
	function sides(match: MatchResult) {
		const groups = new Map<string, { label: string; players: MatchResult['participants']; order: number }>();
		for (const participant of match.participants) {
			const key = participant.team || participant.playerId;
			const order = participant.finish >= 0 ? participant.finish : -participant.score;
			const existing = groups.get(key);
			if (existing) {
				existing.players.push(participant);
				existing.order = Math.min(existing.order, order);
			} else {
				groups.set(key, { label: participant.team || participant.handle || participant.playerId, players: [participant], order });
			}
		}
		const ordered = [...groups.values()].sort((a, b) => a.order - b.order);
		return ordered.map((group, index) => ({ ...group, place: index + 1, won: group.order === ordered[0].order }));
	}

	function when(value: string) {
		const parsed = new Date(value.includes('T') ? value : value.replace(' ', 'T') + 'Z');
		return Number.isNaN(parsed.getTime()) ? value : parsed.toLocaleString();
	}
</script>

<svelte:head><title>Match history — Open-Skill</title></svelte:head>
<div class="page-wrap"><div class="page-inner">
	<div class="nav">
		{#if communityId}
			<a class="btn btn-ghost" href={`${base}/leaderboard?communityId=${encodeURIComponent(communityId)}`}>← Leaderboard</a>
		{:else}
			<a class="btn btn-ghost" href={`${base}/profile/${encodeURIComponent(playerId)}`}>← Profile</a>
		{/if}
	</div>

	<section class="card hero">
		<p class="label-caps">Competitive record</p>
		<h1>Match history</h1>
		<p>{community ? `Every match reported in ${community.name}, newest first.` : 'Chronological results and rating changes for this player.'}</p>
		{#if canManage}<p class="hint">As an owner or admin you can correct or remove a match. Community ratings are recalculated from the remaining match log.</p>{/if}
	</section>

	<section class="toolbar">
		<div class="filters">
			<button class:active={!mode} onclick={() => filter('')}>All</button>
			<button class:active={mode === '1v1'} onclick={() => filter('1v1')}>1v1</button>
			<button class:active={mode === 'team'} onclick={() => filter('team')}>Team</button>
			<button class:active={mode === 'ffa'} onclick={() => filter('ffa')}>FFA</button>
		</div>
		{#if loading}<span class="muted">Loading…</span>{/if}
	</section>

	{#if error}<p class="alert" role="alert">{error}</p>{/if}

	<section class="card panel" aria-live="polite">
		{#if history.length === 0}
			<div class="empty">
				<h2>No matches yet</h2>
				<p class="muted">Report a match to start building history.</p>
				{#if communityId}<a class="btn btn-primary" href={`${base}/report?communityId=${encodeURIComponent(communityId)}`}>Report match</a>{/if}
			</div>
		{:else}
			<div class="history">
				{#each history as match (match.matchId)}
					<article class="match">
						<div class="match-head">
							<div class="match-title">
								<span class="mode-chip">{modeLabel[match.mode] ?? match.mode}</span>
								<span class="muted">{when(match.playedAt)}</span>
							</div>
							{#if canManage}
								<div class="match-actions">
									<a class="btn btn-ghost" href={`${base}/report?communityId=${encodeURIComponent(match.communityId)}&matchId=${encodeURIComponent(match.matchId)}`}>Edit</a>
									<button class="btn btn-ghost danger" onclick={() => (pendingDelete = pendingDelete === match.matchId ? null : match.matchId)}>Delete</button>
								</div>
							{/if}
						</div>

						{#if pendingDelete === match.matchId}
							<div class="confirm">
								<p>Delete this match? Every rating in the community will be recalculated without it.</p>
								<div class="confirm-actions">
									<button class="btn btn-ghost" onclick={() => (pendingDelete = null)}>Keep it</button>
									<button class="btn btn-danger" disabled={busyMatch === match.matchId} onclick={() => removeMatch(match.matchId)}>{busyMatch === match.matchId ? 'Deleting…' : 'Delete match'}</button>
								</div>
							</div>
						{/if}

						<ol class="sides">
							{#each sides(match) as side}
								<li class:won={side.won}>
									<span class="place">{side.place}</span>
									<div class="side-body">
										<strong>{side.label}</strong>
										<div class="side-players">
											{#each side.players as participant}
												<a class="chip" href={`${base}/profile/${encodeURIComponent(participant.playerId)}`}>
													{participant.handle ?? participant.playerId}
													<b class:down={participant.ratingDelta < 0}>{participant.ratingDelta >= 0 ? '+' : ''}{Math.round(participant.ratingDelta * 10) / 10}</b>
												</a>
											{/each}
										</div>
									</div>
								</li>
							{/each}
						</ol>
					</article>
				{/each}
			</div>
		{/if}
	</section>
</div></div>

<style>
.page-inner{display:flex;flex-direction:column;gap:18px;max-width:980px}.nav{display:flex;gap:10px;flex-wrap:wrap}.hero,.panel{padding:28px}.hero h1{font-size:clamp(32px,5vw,48px);line-height:1.05;margin-bottom:10px}.hero p:not(.label-caps){color:var(--theme-text-muted)}.hero .hint{margin-top:10px;font-size:13px}.label-caps{font-size:12px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:var(--theme-accent);margin-bottom:8px}.toolbar{display:flex;justify-content:space-between;align-items:center;gap:12px}.filters{display:flex;gap:8px;flex-wrap:wrap}.filters button{padding:8px 12px;border:1px solid var(--theme-border);border-radius:999px;background:rgba(255,255,255,.04);color:var(--theme-text-muted);cursor:pointer}.filters button.active{background:rgba(240,90,26,.14);border-color:var(--theme-accent);color:#ffd9c9}.match{padding:18px 0;border-bottom:1px solid var(--theme-border)}.match:first-child{padding-top:0}.match:last-child{border-bottom:0;padding-bottom:0}.match-head{display:flex;justify-content:space-between;align-items:center;gap:12px;flex-wrap:wrap;margin-bottom:12px}.match-title{display:flex;align-items:center;gap:10px}.mode-chip{padding:4px 10px;border-radius:999px;border:1px solid rgba(240,90,26,.5);background:rgba(240,90,26,.16);color:#ffd9c9;font-size:11px;font-weight:800;text-transform:uppercase}.match-actions{display:flex;gap:8px}.match-actions .btn{padding:7px 14px;font-size:12px}.danger{color:var(--danger)}.confirm{margin-bottom:12px;padding:14px 16px;border:1px solid var(--danger-border);border-radius:12px;background:var(--danger-bg);font-size:13px}.confirm-actions{display:flex;gap:8px;justify-content:flex-end;margin-top:12px}.btn-danger{background:var(--danger);color:#1a0000;border:1px solid var(--danger)}.sides{list-style:none;display:flex;flex-direction:column;gap:8px}.sides li{display:flex;align-items:center;gap:14px;padding:10px 12px;border:1px solid var(--theme-border);border-radius:12px;background:rgba(255,255,255,.02)}.sides li.won{border-color:rgba(110,231,183,.4);background:rgba(6,118,71,.14)}.place{min-width:24px;font-weight:800;color:var(--brand-muted)}.sides li.won .place{color:var(--ok)}.side-body{display:flex;flex-direction:column;gap:6px}.side-players{display:flex;gap:8px;flex-wrap:wrap}.chip{padding:5px 10px;border-radius:8px;background:rgba(255,255,255,.06);font-size:12px;color:var(--theme-text)}.chip b{color:var(--ok);margin-left:4px}.chip b.down{color:var(--danger)}.empty{text-align:center;padding:48px 20px}.empty .btn{margin-top:16px}.alert{padding:12px;border-radius:10px;color:var(--danger);background:var(--danger-bg);border:1px solid var(--danger-border)}
</style>
