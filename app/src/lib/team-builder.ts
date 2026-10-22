import type { LeaderboardRow } from '$lib/api';

export function buildTeams(pool: LeaderboardRow[], count: number, capacity: number, rotation = 0) {
	const groups = Array.from({ length: count }, (_, index) => ({ players: [] as LeaderboardRow[], total: 0, index }));
	if (!Number.isInteger(count) || count < 2 || !Number.isInteger(capacity) || capacity < 1 || pool.length !== count * capacity) return [];
	const ordered = [...pool].sort((first, second) => second.rating - first.rating);
	for (const player of ordered) {
		const candidates = groups.filter((group) => group.players.length < capacity);
		candidates.sort((first, second) => first.total - second.total || first.players.length - second.players.length || (first.index + rotation) % count - (second.index + rotation) % count);
		const target = candidates[0];
		target.players.push(player);
		target.total += player.rating;
	}
	return groups;
}