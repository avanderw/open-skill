import type { PageLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { base } from '$app/paths';
import { createApi, type LeaderboardRow, type Community } from '$lib/api';

export const load: PageLoad = async ({ fetch, url }) => {
	const api = createApi(fetch);
	const communityId = url.searchParams.get('communityId');
	if (!communityId) {
		// A match always belongs to a community; make the caller pick one first.
		const mine: Community[] = await api.communities().catch(() => []);
		if (mine.length === 1) redirect(307, `${base}/report?communityId=${encodeURIComponent(mine[0].id)}`);
		redirect(307, `${base}/communities`);
	}
	const community = await api.findCommunity(communityId).catch(() => null);
	// Offer the leaderboard players as pickers for a reported match.
	const rows: LeaderboardRow[] = await api.leaderboard('1v1', { communityId, full: '1' });
	const matchId = url.searchParams.get('matchId');
	const match = matchId ? await api.match(matchId).catch(() => null) : null;
	return { players: rows, communityId, community, match };
};