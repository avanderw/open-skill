import type { PageLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { base } from '$app/paths';
import { createApi, type Community } from '$lib/api';

export const load: PageLoad = async ({ fetch, url }) => {
	const api = createApi(fetch);
	const communityId = url.searchParams.get('communityId');
	if (!communityId) {
		// A match always belongs to a community; make the caller pick one first.
		const mine: Community[] = await api.communities().catch(() => []);
		if (mine.length === 1) redirect(307, `${base}/report?communityId=${encodeURIComponent(mine[0].id)}`);
		redirect(307, `${base}/communities`);
	}
	const matchId = url.searchParams.get('matchId');
	// These reads are independent; do not make roster loading wait for community metadata.
	const [community, rows, match] = await Promise.all([
		api.findCommunity(communityId).catch(() => null),
		api.leaderboard('1v1', { communityId, full: '1' }),
		matchId ? api.match(matchId).catch(() => null) : null
	]);
	return { players: rows, communityId, community, match };
};