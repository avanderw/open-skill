// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
import type { MatchDelta, MatchOutcome } from '$lib/api';

declare global {
	namespace App {
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		interface PageState {
			reportedMatch?: MatchOutcome & {
				communityId: string;
				playedAt: string;
				standings: { place: string; side: string; won: boolean }[];
				deltas: (MatchDelta & { handle: string; unclaimed: boolean })[];
			};
		}
		// interface Platform {}
	}
}

export {};
