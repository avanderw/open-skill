<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { base } from '$app/paths';

	let retrying = $state(false);
	let retryError = $state('');
	const savedReport = $derived($page.state.reportedMatch?.communityId === $page.url.searchParams.get('communityId') ? $page.state.reportedMatch : undefined);

	async function retry() {
		if (retrying) return;
		retrying = true;
		retryError = '';
		try {
			await goto($page.url, { invalidateAll: true, state: $page.state });
		} catch (error) {
			console.error('Unable to retry loading the leaderboard.', error);
			retryError = 'The leaderboard is still unavailable. Check your connection and try again.';
		} finally {
			retrying = false;
		}
	}
</script>

<svelte:head><title>Leaderboard unavailable — Open-Skill</title></svelte:head>

<div class="page-wrap">
	<section class="card error-card" aria-labelledby="error-heading">
		{#if savedReport}
			<p class="saved-status" role="status">Match saved successfully</p>
			<h1 id="error-heading">Your match was saved</h1>
			<p>We could not load the leaderboard. Do not resubmit your match; retrying here only reloads the leaderboard.</p>
		{:else}
			<h1 id="error-heading">Leaderboard unavailable</h1>
		{/if}
		<p class="error" role="alert">{retryError || `Unable to load the leaderboard (${$page.status}): ${$page.error?.message ?? 'Please try again.'}`}</p>
		<div class="actions">
			<button class="btn btn-primary" disabled={retrying} onclick={retry}>{retrying ? 'Loading leaderboard…' : 'Try leaderboard again'}</button>
			{#if savedReport}
				<a class="btn btn-secondary" href={`${base}/report?communityId=${encodeURIComponent(savedReport.communityId)}`}>Log another match</a>
			{/if}
			<a class="btn btn-ghost" href={`${base}/communities`}>My communities</a>
		</div>
	</section>
</div>

<style>
	.error-card { width:100%; max-width:680px; padding:24px; }
	h1 { font-size:28px; margin-bottom:12px; }
	p { line-height:1.6; margin-bottom:16px; }
	.saved-status { color:var(--ok); font-weight:800; }
	.error { color:var(--danger); }
	.actions { display:flex; flex-wrap:wrap; gap:12px; }
</style>
