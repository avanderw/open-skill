import type { PageLoad } from './$types';
import { createApi, type LeaderboardRow, type Community, type Mode } from '$lib/api';

export const load: PageLoad = async ({ fetch, url }) => {
	const api = createApi(fetch);
	const communityId = url.searchParams.get('communityId');
	const communities: Community[] = await api.communities().catch(() => []);
	const availableModes = communityId ? await api.leaderboardModes(communityId) : [];
	const requestedMode = url.searchParams.get('mode');
	const mode: Mode = requestedMode === '1v1' || requestedMode === 'team' || requestedMode === 'ffa'
		? requestedMode : availableModes.includes('team') ? 'team' : availableModes[0] ?? '1v1';
	const players: LeaderboardRow[] = communityId ? await api.leaderboard(mode, { communityId, full: '1' }) : [];
	return { players, communityId, communities, mode };
};