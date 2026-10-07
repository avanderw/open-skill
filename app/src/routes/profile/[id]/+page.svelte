<script lang="ts">
	import { base } from '$app/paths';
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { AUTH_CHANGE_EVENT } from '$lib/auth';
	import type { Profile, Rating } from '$lib/api';
	let { data } = $props();
	const profile = data.profile as Profile;
	let handle = $state(profile.player.handle);
	let isOwner = $state(false);
	let editing = $state(false);
	let draft = $state('');
	let saving = $state(false);
	let renameError = $state('');
	onMount(async () => {
		try { isOwner = (await api.me()).playerId === profile.player.id; } catch { isOwner = false; }
	});
	function startEdit() { draft = handle; renameError = ''; editing = true; }
	async function saveName() {
		const next = draft.trim();
		if (!next) { renameError = 'Name is required.'; return; }
		if (next === handle) { editing = false; return; }
		saving = true; renameError = '';
		try { handle = (await api.renamePlayer(profile.player.id, next)).handle; editing = false; window.dispatchEvent(new Event(AUTH_CHANGE_EVENT)); }
		catch (e) { renameError = (e as Error).message; }
		finally { saving = false; }
	}
	const totalGames = profile.ratings.reduce((sum: number, rating: Rating) => sum + rating.games, 0);
	const totalWins = profile.ratings.reduce((sum: number, rating: Rating) => sum + rating.wins, 0);
	const totalLosses = profile.ratings.reduce((sum: number, rating: Rating) => sum + rating.losses, 0);
	const totalDraws = profile.ratings.reduce((sum: number, rating: Rating) => sum + rating.draws, 0);
	function avatar(name: string) { return name.split(/\s+/).slice(0, 2).map((w) => w[0]?.toUpperCase()).join(''); }
	const modeLabel: Record<string, string> = { '1v1': '1v1', team: 'Team', ffa: 'FFA' };
	const byCommunity = $derived.by(() => {
		const groups = new Map<string, { id: string; name: string; ratings: Rating[]; games: number }>();
		for (const rating of profile.ratings) {
			let group = groups.get(rating.communityId);
			if (!group) {
				group = { id: rating.communityId, name: rating.communityName || rating.communityId, ratings: [], games: 0 };
				groups.set(rating.communityId, group);
			}
			// Unplayed modes carry default ratings, so only show modes with games.
			if (rating.games > 0) {
				group.ratings.push(rating);
				group.games += rating.games;
			}
		}
		return [...groups.values()].sort((a, b) => a.name.localeCompare(b.name));
	});
</script>
<svelte:head><title>{handle} — Open-Skill</title></svelte:head>
<div class="page-wrap"><div class="page-inner"><div class="nav-row"><a class="btn btn-ghost" href={`${base}/leaderboard`}>← Leaderboard</a></div>
<section class="card header-card"><div class="header-top"><div class="header-title"><div class="profile-avatar">{avatar(handle)}</div><div><p class="label-caps">Player profile</p>{#if editing}<form class="rename" onsubmit={(e) => { e.preventDefault(); saveName(); }}><input class="text-input" aria-label="Display name" maxlength="64" bind:value={draft} disabled={saving} /><button class="btn btn-primary" type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save'}</button><button class="btn btn-ghost" type="button" disabled={saving} onclick={() => (editing = false)}>Cancel</button></form>{#if renameError}<p class="rename-error" role="alert">{renameError}</p>{/if}{:else}<h1>{handle}{#if isOwner} <button class="btn btn-ghost edit-name" type="button" onclick={startEdit}>Edit name</button>{/if}</h1>{/if}<p class="muted">{profile.player.id} · Open-Skill player</p></div></div><a class="btn btn-secondary" href={`${base}/history?playerId=${encodeURIComponent(profile.player.id)}`}>View match history →</a></div><div class="summary-grid"><div><span>Total games</span><strong>{totalGames}</strong></div><div><span>Wins</span><strong>{totalWins}</strong></div><div><span>Losses</span><strong>{totalLosses}</strong></div><div><span>Draws</span><strong>{totalDraws}</strong></div></div></section>
<section class="card section-card"><p class="label-caps">Ratings</p><h2>Ratings by community</h2>{#if profile.ratings.length === 0}<div class="empty"><h3>No ratings yet</h3><p class="muted">This player hasn't competed in a rated match.</p></div>{:else}{#each byCommunity as group (group.id)}<div class="community-group"><div class="community-head"><h3><a href={`${base}/leaderboard?communityId=${encodeURIComponent(group.id)}`}>{group.name}</a></h3><span class="muted">{group.games} games</span></div>{#if group.ratings.length === 0}<p class="group-empty muted">No games played in this community yet — report a match to get rated here.</p>{:else}<div class="ratings-grid">{#each group.ratings as rating (rating.mode)}<article class="rating-card" class:primary={rating.mode === '1v1'}><span class="stat-label">{modeLabel[rating.mode] || rating.mode}</span><strong>{Math.round(rating.rating)}</strong>{#if rating.rank}<span class="rank-chip">Rank #{rating.rank}</span>{/if}<p>{rating.games} games · uncertainty {Math.round(rating.uncertainty)}</p><small>{rating.wins} W · {rating.losses} L · {rating.draws} D</small>{#if rating.lastPlayed}<small>Last played {rating.lastPlayed}</small>{/if}</article>{/each}</div>{/if}</div>{/each}{/if}</section></div></div>
<style>
.rename{display:flex;gap:8px;flex-wrap:wrap;align-items:center;margin-bottom:6px}.rename-error{color:#f87171;font-size:13px;margin-bottom:6px}.edit-name{font-size:13px;vertical-align:middle}
.page-inner{display:flex;flex-direction:column;gap:18px}.nav-row{display:flex;gap:10px;flex-wrap:wrap}.header-card,.section-card{padding:28px}.header-top{display:flex;justify-content:space-between;align-items:flex-start;gap:18px;flex-wrap:wrap;margin-bottom:24px}.header-title{display:flex;align-items:center;gap:16px}.profile-avatar{width:76px;height:76px;border-radius:20px;display:grid;place-items:center;background:linear-gradient(135deg,#b83c08,var(--brand-primary));color:#fff;font-size:28px;font-weight:800}.label-caps{font-size:12px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:var(--theme-accent);margin-bottom:8px}.header-title h1{font-size:clamp(28px,4vw,42px);line-height:1.05;margin-bottom:6px}.summary-grid{display:grid;grid-template-columns:repeat(4,1fr);gap:12px}.summary-grid div{padding:16px;border:1px solid var(--theme-border);border-radius:14px;background:rgba(255,255,255,.03)}.summary-grid span,.summary-grid strong{display:block}.summary-grid span{font-size:11px;color:var(--brand-muted);text-transform:uppercase;letter-spacing:.08em}.summary-grid strong{font-size:28px;margin-top:8px}.section-card h2{font-size:24px;margin-bottom:18px}.ratings-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(180px,1fr));gap:14px;margin-bottom:20px}.community-group{margin-bottom:26px}.community-group:last-child{margin-bottom:0}.community-head{display:flex;align-items:baseline;justify-content:space-between;gap:12px;margin-bottom:12px}.community-head h3{font-size:16px;font-weight:800}.community-head .muted{font-size:12px}.group-empty{padding:16px 18px;border:1px dashed var(--theme-border);border-radius:15px;font-size:13px}.rating-card{padding:18px;border:1px solid var(--theme-border);border-radius:15px;background:rgba(255,255,255,.03)}.rating-card.primary{background:rgba(240,90,26,.08);border-color:rgba(240,90,26,.3)}.rating-card strong{display:block;font-size:34px;margin:8px 0}.rank-chip{display:inline-block;padding:3px 9px;border-radius:999px;background:rgba(240,90,26,.14);border:1px solid rgba(240,90,26,.5);color:#ffd9c9;font-size:11px;font-weight:800;text-transform:uppercase}.rating-card p,.rating-card small{display:block;color:var(--theme-text-muted);font-size:12px;margin-top:4px}.stat-label{font-size:11px;color:var(--brand-muted);font-weight:800;text-transform:uppercase}.empty{text-align:center;padding:48px 20px}@media(max-width:620px){.summary-grid{grid-template-columns:repeat(2,1fr)}.header-card,.section-card{padding:20px}}
</style>
