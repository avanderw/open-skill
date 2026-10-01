<script lang="ts">
	import { goto } from '$app/navigation';
</script>

<div class="page-wrap">
	<div class="page-inner home">
		<section class="card hero">
			<p class="label-caps">Competitive ranking</p>
			<h1>Open&#8209;Skill</h1>
			<p>Rank players and teams across any community — sport, esports, or anything competitive.</p>
			<div class="actions">
				<button class="btn btn-primary" onclick={() => goto('/leaderboard')}>View leaderboard</button>
				<button class="btn btn-secondary" onclick={() => goto('/communities')}>My communities</button>
			</div>
		</section>

		<section class="cards">
			<a href="/leaderboard" class="card tile">
				<p class="label-caps">Leaderboard</p>
				<h2>See the top players</h2>
				<p>OpenSkill ratings across 1v1, team, and FFA — updated after every reported match.</p>
			</a>
			<a href="/communities" class="card tile">
				<p class="label-caps">Communities</p>
				<h2>Join by code or create your own</h2>
				<p>Find your group, open or approval-based, and start competing.</p>
			</a>
			<a href="/report" class="card tile">
				<p class="label-caps">Report match</p>
				<h2>Submit a result</h2>
				<p>Pick participants and a winner — ratings update automatically.</p>
			</a>
		</section>
	</div>
</div>

<style>
	.home {
		display: flex;
		flex-direction: column;
		gap: 20px;
	}
	.hero {
		padding: 40px 32px;
	}
	.hero h1 {
		font-size: clamp(32px, 6vw, 52px);
		font-weight: 800;
		line-height: 1.05;
		margin-bottom: 12px;
	}
	.hero p {
		max-width: 560px;
		font-size: 16px;
		color: var(--theme-text-muted);
		line-height: 1.55;
		margin-bottom: 24px;
	}
	.actions {
		display: flex;
		gap: 12px;
		flex-wrap: wrap;
	}
	.cards {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
		gap: 16px;
	}
	.tile {
		display: block;
		padding: 24px;
		transition: transform 0.15s ease, border-color 0.15s ease;
	}
	.tile:hover {
		transform: translateY(-3px);
		border-color: rgba(240, 90, 26, 0.5);
	}
	.tile h2 {
		font-size: 20px;
		font-weight: 700;
		color: var(--theme-text);
		margin-bottom: 8px;
	}
	.tile p:last-child {
		color: var(--theme-text-muted);
		font-size: 14px;
		line-height: 1.5;
	}
</style>