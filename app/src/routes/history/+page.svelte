<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { navigating, page } from '$app/state';
	import { base } from '$app/paths';
	import { api, type MatchResult } from '$lib/api';
	import FilterPicker from '$lib/FilterPicker.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	let removedMatches = $state<string[]>([]);
	const history = $derived(data.history.filter((match) => !removedMatches.includes(match.matchId)));
	const mode = $derived(data.mode);
	let changingFilters = $state(false);
	const loading = $derived(changingFilters || navigating.to?.route.id === '/history');
	let error = $state('');
	let pendingDelete = $state<string | null>(null);
	let busyMatch = $state<string | null>(null);
	const playerId = $derived(data.playerId);
	const communityId = $derived(data.communityId);
	const canManage = $derived(data.canManage);
	const controlsDisabled = $derived(loading || busyMatch !== null);
	const modeLabel: Record<string, string> = { '1v1': '1v1', team: 'Team', ffa: 'FFA' };

	async function filter(key: 'playerId' | 'communityId' | 'mode', value: string) {
		if (controlsDisabled) return;
		const params = new URLSearchParams(page.url.searchParams);
		if (value) params.set(key, value);
		else params.delete(key);
		changingFilters = true;
		pendingDelete = null;
		error = '';
		try {
			await goto(`${base}/history?${params}`, { noScroll: true, keepFocus: true });
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not load match history.';
		} finally {
			changingFilters = false;
		}
	}

	function historyHref(nextPlayerId: string, nextCommunityId: string) {
		const params = new URLSearchParams();
		if (nextPlayerId) params.set('playerId', nextPlayerId);
		if (nextCommunityId) params.set('communityId', nextCommunityId);
		if (mode) params.set('mode', mode);
		return `${base}/history?${params}`;
	}

	function communityName(id: string) {
		return data.communityOptions.find((option) => option.value === id)?.label || id;
	}

	async function removeMatch(matchId: string) {
		if (controlsDisabled) return;
		busyMatch = matchId;
		error = '';
		let deleted = false;
		try {
			await api.deleteMatch(matchId);
			deleted = true;
			removedMatches = [...removedMatches, matchId];
			pendingDelete = null;
			await invalidateAll();
		} catch (e) {
			const message = e instanceof Error ? e.message : 'Could not update match history.';
			error = deleted ? `Match deleted, but refreshing history failed. Reload to try again. ${message}` : message;
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
		if (value.endsWith(' 00:00:00')) {
			return new Date(`${value.slice(0, 10)}T00:00:00`).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
		}
		const parsed = new Date(value.includes('T') ? value : value.replace(' ', 'T') + 'Z');
		return Number.isNaN(parsed.getTime()) ? value : parsed.toLocaleString();
	}
</script>

<svelte:head><title>Match history — Open-Skill</title></svelte:head>
<div class="page-wrap"><div class="page-inner">
	<div class="nav">
		{#if communityId}
			<a class="btn btn-ghost" href={`${base}/leaderboard?communityId=${encodeURIComponent(communityId)}`}>← Leaderboard</a>
		{/if}
		{#if playerId}
			<a class="btn btn-ghost" href={`${base}/profile/${encodeURIComponent(playerId)}`}>← Profile</a>
		{/if}
	</div>

	<section class="card hero">
		<p class="label-caps">Competitive record</p>
		<h1>Match history</h1>
		<p class="scope">{data.player?.handle || 'All players'} · {data.community?.name || 'All communities'} · {modeLabel[mode] || 'All modes'}</p>
		{#if canManage}<p class="hint">As an owner or admin you can correct or remove a match. Community ratings are recalculated from the remaining match log.</p>{/if}
	</section>

	<section class="scope-filters" aria-label="History filters">
		<FilterPicker label="Community" value={communityId} options={data.communityOptions}
			allLabel="All communities" allowClear={!!playerId} disabled={controlsDisabled}
			emptyText="No communities with recorded ratings for this player yet."
			onchange={(value) => filter('communityId', value)} />
		<FilterPicker label="Player" value={playerId} options={data.playerOptions}
			allLabel="All players" allowClear={!!communityId} disabled={controlsDisabled}
			hint={communityId ? 'Includes current players and former match participants.' : 'Choose a community to browse its players.'}
			emptyText="No players in this community yet."
			onchange={(value) => filter('playerId', value)} />
		{#if playerId && communityId}
			<div class="clear-filters">
				<button class="btn btn-ghost" disabled={controlsDisabled} onclick={() => filter('playerId', '')}>All players</button>
				<button class="btn btn-ghost" disabled={controlsDisabled} onclick={() => filter('communityId', '')}>All communities</button>
			</div>
		{/if}
	</section>

	<section class="toolbar">
		<div class="filters" role="group" aria-label="Match mode">
			<button class:active={!mode} aria-pressed={!mode} disabled={controlsDisabled} onclick={() => filter('mode', '')}>All modes</button>
			<button class:active={mode === '1v1'} aria-pressed={mode === '1v1'} disabled={controlsDisabled} onclick={() => filter('mode', '1v1')}>1v1</button>
			<button class:active={mode === 'team'} aria-pressed={mode === 'team'} disabled={controlsDisabled} onclick={() => filter('mode', 'team')}>Team</button>
			<button class:active={mode === 'ffa'} aria-pressed={mode === 'ffa'} disabled={controlsDisabled} onclick={() => filter('mode', 'ffa')}>FFA</button>
		</div>
	</section>
	<p class="muted result-status" role="status">{loading ? 'Loading matches…' : `${history.length} ${history.length === 1 ? 'match' : 'matches'} shown · Newest first · Up to 50 matching matches`}</p>

	{#if error}<p class="alert" role="alert">{error}</p>{/if}

	<section class="card panel" aria-live="polite" aria-busy={loading}>
		{#if history.length === 0}
			<div class="empty">
				<h2>No matches for these filters</h2>
				<p class="muted">Try another player, community, or match mode.</p>
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
								{#if !communityId}<a class="community-link" href={historyHref(playerId, match.communityId)}>{communityName(match.communityId)}</a>{/if}
							</div>
							{#if canManage}
								<div class="match-actions">
									<a class="btn btn-ghost" href={`${base}/report?communityId=${encodeURIComponent(match.communityId)}&matchId=${encodeURIComponent(match.matchId)}`}>Edit</a>
									<button class="btn btn-ghost danger" disabled={controlsDisabled} onclick={() => (pendingDelete = pendingDelete === match.matchId ? null : match.matchId)}>Delete</button>
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
												<div class="participant">
													<a class="chip" class:focused={participant.playerId === playerId} href={`${base}/profile/${encodeURIComponent(participant.playerId)}`}>
														{participant.handle ?? participant.playerId}
														{#if participant.unclaimed}<span class="unclaimed-badge">Unclaimed</span>{/if}
														<b class:down={participant.ratingDelta < 0}>{participant.ratingDelta >= 0 ? '+' : ''}{Math.round(participant.ratingDelta * 10) / 10}</b>
													</a>
													{#if participant.playerId !== playerId || match.communityId !== communityId}
														<a class="player-history" href={historyHref(participant.playerId, match.communityId)}
															aria-label={`Filter history to ${participant.handle || participant.playerId} in ${communityName(match.communityId)}`}>History here</a>
													{/if}
												</div>
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
.scope{overflow-wrap:anywhere}.scope-filters{display:grid;grid-template-columns:minmax(0,1fr);gap:10px}.clear-filters{display:flex;flex-wrap:wrap;gap:8px;grid-column:1/-1}.clear-filters .btn{font-size:12px;min-height:44px}.result-status{font-size:12px;min-height:18px}.participant{display:flex;flex-wrap:wrap;align-items:center;gap:4px 8px;min-width:0;max-width:100%}.participant .chip{display:inline-flex;align-items:center;flex-wrap:wrap;gap:4px;overflow-wrap:anywhere;min-height:44px}.chip.focused{outline:1px solid var(--theme-accent)}.player-history{display:inline-flex;align-items:center;min-height:44px;font-size:12px;color:var(--theme-accent)}.community-link{font-size:12px;overflow-wrap:anywhere}.match-title{flex-wrap:wrap;min-width:0}.side-body{min-width:0;overflow-wrap:anywhere}.filters button:disabled{opacity:.5;cursor:default}@media(min-width:540px){.scope-filters{grid-template-columns:repeat(2,minmax(0,1fr))}}
.page-inner{display:flex;flex-direction:column;gap:18px;max-width:980px}.nav{display:flex;gap:10px;flex-wrap:wrap}.hero,.panel{padding:28px}.hero h1{font-size:clamp(32px,5vw,48px);line-height:1.05;margin-bottom:10px}.hero p:not(.label-caps){color:var(--theme-text-muted)}.hero .hint{margin-top:10px;font-size:13px}.label-caps{font-size:12px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:var(--theme-accent);margin-bottom:8px}.toolbar{display:flex;justify-content:space-between;align-items:center;gap:12px}.filters{display:flex;gap:8px;flex-wrap:wrap}.filters button{padding:8px 12px;border:1px solid var(--theme-border);border-radius:999px;background:rgba(255,255,255,.04);color:var(--theme-text-muted);cursor:pointer}.filters button.active{background:rgba(240,90,26,.14);border-color:var(--theme-accent);color:#ffd9c9}.match{padding:18px 0;border-bottom:1px solid var(--theme-border)}.match:first-child{padding-top:0}.match:last-child{border-bottom:0;padding-bottom:0}.match-head{display:flex;justify-content:space-between;align-items:center;gap:12px;flex-wrap:wrap;margin-bottom:12px}.match-title{display:flex;align-items:center;gap:10px}.mode-chip{padding:4px 10px;border-radius:999px;border:1px solid rgba(240,90,26,.5);background:rgba(240,90,26,.16);color:#ffd9c9;font-size:11px;font-weight:800;text-transform:uppercase}.match-actions{display:flex;gap:8px}.match-actions .btn{padding:7px 14px;font-size:12px}.danger{color:var(--danger)}.confirm{margin-bottom:12px;padding:14px 16px;border:1px solid var(--danger-border);border-radius:12px;background:var(--danger-bg);font-size:13px}.confirm-actions{display:flex;gap:8px;justify-content:flex-end;margin-top:12px}.btn-danger{background:var(--danger);color:#1a0000;border:1px solid var(--danger)}.sides{list-style:none;display:flex;flex-direction:column;gap:8px}.sides li{display:flex;align-items:center;gap:14px;padding:10px 12px;border:1px solid var(--theme-border);border-radius:12px;background:rgba(255,255,255,.02)}.sides li.won{border-color:rgba(110,231,183,.4);background:rgba(6,118,71,.14)}.place{min-width:24px;font-weight:800;color:var(--brand-muted)}.sides li.won .place{color:var(--ok)}.side-body{display:flex;flex-direction:column;gap:6px}.side-players{display:flex;gap:8px;flex-wrap:wrap}.chip{padding:5px 10px;border-radius:8px;background:rgba(255,255,255,.06);font-size:12px;color:var(--theme-text)}.chip b{color:var(--ok);margin-left:4px}.chip b.down{color:var(--danger)}.empty{text-align:center;padding:48px 20px}.empty .btn{margin-top:16px}.alert{padding:12px;border-radius:10px;color:var(--danger);background:var(--danger-bg);border:1px solid var(--danger-border)}
@media(max-width:539px){.hero,.panel{padding:20px}.empty{padding:28px 0}.sides li{gap:8px}}
</style>
