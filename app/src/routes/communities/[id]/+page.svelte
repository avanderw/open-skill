<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import type { Community } from '$lib/api';
	let { data } = $props();
	const community = data.community as Community;
	onMount(() => goto(`/leaderboard?communityId=${encodeURIComponent(community.id)}`));
</script>
<svelte:head><title>{community.name} leaderboard — Open-Skill</title></svelte:head>
<div class="page-wrap"><div class="page-inner"><section class="card redirect"><p class="label-caps">Community</p><h1>{community.name}</h1><p>Opening this community's leaderboard…</p><a class="btn btn-primary" href={`/leaderboard?communityId=${encodeURIComponent(community.id)}`}>Continue to leaderboard</a></section></div></div>
<style>
.redirect{max-width:680px;margin:60px auto;padding:36px}.redirect h1{font-size:clamp(30px,5vw,48px);line-height:1.05;margin-bottom:10px}.redirect p:not(.label-caps){color:var(--theme-text-muted);margin-bottom:18px}.label-caps{font-size:12px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:var(--theme-accent);margin-bottom:8px}
</style>
