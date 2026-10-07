import type { PageLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { base } from '$app/paths';

// A community has no page of its own; the leaderboard is its home.
export const load: PageLoad = ({ params }) => {
	redirect(307, `${base}/leaderboard?communityId=${encodeURIComponent(params.id)}`);
};
