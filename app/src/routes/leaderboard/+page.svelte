<script lang="ts">
	import { page } from '$app/stores';
	import { goto, invalidateAll, replaceState } from '$app/navigation';
	import type { LeaderboardRow, Mode } from '$lib/api';
	import { api } from '$lib/api';
	import { AUTH_CHANGE_EVENT, authReady, isAuthenticated } from '$lib/auth';
	import { onMount } from 'svelte';

	let { data } = $props();
	const modes: { id: Mode; label: string }[] = [{ id: '1v1', label: '1v1' }, { id: 'team', label: 'Team' }, { id: 'ffa', label: 'FFA' }];	let search = $state('');
	let sort = $state<'rating' | 'games'>('rating');
	let joining = $state(false);
	let toast = $state<{ text: string; tone: 'ok' | 'error' } | null>(null);
	let toastTimer: ReturnType<typeof setTimeout> | undefined;
	let authenticated = $state(false);
	type ReportSummary = {
		communityId: string;
		mode: Mode;
		standings: { place: string; side: string; won: boolean }[];
		deltas: { playerId: string; handle: string; delta: number; newRating: number }[];
	};
	let lastReport = $state<ReportSummary | null>(null);
	const rows = $derived((data.rows as LeaderboardRow[]) ?? []);
	const activeMode = $derived(($page.url.searchParams.get('mode') as Mode) || '1v1');
	const communityId = $derived(data.communityId as string | null);
	const community = $derived(data.community);
	// The API caps the default view, so the link only helps when members are missing.
	const truncated = $derived(!data.full && (community?.members ?? 0) > rows.length);
	// Only offer modes that have at least one reported match.
	const visibleModes = $derived.by(() => {
		const available = new Set((data.availableModes as Mode[]) ?? []);
		return modes.filter((mode) => available.has(mode.id));
	});
	onMount(() => {
		const sync = () => (authenticated = isAuthenticated());
		authReady().then(sync);
		window.addEventListener(AUTH_CHANGE_EVENT, sync);
		const joined = $page.url.searchParams.get('joined');
		if (joined) {
			showToast(joined === 'pending' ? `Join request sent to ${community?.name ?? 'the community'}.` : `You joined ${community?.name ?? 'the community'}.`);
			const clean = new URL($page.url);
			clean.searchParams.delete('joined');
			replaceState(`${clean.pathname}${clean.search}`, {});
		}
		// One-shot handoff from the report page so the reporter sees what was captured.
		const stored = sessionStorage.getItem('open-skill.last-report');
		if (stored) {
			sessionStorage.removeItem('open-skill.last-report');
			try {
				const parsed = JSON.parse(stored) as ReportSummary;
				if (parsed.communityId === communityId) lastReport = parsed;
			} catch {
				// Ignore a malformed handoff.
			}
		}
		return () => window.removeEventListener(AUTH_CHANGE_EVENT, sync);
	});
	const filtered = $derived.by(() => {
		const q = search.trim().toLowerCase();
		const out = q ? rows.filter((r) => r.handle.toLowerCase().includes(q)) : [...rows];
		if (sort === 'games') out.sort((a, b) => b.games - a.games); else out.sort((a, b) => a.rank - b.rank);
		return out;
	});
	function avatar(h: string) { return h.split(/\s+/).slice(0, 2).map((w) => w[0]?.toUpperCase()).join(''); }
	// Sub-0.1 movement reads as noise, so it shows as no trend.
	function trendOf(row: LeaderboardRow) { return Math.round((row.trend ?? 0) * 10) / 10; }
	function modeLabel(id: Mode) { return modes.find((mode) => mode.id === id)?.label ?? id; }
	function switchMode(mode: Mode) { goto(`/leaderboard?mode=${mode}${communityId ? `&communityId=${encodeURIComponent(communityId)}` : ''}`); }
	function fullLink() { return `/leaderboard?mode=${activeMode}${communityId ? `&communityId=${encodeURIComponent(communityId)}` : ''}&full=1`; }
	function showToast(text: string, tone: 'ok' | 'error' = 'ok') {
		clearTimeout(toastTimer);
		toast = { text, tone };
		toastTimer = setTimeout(() => (toast = null), 4000);
	}
	async function joinCommunity() {
		if (!community) return;
		joining = true;
		try {
			const result = await api.joinCommunity(community.code);
			await invalidateAll();
			showToast(result.status === 'pending' ? `Join request sent to ${community.name}.` : `You joined ${community.name}.`);
		} catch (error) {
			showToast((error as Error).message, 'error');
		} finally {
			joining = false;
		}
	}
	async function shareCommunity() {
		if (!community) return;
		const url = `${window.location.origin}/communities?code=${encodeURIComponent(community.code)}`;
		const text = `Join ${community.name} on Open-Skill — use code ${community.code}`;
		if (navigator.share) {
			try {
				await navigator.share({ title: community.name, text, url });
			} catch {
				// User cancelled the share sheet; nothing to do.
			}
			return;
		}
		try {
			await navigator.clipboard.writeText(url);
			showToast('Invite link copied!');
		} catch {
			showToast(url);
		}
	}
</script>

<svelte:head><title>Leaderboard — Open-Skill</title></svelte:head>
<div class="page-wrap"><div class="page-inner">
	<div class="nav-row"><a class="btn btn-ghost" href="/communities">← My Communities</a></div>
	<section class="card header-card"><div><p class="label-caps">{community ? 'Community leaderboard' : 'Community rankings'}</p><h1>{community?.name ?? 'Leaderboard'}</h1><p>{community ? `${community.members} member${community.members === 1 ? '' : 's'} · ${community.mode === 'approval' ? 'Approval required' : 'Open joining'} · ${community.draws ? 'Draws allowed' : 'No draws'} · Invite code ${community.code}` : 'Browse the top players across each ruleset. Sign in to track your own position and report matches.'}</p></div><div class="header-actions">{#if community}{#if community.membership === 'owner'}<a class="btn btn-secondary" href={`/communities/${encodeURIComponent(community.id)}/settings`}>Settings</a>{/if}{#if community.membership === 'non_member'}<button class="btn btn-primary" disabled={joining} onclick={joinCommunity}>{joining ? 'Joining…' : community.mode === 'approval' ? 'Request to join' : 'Join community'}</button>{:else if community.membership === 'pending'}<span class="status">Join request pending</span>{/if}<a class="btn btn-primary" href={`/report?communityId=${encodeURIComponent(community.id)}`}>Report match</a><a class="btn btn-ghost" href={`/team?communityId=${encodeURIComponent(community.id)}`}>Team builder</a>{#if community.membership === 'owner' || community.membership === 'member'}<button class="btn btn-ghost" onclick={shareCommunity}>Share invite</button>{/if}{:else if !authenticated}<a class="btn btn-ghost" href="/login?returnTo=%2Fleaderboard">Sign in to compete</a>{/if}</div></section>
	{#if lastReport}
		<section class="card report-card" aria-live="polite">
			<div class="report-head">
				<div>
					<p class="label-caps">Match reported</p>
					<h2>{modeLabel(lastReport.mode)} result recorded</h2>
				</div>
				<button class="btn btn-ghost" onclick={() => (lastReport = null)}>Dismiss</button>
			</div>
			<div class="report-grid">
				<div>
					<p class="report-subtitle">Finishing order</p>
					<ol class="report-standings">{#each lastReport.standings as standing}<li class:won={standing.won}><span class="place">{standing.place}</span><span>{standing.side}</span></li>{/each}</ol>
				</div>
				<div>
					<p class="report-subtitle">Rating changes</p>
					<div class="report-deltas">{#each lastReport.deltas as delta}<div class="delta-row"><span>{delta.handle}</span><span class="delta" class:up={delta.delta >= 0}>{delta.delta >= 0 ? '+' : ''}{Math.round(delta.delta * 10) / 10}</span><span class="muted">{Math.round(delta.newRating)}</span></div>{/each}</div>
				</div>
			</div>
		</section>
	{/if}
	<section class="controls"><div class="filters"><input class="text-input search" aria-label="Search player" placeholder="Search player…" bind:value={search} /><select class="sort" aria-label="Sort leaderboard" bind:value={sort}><option value="rating">Sort: Rating</option><option value="games">Sort: Most matches</option></select></div></section>
	<section class="card table-card"><div class="table-meta"><span>{filtered.length} players</span>{#if visibleModes.length}<div class="mode-tabs" role="tablist" aria-label="Game mode">{#each visibleModes as mode}<button class:active={activeMode === mode.id} class="mode-tab" role="tab" aria-selected={activeMode === mode.id} onclick={() => switchMode(mode.id)}>{mode.label}</button>{/each}</div>{/if}<div class="meta-links"><a href={`/history?communityId=${encodeURIComponent(communityId ?? '')}`}>Match history</a>{#if truncated}<a href={fullLink()}>See full leaderboard →</a>{/if}</div></div>{#if filtered.length === 0}<div class="empty"><div class="empty-icon">⌁</div><h2>No players found</h2><p class="muted">{communityId ? 'This community has no rated games yet.' : 'Try a different search or browse another ruleset.'}</p>{#if communityId}<a class="btn btn-primary" href="/communities">Back to communities</a>{/if}</div>{:else}<div class="table-wrap"><table><thead><tr><th>Rank</th><th>Player</th><th class="center">Rating</th><th class="center">Trend</th><th class="center">Matches</th></tr></thead><tbody>{#each filtered as row}<tr class:row-you={row.isYou}><td class="rank">{row.rank}</td><td><a class="player" href={`/profile/${row.playerId}`}><span class="avatar">{avatar(row.handle)}</span><span class="player-name">{row.handle}{#if row.isYou}<span class="you-badge">You</span>{/if}{#if row.role === 'owner' || row.role === 'admin'}<span class="admin-badge">{row.role === 'owner' ? 'Owner' : 'Admin'}</span>{/if}</span></a></td><td class="center rating">{Math.round(row.rating)}</td><td class="center trend" title={`Rating change over the last ${row.games === 0 ? 0 : Math.min(row.games, 3)} match${row.games === 1 ? '' : 'es'}`}>{#if trendOf(row) > 0}<span class="trend-up">▲ {trendOf(row)}</span>{:else if trendOf(row) < 0}<span class="trend-down">▼ {Math.abs(trendOf(row))}</span>{:else}—{/if}</td><td class="center matches">{row.games}</td></tr>{/each}</tbody></table></div>{/if}</section>
</div></div>

{#if toast}<div class="toast" class:toast-error={toast.tone === 'error'} role="status" aria-live="polite">{toast.text}</div>{/if}

<style>
	.page-inner{display:flex;flex-direction:column;gap:16px}.nav-row{display:flex}.nav-row .btn{padding:8px 14px;font-size:12px}.header-card{padding:26px 28px;display:flex;align-items:center;justify-content:space-between;gap:20px;flex-wrap:wrap}.header-card h1{font-size:34px;line-height:1.1;margin-bottom:8px}.header-card p:last-child{color:var(--theme-text-muted);line-height:1.5;max-width:640px}.header-actions{display:flex;gap:8px;flex-wrap:wrap;align-items:center}.status{padding:8px 10px;border-radius:999px;background:var(--ok-bg);color:var(--ok);font-size:12px}.toast{position:fixed;left:50%;bottom:28px;transform:translateX(-50%);z-index:50;max-width:min(92vw,520px);padding:13px 20px;border-radius:999px;border:1px solid var(--ok-border);background:var(--ok-bg);backdrop-filter:blur(8px);color:var(--ok);font-size:14px;font-weight:700;box-shadow:var(--shadow-card);animation:toast-in .2s ease-out}.toast-error{border-color:var(--danger-border);background:var(--danger-bg);color:var(--danger)}@keyframes toast-in{from{opacity:0;transform:translate(-50%,12px)}to{opacity:1;transform:translate(-50%,0)}}.label-caps{font-size:12px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:var(--theme-accent);margin-bottom:8px}.controls{display:flex;justify-content:flex-end;align-items:center;gap:14px;flex-wrap:wrap}.report-card{padding:22px 24px;border-color:rgba(110,231,183,.4);background:rgba(6,118,71,.12)}.report-head{display:flex;align-items:flex-start;justify-content:space-between;gap:16px;flex-wrap:wrap;margin-bottom:16px}.report-head h2{font-size:20px}.report-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(240px,1fr));gap:20px}.report-subtitle{font-size:11px;font-weight:800;letter-spacing:.1em;text-transform:uppercase;color:var(--brand-muted);margin-bottom:8px}.report-standings{list-style:none;display:flex;flex-direction:column;gap:6px}.report-standings li{display:flex;gap:12px;align-items:center;font-size:14px}.report-standings li.won{color:var(--ok);font-weight:800}.report-standings .place{min-width:44px;font-size:12px;font-weight:800;color:var(--brand-muted)}.report-standings li.won .place{color:var(--ok)}.report-deltas{display:flex;flex-direction:column}.delta-row{display:grid;grid-template-columns:1fr auto auto;gap:12px;padding:7px 0;font-size:14px;border-top:1px solid var(--theme-border)}.delta{font-weight:800;color:var(--danger)}.delta.up{color:var(--ok)}.mode-tabs,.filters{display:flex;gap:8px;flex-wrap:wrap}.mode-tab,.sort{padding:9px 16px;border-radius:999px;border:1px solid var(--theme-border);background:rgba(255,255,255,.03);color:var(--theme-text-muted);font-size:12px;font-weight:800;cursor:pointer}.mode-tab.active{border-color:rgba(240,90,26,.5);background:rgba(240,90,26,.14);color:#ffd9c9}.search{width:220px;padding:10px 12px;border-radius:10px}.sort{border-radius:10px;padding:10px 12px}.table-card{overflow:hidden}.table-meta{display:flex;justify-content:space-between;align-items:center;gap:14px;flex-wrap:wrap;padding:14px 16px;border-bottom:1px solid var(--theme-border);font-size:13px}.meta-links{display:flex;gap:16px;align-items:center}.table-wrap{overflow-x:auto}table{width:100%;border-collapse:collapse;min-width:620px}thead tr{background:rgba(255,255,255,.04)}th,td{padding:12px 14px;text-align:left;border-top:1px solid var(--theme-border);font-size:14px}th{border-top:0;font-size:11px;color:var(--brand-muted);text-transform:uppercase;letter-spacing:.08em}.center{text-align:center}.rank{font-weight:800;color:var(--theme-accent);width:60px}.player{display:flex;align-items:center;gap:10px;min-width:200px}.avatar{width:34px;height:34px;border-radius:50%;display:grid;place-items:center;background:linear-gradient(135deg,#b83c08,var(--brand-primary));color:#fff;font-size:12px;font-weight:800}.player-name{font-weight:700;color:var(--theme-text)}.rating,.matches{font-weight:700}.trend-up{color:#6ee7b7}.trend-down{color:var(--danger)}.trend{font-weight:700;white-space:nowrap}.row-you{background:rgba(240,90,26,.09)}.you-badge{margin-left:7px;padding:2px 5px;border-radius:4px;background:var(--theme-accent);color:#fff;font-size:9px;text-transform:uppercase}.admin-badge{margin-left:7px;padding:2px 5px;border-radius:4px;border:1px solid var(--theme-border);background:rgba(255,255,255,.06);color:var(--theme-text-muted);font-size:9px;text-transform:uppercase}.empty{text-align:center;padding:64px 20px}.empty-icon{font-size:48px;color:var(--theme-accent);margin-bottom:10px}.empty h2{margin-bottom:8px}@media(max-width:640px){.header-card{align-items:flex-start}.search{width:100%}}
</style>