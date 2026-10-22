import { describe, expect, it, vi } from 'vitest';
import { createApi, type LeaderboardRow, type Mode } from '$lib/api';
import { buildTeams } from '$lib/team-builder';
import { load } from '../routes/team/+page';

vi.mock('$lib/api', () => ({ createApi: vi.fn() }));

function players(ratings: number[]): LeaderboardRow[] {
	return ratings.map((rating, index) => ({ rank: index + 1, playerId: `player-${index}`, handle: `Player ${index}`, avatar: '', rating, games: 0, wins: 0, losses: 0, isYou: false, unclaimed: false, trend: 0 }));
}

describe('team builder', () => {
	it('balances four players into two teams of two, including regeneration', () => {
		const pool = players([1800, 1600, 1400, 1200]);
		for (let rotation = 0; rotation < 10; rotation += 1) {
			const teams = buildTeams(pool, 2, 2, rotation);
			expect(teams.map((team) => team.players.length)).toEqual([2, 2]);
			expect(teams.map((team) => team.total)).toEqual([3000, 3000]);
		}
	});

	it('enforces capacity and assigns every player exactly once across supported configurations', () => {
		for (const count of [2, 3, 4]) {
			for (let capacity = 1; capacity <= 10; capacity += 1) {
				const pool = players(Array.from({ length: count * capacity }, (_, index) => index === 0 ? 10000 : index % 3));
				for (let rotation = 0; rotation < 8; rotation += 1) {
					const teams = buildTeams(pool, count, capacity, rotation);
					expect(teams.map((team) => team.players.length)).toEqual(Array(count).fill(capacity));
					expect(teams.flatMap((team) => team.players).map((player) => player.playerId).sort()).toEqual(pool.map((player) => player.playerId).sort());
				}
			}
		}
	});

	it('rejects incomplete rosters and invalid capacities', () => {
		expect(buildTeams(players([100, 200, 300]), 2, 2)).toEqual([]);
		expect(buildTeams(players([100, 200, 300, 400]), 2, 1.5)).toEqual([]);
		expect(buildTeams([], 2, 0)).toEqual([]);
	});
});

describe('team builder ratings', () => {
	it.each([
		{ modes: ['1v1'], query: '', expected: '1v1' },
		{ modes: ['1v1', 'team'], query: '', expected: 'team' },
		{ modes: ['ffa'], query: '', expected: 'ffa' },
		{ modes: [], query: '', expected: '1v1' },
		{ modes: ['team', '1v1'], query: '&mode=1v1', expected: '1v1' },
		{ modes: ['1v1'], query: '&mode=team', expected: 'team' },
		{ modes: ['1v1'], query: '&mode=invalid', expected: '1v1' }
	])('loads the full roster using $expected ratings for $modes $query', async ({ modes, query, expected }) => {
		const pool = players([1800, 1600, 1400, 1200]);
		const leaderboard = vi.fn().mockResolvedValue(pool);
		vi.mocked(createApi).mockReturnValue({
			communities: vi.fn().mockResolvedValue([]),
			leaderboardModes: vi.fn().mockResolvedValue(modes as Mode[]),
			leaderboard
		} as unknown as ReturnType<typeof createApi>);
		const result = await load({ fetch, url: new URL(`http://localhost/team?communityId=community-one${query}`) } as Parameters<typeof load>[0]);
		expect(leaderboard).toHaveBeenCalledWith(expected, { communityId: 'community-one', full: '1' });
		expect(result).toMatchObject({ players: pool, mode: expected });
	});
});