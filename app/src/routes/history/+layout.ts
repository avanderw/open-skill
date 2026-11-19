import { redirect } from '@sveltejs/kit';
import { base } from '$app/paths';
import { createApi, type HistoryPlayer } from '$lib/api';
import type { LayoutLoad } from './$types';

// Keep scope metadata separate so changing only the mode reloads just the matches.
export const load: LayoutLoad = async ({ fetch, url }) => {
	const api = createApi(fetch);
	const playerId = url.searchParams.get('playerId') || '';
	const communityId = url.searchParams.get('communityId') || '';
	if (!playerId && !communityId) {
		const me = await api.me();
		const params = new URLSearchParams(url.searchParams);
		params.set('playerId', me.playerId);
		redirect(307, `${base}/history?${params}`);
	}

	const [profile, community, membership, players, communities] = await Promise.all([
		playerId ? api.profile(playerId) : null,
		communityId ? api.findCommunity(communityId) : null,
		communityId ? api.membership(communityId) : null,
		communityId ? api.historyPlayers(communityId) : Promise.resolve<HistoryPlayer[]>([]),
		playerId ? [] : api.communities()
	]);
	const communityNames = new Map(communities.map((entry) => [entry.id, entry.name]));
	for (const rating of profile?.ratings ?? []) {
		communityNames.set(rating.communityId, rating.communityName || rating.communityId);
	}
	if (community) communityNames.set(community.id, community.name);
	if (profile?.player.communityId && !communityNames.has(profile.player.communityId)) {
		const home = await api.findCommunity(profile.player.communityId);
		communityNames.set(home.id, home.name);
	}
	if (profile && !players.some((player) => player.playerId === profile.player.id)) {
		players.push({
			playerId: profile.player.id, handle: profile.player.handle, unclaimed: profile.player.unclaimed
		});
	}
	return {
		playerId: profile?.player.id || playerId,
		communityId,
		player: profile?.player ?? null,
		community,
		canManage: membership?.canManage ?? false,
		playerOptions: players.map((player) => ({
			value: player.playerId, label: player.handle, detail: player.unclaimed ? 'Unclaimed' : ''
		})).sort((a, b) => a.label.localeCompare(b.label)),
		communityOptions: [...communityNames].map(([value, label]) => ({ value, label }))
			.sort((a, b) => a.label.localeCompare(b.label))
	};
};
