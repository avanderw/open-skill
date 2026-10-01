import type { PageLoad } from './$types';
import { createApi, type LeaderboardRow, type Community } from '$lib/api';

export const load: PageLoad = async ({ fetch, url }) => {
	const api = createApi(fetch);
	const communityId = url.searchParams.get('communityId');
	const communities: Community[] = await api.communities().catch(() => []);
	const players: LeaderboardRow[] = communityId ? await api.leaderboard('team', { communityId }) : [];
	return { players, communityId, communities };
};