import type { PageLoad } from './$types';
import { createApi, type Community, type MatchResult } from '$lib/api';

export const load: PageLoad = async ({ fetch, url }) => {
	const api = createApi(fetch);
	const playerId = url.searchParams.get('playerId') || '';
	const communityId = url.searchParams.get('communityId') || '';
	const history: MatchResult[] = await api.history(playerId, { communityId: communityId || undefined, limit: 50 });
	const community: Community | null = communityId ? await api.findCommunity(communityId).catch(() => null) : null;
	const membership = communityId ? await api.membership(communityId).catch(() => null) : null;
	return { playerId, communityId, history, community, canManage: membership?.canManage ?? false };
};
