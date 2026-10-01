<script lang="ts">
	import { page } from '$app/stores';
	import { goto, invalidateAll } from '$app/navigation';
	import { api, type LeaderboardRow, type MatchOutcome } from '$lib/api';

	let { data } = $props();
	const players = (data.players as LeaderboardRow[] | null) ?? [];
	const communityId = data.communityId as string;
	const community = data.community as import('$lib/api').Community | null;
	const editing = data.match as import('$lib/api').MatchResult | null;
	let playerSearch = $state('');
	let draggedId = $state<string | null>(null);
	let error = $state('');
	let outcome = $state<MatchOutcome | null>(null);
	let submitting = $state(false);
	let review = $state(false);

	type TeamAssignment = [string, string];

	function handoffTeams(): TeamAssignment[] {
		if (editing) return editing.participants.filter((p) => p.team).map((p) => [p.playerId, p.team] as TeamAssignment);
		try {
			const raw = $page.url.searchParams.get('teams');
			const value = raw ? JSON.parse(raw) : [];
			return Array.isArray(value) ? value.filter((entry): entry is TeamAssignment => Array.isArray(entry) && typeof entry[0] === 'string' && typeof entry[1] === 'string') : [];
		} catch {
			return [];
		}
	}

	const assignments = handoffTeams();
	let selected = $state<string[]>(editing ? editing.participants.map((p) => p.playerId) : assignments.length ? assignments.map(([id]) => id) : []);
	let teamNames = $state<string[]>(assignments.length ? [...new Set(assignments.map(([, team]) => team))] : []);
	let teams = $state<Record<string, string>>(Object.fromEntries(assignments));
	let playerWinners = $state<Record<string, boolean>>(Object.fromEntries(selected.map((id) => [id, false])));
	let winningTeam = $state<string | null>(null);
	// Optional finishing place per side, keyed by team name or player id. Empty => fall back to winner ticks.
	let places = $state<Record<string, number>>(editing ? existingPlaces(editing) : {});

	// Rebuild the place inputs from a stored match so an edit starts from what was recorded.
	function existingPlaces(match: import('$lib/api').MatchResult) {
		const out: Record<string, number> = {};
		const ranked = [...match.participants].sort((a, b) => (a.finish >= 0 ? a.finish : -a.score) - (b.finish >= 0 ? b.finish : -b.score));
		const seen: string[] = [];
		for (const participant of ranked) {
			const key = participant.team || participant.playerId;
			if (!seen.includes(key)) seen.push(key);
			out[key] = seen.indexOf(key) + 1;
		}
		return out;
	}

	const sideKeys = $derived(teamNames.length ? teamNames : selected);
	const usingPlaces = $derived(sideKeys.some((key) => (places[key] ?? 0) > 0));
	const visiblePlayers = $derived(players.filter((player) => player.handle.toLowerCase().includes(playerSearch.trim().toLowerCase())));
	const leaderboardHref = communityId ? `/leaderboard?communityId=${encodeURIComponent(communityId)}` : '/leaderboard';

	function playerName(id: string) { return players.find((player) => player.playerId === id)?.handle ?? 'Unknown player'; }
	function playerRating(id: string) { return Math.round(players.find((player) => player.playerId === id)?.rating ?? 0); }
	function playerAvatar(id: string) { return playerName(id).split(/\s+/).map((word) => word[0] ?? '').slice(0, 2).join('').toUpperCase(); }
	function avatar(name: string) { return name.split(/\s+/).slice(0, 2).map((word) => word[0]?.toUpperCase()).join(''); }
	function friendlyMode() { return teamNames.length ? `${teamNames.length} teams` : selected.length > 2 ? 'Free-for-all' : '1v1'; }

	function toggle(id: string) {
		if (selected.includes(id)) {
			selected = selected.filter((playerId) => playerId !== id);
			const { [id]: _removedWinner, ...nextWinners } = playerWinners;
			const { [id]: _removedTeam, ...nextTeams } = teams;
			const { [id]: _removedPlace, ...nextPlaces } = places;
			playerWinners = nextWinners;
			teams = nextTeams;
			places = nextPlaces;
		} else {
			selected = [...selected, id];
			playerWinners = { ...playerWinners, [id]: false };
			if (teamNames.length) teams = { ...teams, [id]: leastPopulatedTeam() };
		}
		review = false;
	}

	function setPlayerWinner(id: string, checked: boolean) {
		playerWinners = { ...playerWinners, [id]: checked };
		review = false;
	}

	function setTeamWinner(team: string, checked: boolean) {
		winningTeam = checked ? team : null;
		review = false;
	}

	function playersForTeam(team: string) {
		return selected.filter((id) => teams[id] === team);
	}

	function leastPopulatedTeam() {
		return teamNames.reduce((least, team) => playersForTeam(team).length < playersForTeam(least).length ? team : least, teamNames[0] ?? 'Team 1');
	}

	function addTeam() {
		places = {};
		if (teamNames.length === 0) {
			const nextTeams = ['Team 1', 'Team 2'];
			teamNames = nextTeams;
			teams = Object.fromEntries(selected.map((id, index) => [id, nextTeams[index % nextTeams.length]]));
			playerWinners = Object.fromEntries(selected.map((id) => [id, false]));
			winningTeam = null;
		} else {
			teamNames = [...teamNames, `Team ${teamNames.length + 1}`];
		}
		review = false;
	}

	function assignPlayerToTeam(id: string, team: string) {
		teams = { ...teams, [id]: team };
		review = false;
	}

	function dropOnTeam(team: string) {
		if (draggedId) assignPlayerToTeam(draggedId, team);
		draggedId = null;
	}

	function setPlace(key: string, value: string) {
		const place = Number(value);
		places = { ...places, [key]: Number.isFinite(place) && place > 0 ? place : 0 };
		review = false;
	}

	function winnerSides() {
		if (usingPlaces) {
			const best = Math.min(...sideKeys.map((key) => places[key] || Number.MAX_SAFE_INTEGER));
			return new Set(sideKeys.filter((key) => places[key] === best));
		}
		if (teamNames.length) return winningTeam ? new Set([winningTeam]) : new Set<string>();
		return new Set(selected.filter((id) => playerWinners[id]));
	}

	function participantScore(id: string) {
		const sides = winnerSides();
		if (sides.size === 0) return 0;
		return sides.has(teamNames.length ? teams[id] : id) ? 1 : 0;
	}

	function matchResultLabel() {
		const sides = winnerSides();
		if (sides.size === 0) return 'Draw';
		if (usingPlaces) return `${[...sides].map((key) => (teamNames.length ? key : playerName(key))).join(', ')} wins`;
		if (!teamNames.length) return selected.length > 2 ? `${sides.size} winner${sides.size === 1 ? '' : 's'}` : `${playerName([...sides][0])} wins`;
		return `${[...sides].join(', ')} wins`;
	}

	type Standing = { place: string; side: string; players: string[]; won: boolean };

	function ordinal(n: number) {
		const suffix = n === 1 ? 'st' : n === 2 ? 'nd' : n === 3 ? 'rd' : 'th';
		return `${n}${suffix}`;
	}

	// With explicit places the reported order is exact; otherwise the form only
	// captures winners, so everyone else shares the next place.
	function standings(): Standing[] {
		const sides = winnerSides();
		const groups: Standing[] = teamNames.length
			? teamNames.map((team) => ({ place: '', side: team, players: playersForTeam(team), won: sides.has(team) }))
			: selected.map((id) => ({ place: '', side: playerName(id), players: [id], won: sides.has(id) }));
		if (usingPlaces) {
			const keyed = groups.map((group, index) => ({ group, key: sideKeys[index] }));
			return keyed
				.sort((a, b) => (places[a.key] || Infinity) - (places[b.key] || Infinity))
				.map(({ group, key }) => ({
					...group,
					place: places[key] ? ordinal(places[key]) + (keyed.filter((other) => places[other.key] === places[key]).length > 1 ? '=' : '') : 'Unplaced'
				}));
		}
		const winners = groups.filter((group) => group.won).length;
		const losers = groups.length - winners;
		return [...groups]
			.sort((a, b) => Number(b.won) - Number(a.won))
			.map((group) => ({
				...group,
				place: winners === 0 ? 'Tied' : group.won ? ordinal(1) + (winners > 1 ? '=' : '') : ordinal(winners + 1) + (losers > 1 ? '=' : '')
			}));
	}

	async function submit() {
		if (!communityId) { error = 'Select a community first.'; return; }
		if (selected.length < 2) { error = 'Select at least two players.'; return; }
		if (teamNames.length && teamNames.some((team) => playersForTeam(team).length === 0)) { error = 'Each team needs at least one participant.'; return; }
		if (usingPlaces && sideKeys.some((key) => !places[key])) { error = `Give every ${teamNames.length ? 'team' : 'player'} a finishing place, or clear them all to report a simple win.`; return; }
		if (!usingPlaces && winnerSides().size === 0 && community?.draws === false) { error = 'This community does not allow draws. Select a winner before submitting.'; return; }
		submitting = true;
		error = '';
		outcome = null;
		try {
			const participants = selected.map((id) => {
				const key = teamNames.length ? teams[id] : id;
				const teamSize = teamNames.length ? playersForTeam(teams[id]).length : 1;
				return {
					playerId: id,
					team: teamNames.length ? teams[id] : '',
					// Team scores are summed server-side, so split a side's score across its players.
					score: usingPlaces ? (sideKeys.length - places[key] + 1) / teamSize : participantScore(id),
					finish: !teamNames.length && selected.length > 2
						? (usingPlaces ? places[id] - 1 : playerWinners[id] ? 0 : winnerSides().size === 0 ? 0 : 1)
						: -1
				};
			});
			outcome = editing
				? await api.updateMatch(editing.matchId, { participants })
				: await api.reportMatch({ communityId, participants });
			sessionStorage.setItem('open-skill.last-report', JSON.stringify({
				communityId,
				mode: outcome.mode,
				standings: standings().map((standing) => ({ place: standing.place, side: standing.side, won: standing.won })),
				deltas: outcome.deltas.map((delta) => ({ ...delta, handle: playerName(delta.playerId) }))
			}));
			await invalidateAll();
			await goto(`/leaderboard?communityId=${encodeURIComponent(communityId)}&mode=${outcome.mode}`);
		} catch (e) { error = (e as Error).message; } finally { submitting = false; }
	}
</script>

<div class="page-wrap">
	<div class="page-inner">
		<div class="nav-row">
			<a class="btn btn-ghost" href={editing ? `/history?communityId=${encodeURIComponent(communityId)}` : leaderboardHref} style="padding:8px 14px; font-size:12px;">&larr; {editing ? 'Match history' : 'Leaderboard'}</a>
		</div>

		<section class="card hero">
			<p class="label-caps">Community operations</p>
			<h1>{editing ? 'Edit Match' : 'Report Match'}</h1>
			<p>{editing ? 'Correct the participants or the finishing order. Saving replays every match in this community so all ratings stay consistent.' : 'Select participants, add teams when needed, drag players into their sides, then mark the winning player or team. Leaving winners unticked reports a draw when allowed.'}</p>
		</section>

		<section class="card form-shell">
			<div class="community-banner"><span class="community-avatar">{avatar(community?.name ?? 'Community')}</span><div><p class="label-caps">Community</p><strong>{community?.name ?? communityId}</strong><p class="hint">{community?.draws ? 'Draws allowed' : 'No draws'} · {players.length} players on the roster</p></div><a class="btn btn-ghost" href="/communities">Change</a></div>
			<div>
				<p class="section-title">Participants</p>
				<p class="hint">Select players from the roster. Add teams when the match was played by sides, then drag participants between teams.</p>
				<p class="hint">Set a finishing place to record the exact order. Leave places blank to just tick the winner.</p>
				<div class="team-actions">
					<button class="btn btn-secondary" type="button" onclick={addTeam}>{teamNames.length ? 'Add team' : 'Add teams'}</button>
					<span class="hint">{teamNames.length ? `${teamNames.length} teams configured` : 'No teams: two players is 1v1, three or more is free-for-all.'}</span>
				</div>

				<div class="player-picker">
					<input class="text-input player-search" placeholder="Search players by name…" aria-label="Search players" bind:value={playerSearch} />
					{#each visiblePlayers as p}
						<label class="player-option" class:sel={selected.includes(p.playerId)}>
							<input type="checkbox" checked={selected.includes(p.playerId)} onchange={() => toggle(p.playerId)} />
							<span class="avatar">{avatar(p.handle)}</span>
							<span class="opt-name">{p.handle}</span>
							<span class="opt-rating muted">{Math.round(p.rating)}</span>
							{#if selected.includes(p.playerId) && teamNames.length === 0}
								<span class="place-input" onclick={(event) => event.stopPropagation()}>
									<input type="number" min="1" max={selected.length} placeholder="–" aria-label={`Finishing place for ${p.handle}`} value={places[p.playerId] || ''} oninput={(event) => setPlace(p.playerId, (event.target as HTMLInputElement).value)} />
									Place
								</span>
								{#if !usingPlaces}
									<span class="winner-toggle" onclick={(event) => event.stopPropagation()}>
										<input type="checkbox" checked={playerWinners[p.playerId] ?? false} onchange={(event) => setPlayerWinner(p.playerId, (event.target as HTMLInputElement).checked)} />
										Winner
									</span>
								{/if}
							{/if}
						</label>
					{/each}
				</div>

				{#if teamNames.length}
					<div class="team-grid">
						{#each teamNames as team}
							<section class="team-card" ondragover={(event) => event.preventDefault()} ondrop={() => dropOnTeam(team)}>
								<div class="team-head">
									<strong>{team}</strong>
									<span class="place-input">
										<input type="number" min="1" max={teamNames.length} placeholder="–" aria-label={`Finishing place for ${team}`} value={places[team] || ''} oninput={(event) => setPlace(team, (event.target as HTMLInputElement).value)} />
										Place
									</span>
									{#if !usingPlaces}
										<label class="winner-toggle">
											<input type="checkbox" checked={winningTeam === team} onchange={(event) => setTeamWinner(team, (event.target as HTMLInputElement).checked)} />
											Winner
										</label>
									{/if}
								</div>
								{#if playersForTeam(team).length === 0}
									<p class="team-empty">Drop participants here</p>
								{:else}
									<div class="team-list">
										{#each playersForTeam(team) as id}
											<div class="team-player" draggable="true" ondragstart={() => (draggedId = id)} ondragend={() => (draggedId = null)}>
												<span class="avatar">{playerAvatar(id)}</span>
												<span>{playerName(id)}</span>
											</div>
										{/each}
									</div>
								{/if}
							</section>
						{/each}
					</div>
				{/if}
			</div>

			{#if error}<p class="alert-err">{error}</p>{/if}
			{#if review}
				<div class="review" aria-live="polite">
					<div class="review-head">
						<div>
							<p class="section-title">Review before you publish</p>
							<p class="hint">Check the result carefully &mdash; {editing ? 'saving recalculates every rating in this community.' : "reported matches update everyone's rating."}</p>
						</div>
						<button class="btn btn-secondary" onclick={() => (review = false)}>Edit</button>
					</div>

					<div class="review-chips">
						<span class="chip chip-mode">{friendlyMode()}</span>
						<span class="chip">{community?.name ?? communityId}</span>
						<span class="chip">{selected.length} players</span>
						<span class="chip" class:chip-draw={winnerSides().size === 0}>{matchResultLabel()}</span>
					</div>

					<p class="review-subtitle">Finishing order</p>
					<ol class="standings">
						{#each standings() as standing}
							<li class="standing" class:standing-won={standing.won}>
								<span class="place">{standing.place}</span>
								<div class="standing-body">
									<strong>{standing.side}</strong>
									{#if teamNames.length}
										<div class="standing-players">{#each standing.players as id}<span class="standing-player"><span class="avatar">{playerAvatar(id)}</span>{playerName(id)} <span class="muted">{playerRating(id)}</span></span>{/each}</div>
									{:else}
										<span class="muted">rating {playerRating(standing.players[0])}</span>
									{/if}
								</div>
							</li>
						{/each}
					</ol>
				</div>
			{/if}

			<div style="display:flex; gap:12px; flex-wrap:wrap; justify-content:flex-end;">
				<a class="btn btn-secondary" href={editing ? `/history?communityId=${encodeURIComponent(communityId)}` : leaderboardHref}>Cancel</a>
				<button class="btn btn-primary" disabled={submitting} onclick={() => (review ? submit() : (review = true))}>
					{submitting ? 'Saving…' : review ? (editing ? 'Save changes' : 'Confirm report') : (editing ? 'Review changes' : 'Review report')}
				</button>
			</div>
		</section>
	</div>
</div>

<style>
	.community-banner { display:flex; align-items:center; gap:14px; padding:16px 20px; margin-bottom:18px; border:1px solid var(--theme-border); border-radius:14px; background:rgba(255,255,255,.03); }
	.community-banner div { flex:1; }
	.community-banner strong { display:block; font-size:18px; }
	.community-avatar { width:44px; height:44px; border-radius:12px; display:grid; place-items:center; background:linear-gradient(135deg,#b83c08,var(--brand-primary)); color:#fff; font-weight:800; }
	.page-inner { display: flex; flex-direction: column; gap: 18px; max-width: 980px; }
	.nav-row { display: flex; }
	.hero { padding: 26px; }
	.hero h1 { font-size: clamp(28px, 3.8vw, 38px); line-height: 1.05; margin-bottom: 10px; }
	.hero p:last-child { color: var(--theme-text-muted); font-size: 15px; line-height: 1.5; max-width: 620px; }
	.form-shell { padding: 24px; }
	.section-title { font-size: 13px; text-transform: uppercase; letter-spacing: 0.1em; color: var(--brand-muted); font-weight: 800; margin-bottom: 12px; }
	.hint { color: var(--theme-text-muted); font-size: 13px; margin-bottom: 14px; }
	.team-actions { display:flex; align-items:center; gap:12px; flex-wrap:wrap; margin-bottom:14px; }
	.player-picker { display: grid; gap: 8px; margin-bottom: 20px; }
	.player-option { display: flex; align-items: center; gap: 12px; padding: 10px 14px; border-radius: 12px; border: 1px solid var(--theme-border); background: rgba(255, 255, 255, 0.02); cursor: pointer; transition: background 0.15s ease, border-color 0.15s ease; }
	.player-option:hover { background: rgba(255, 255, 255, 0.05); }
	.player-option.sel { border-color: rgba(240, 90, 26, 0.5); background: rgba(240, 90, 26, 0.06); }
	.player-option input[type='checkbox'] { accent-color: var(--theme-accent); width: 18px; height: 18px; flex-shrink: 0; }
	.avatar { width: 32px; height: 32px; border-radius: 50%; background: linear-gradient(135deg, #b83c08, var(--brand-primary)); display: inline-flex; align-items: center; justify-content: center; color: white; font-size: 11px; font-weight: 800; flex-shrink: 0; }
	.opt-name { flex: 1; font-size: 14px; font-weight: 700; }
	.opt-rating { font-size: 12px; }
	.winner-toggle { display:inline-flex; align-items:center; gap:6px; padding:6px 8px; border-radius:999px; background:rgba(255,255,255,.06); color:var(--theme-text); font-size:12px; font-weight:800; }
	.winner-toggle input { accent-color: var(--theme-accent); }
	.place-input { display:inline-flex; align-items:center; gap:6px; padding:4px 8px; border-radius:999px; background:rgba(255,255,255,.06); font-size:12px; font-weight:800; }
	.place-input input { width:44px; padding:4px 6px; border:1px solid var(--theme-border); border-radius:8px; background:rgba(0,0,0,.3); color:var(--theme-text); font-size:12px; text-align:center; }
	.team-grid { display:grid; grid-template-columns:repeat(auto-fit,minmax(220px,1fr)); gap:12px; margin:2px 0 20px; }
	.team-card { min-height:150px; padding:14px; border:1px solid var(--theme-border); border-radius:16px; background:rgba(255,255,255,.03); }
	.team-head { display:flex; align-items:center; justify-content:space-between; gap:10px; margin-bottom:12px; }
	.team-head strong { font-size:14px; }
	.team-empty { min-height:76px; display:grid; place-items:center; border:1px dashed var(--theme-border); border-radius:12px; color:var(--theme-text-muted); font-size:13px; }
	.team-list { display:grid; gap:8px; }
	.team-player { display:flex; align-items:center; gap:10px; padding:8px; border-radius:12px; background:rgba(0,0,0,.22); border:1px solid rgba(255,255,255,.08); cursor:grab; font-size:13px; font-weight:700; }
	.team-player:active { cursor:grabbing; }
	.review { padding:20px; margin-bottom:18px; border:1px solid rgba(240,90,26,.35); border-radius:14px; background:rgba(240,90,26,.08); }
	.review-head { display:flex; align-items:flex-start; justify-content:space-between; gap:16px; flex-wrap:wrap; margin-bottom:14px; }
	.review-chips { display:flex; gap:8px; flex-wrap:wrap; margin-bottom:18px; }
	.chip { padding:6px 12px; border-radius:999px; border:1px solid var(--theme-border); background:rgba(255,255,255,.06); font-size:12px; font-weight:700; }
	.chip-mode { border-color:rgba(240,90,26,.5); background:rgba(240,90,26,.18); color:#ffd9c9; }
	.chip-draw { border-color:rgba(255,209,102,.5); color:#ffd166; }
	.review-subtitle { font-size:11px; font-weight:800; letter-spacing:.1em; text-transform:uppercase; color:var(--brand-muted); margin-bottom:10px; }
	.standings { list-style:none; display:flex; flex-direction:column; gap:8px; }
	.standing { display:flex; align-items:center; gap:14px; padding:12px 14px; border:1px solid var(--theme-border); border-radius:12px; background:rgba(0,0,0,.2); }
	.standing-won { border-color:rgba(110,231,183,.45); background:rgba(6,118,71,.16); }
	.place { min-width:44px; font-size:13px; font-weight:800; color:var(--brand-muted); }
	.standing-won .place { color:var(--ok); }
	.standing-body { display:flex; flex-direction:column; gap:6px; }
	.standing-players { display:flex; gap:10px; flex-wrap:wrap; }
	.standing-player { display:flex; align-items:center; gap:6px; font-size:13px; }
	.standing-player .avatar { width:22px; height:22px; font-size:10px; }
	.alert-err { color: var(--danger); background: var(--danger-bg); border: 1px solid var(--danger-border); border-radius: 14px; padding: 12px 14px; font-size: 13px; font-weight: 600; margin-bottom: 16px; }
</style>
