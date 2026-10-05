import { authHeaders } from '$lib/auth';

// API client — talks to the Go service over /api (proxied in dev).
// Types mirror the service's JSON contracts.
export type Mode = '1v1' | 'team' | 'ffa';

export interface LeaderboardRow {
	rank: number;
	playerId: string;
	handle: string;
	avatar: string;
	rating: number;
	games: number;
	wins: number;
	losses: number;
	isYou: boolean;
	trend: number;
	role?: 'owner' | 'admin' | 'member';
}

export type MembershipStatus = 'owner' | 'member' | 'pending' | 'non_member';

export interface Community {
	id: string;
	name: string;
	mode: 'open' | 'approval';
	draws: boolean;
	code: string;
	ownerId: string;
	members: number;
	yourGames?: number;
	createdAt: string;
	membership?: MembershipStatus;
}

export interface CommunityJoinRequest {
	communityId: string;
	userId: string;
	status: string;
	createdAt: string;
}

export interface MemberMutationResult {
	ok: boolean;
	communityId?: string;
	userId?: string;
	status?: MembershipStatus;
	role?: 'admin' | 'member';
}

export interface Player {
	id: string;
	userId: string;
	handle: string;
	avatar: string;
	createdAt: string;
}

export interface Rating {
	communityId: string;
	communityName?: string;
	playerId: string;
	mode: Mode;
	rating: number;
	uncertainty: number;
	games: number;
	wins: number;
	losses: number;
	draws: number;
	lastPlayed: string;
	rank?: number;
}

export interface Profile {
	player: Player;
	ratings: Rating[];
}

export interface MatchDelta {
	playerId: string;
	mode: Mode;
	delta: number;
	newRating: number;
}

export interface MatchOutcome {
	matchId: string;
	mode: Mode;
	deltas: MatchDelta[];
}

export interface ReportMatchRequest {
	communityId: string;
	participants: { playerId: string; team: string; score: number; finish: number }[];
	notes?: string;
}

export interface DemoUser { user: string; handle: string; playerId: string }

export interface MatchResult {
	matchId: string;
	communityId: string;
	mode: Mode;
	playedAt: string;
	participants: { playerId: string; handle?: string; team: string; score: number; finish: number; ratingDelta: number }[];
}

// Empty base keeps requests same-origin (Vite proxy in dev, nginx in Docker).
const API_BASE_URL: string = import.meta.env.VITE_API_URL ?? '';

function createRequest(fetcher: typeof globalThis.fetch, baseUrl: string) {
	const base = baseUrl.replace(/\/+$/, '');
	return async function req<T>(path: string, init?: RequestInit): Promise<T> {
		const headers: Record<string, string> = await authHeaders();
		if (init?.headers instanceof Headers) init.headers.forEach((value, key) => { headers[key] = value; });
		else if (Array.isArray(init?.headers)) init.headers.forEach(([key, value]) => { headers[key] = value; });
		else Object.assign(headers, init?.headers ?? {});
		if (init?.body !== undefined && !('Content-Type' in headers)) headers['Content-Type'] = 'application/json';
		const res = await fetcher(base + path, {
			...init,
			headers
		});
		if (!res.ok) {
			const body = await res.text();
			let message = body.trim();
			try {
				message = (JSON.parse(body) as { error?: string }).error || message;
			} catch {
				// Preserve useful text responses from proxies and development servers.
			}
			throw new Error(message || `Request failed: ${res.status}`);
		}
		return res.json() as Promise<T>;
	};
}

export function createApi(fetcher: typeof globalThis.fetch = globalThis.fetch, baseUrl: string = API_BASE_URL) {
	const req = createRequest(fetcher, baseUrl);
	return {
		health: () => req<{ ok: boolean }>('/api/health'),
		me: () => req<DemoUser>('/api/me'),
		leaderboard: (mode: Mode, opts?: { q?: string; sort?: 'rank' | 'rating' | 'games'; communityId?: string; full?: string }) => {
			const params = new URLSearchParams({ mode });
			if (opts?.communityId) params.set('communityId', opts.communityId);
			if (opts?.q) params.set('q', opts.q);
			if (opts?.sort) params.set('sort', opts.sort);
			if (opts?.full) params.set('full', opts.full);
			return req<LeaderboardRow[]>(`/api/leaderboard?${params}`);
		},
		leaderboardModes: (communityId?: string) => {
			const params = new URLSearchParams();
			if (communityId) params.set('communityId', communityId);
			const query = params.toString();
			return req<Mode[]>(`/api/leaderboard/modes${query ? `?${query}` : ''}`);
		},
		communities: () => req<Community[]>('/api/communities'),
		createCommunity: (body: { name: string; mode: string; draws: boolean }) =>
			req<Community>('/api/communities', { method: 'POST', body: JSON.stringify(body) }),
		browseCommunities: () => req<Community[]>('/api/communities/browse'),
		membership: (id: string) => req<{ communityId: string; status: MembershipStatus; canManage: boolean }>(`/api/communities/${encodeURIComponent(id)}/membership`),
		approveMember: (communityId: string, userId: string) => req<MemberMutationResult>(`/api/communities/${encodeURIComponent(communityId)}/members/${encodeURIComponent(userId)}/approve`, { method: 'POST' }),
		promoteMember: (communityId: string, userId: string) => req<MemberMutationResult>(`/api/communities/${encodeURIComponent(communityId)}/members/${encodeURIComponent(userId)}/promote`, { method: 'POST' }),
		demoteMember: (communityId: string, userId: string) => req<MemberMutationResult>(`/api/communities/${encodeURIComponent(communityId)}/members/${encodeURIComponent(userId)}/demote`, { method: 'POST' }),
		removeMember: (communityId: string, userId: string) => req<MemberMutationResult>(`/api/communities/${encodeURIComponent(communityId)}/members/${encodeURIComponent(userId)}`, { method: 'DELETE' }),
		findCommunity: (idOrCode: string) => req<Community>(`/api/communities/${encodeURIComponent(idOrCode)}`),
		deleteCommunity: (id: string) => req<{ ok: boolean }>(`/api/communities/${encodeURIComponent(id)}`, { method: 'DELETE' }),
		updateCommunity: (id: string, body: { name: string; mode: 'open' | 'approval'; draws: boolean }) =>
			req<Community>(`/api/communities/${encodeURIComponent(id)}`, { method: 'PUT', body: JSON.stringify(body) }),
		communityMembers: (id: string) => req<CommunityJoinRequest[]>(`/api/communities/${encodeURIComponent(id)}/members`),
		integration: (id: string) => req<{ communityId: string; discord: string; webhook: string }>(`/api/communities/${encodeURIComponent(id)}/integrations`),
		updateIntegration: (id: string, body: { discord: string; webhook: string }) => req<{ communityId: string; discord: string; webhook: string }>(`/api/communities/${encodeURIComponent(id)}/integrations`, { method: 'PUT', body: JSON.stringify(body) }),
		joinCommunity: (code: string) =>
			req<{ ok: boolean; status: MembershipStatus; communityId: string; name: string; mode: string }>(
				`/api/communities/${encodeURIComponent(code)}/join`,
				{ method: 'POST' }
			),
		history: (playerId: string, opts?: { communityId?: string; mode?: Mode; limit?: number; offset?: number }) => {
			const params = new URLSearchParams();
			if (playerId) params.set('playerId', playerId);
			if (opts?.communityId) params.set('communityId', opts.communityId);
			if (opts?.mode) params.set('mode', opts.mode);
			if (opts?.limit !== undefined) params.set('limit', String(opts.limit));
			if (opts?.offset !== undefined) params.set('offset', String(opts.offset));
			return req<MatchResult[]>(`/api/history?${params}`);
		},
		profile: (id: string) => req<Profile>(`/api/profile/${encodeURIComponent(id)}`),
		match: (matchId: string) => req<MatchResult>(`/api/matches/${encodeURIComponent(matchId)}`),
		reportMatch: (body: ReportMatchRequest) => req<MatchOutcome>('/api/matches', { method: 'POST', body: JSON.stringify(body) }),
		updateMatch: (matchId: string, body: { participants: ReportMatchRequest['participants']; notes?: string }) =>
			req<MatchOutcome>(`/api/matches/${encodeURIComponent(matchId)}`, { method: 'PUT', body: JSON.stringify(body) }),
		deleteMatch: (matchId: string) =>
			req<{ ok: boolean; communityId: string }>(`/api/matches/${encodeURIComponent(matchId)}`, { method: 'DELETE' })
	};
}

export const api = createApi();