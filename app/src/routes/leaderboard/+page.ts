import type { PageLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import type { Mode, LeaderboardRow, Community } from '$lib/api';
import { createApi } from '$lib/api';

export const load: PageLoad = async ({ url, fetch }) => {
	const api = createApi(fetch);
	const mode = (url.searchParams.get('mode') as Mode) || '1v1';
	const q = url.searchParams.get('q');
	const sort = url.searchParams.get('sort') as 'rank' | 'rating' | 'games' | null;
	const full = url.searchParams.get('full');
	const communityId = url.searchParams.get('communityId');
	if (!communityId) {
		// A leaderboard is only meaningful inside a community; pick the obvious one
		// when the member belongs to exactly one, otherwise let them choose.
		const mine = await api.communities().catch((): Community[] => []);
		if (mine.length === 1) {
			const target = new URL(url);
			target.searchParams.set('communityId', mine[0].id);
			redirect(307, `${target.pathname}${target.search}`);
		}
		redirect(307, '/communities');
	}
	const rows: LeaderboardRow[] = await api.leaderboard(mode, { q: q || undefined, sort: sort || undefined, full: full || undefined, communityId });
	const availableModes: Mode[] = await api.leaderboardModes(communityId).catch(() => []);
	const community: Community | null = await api.findCommunity(communityId).catch(() => null);
	const membership = await api.membership(communityId);
	if (community && membership) community.membership = membership.status;
	return { mode, rows, communityId, community, availableModes, full: full === '1' };
};