<script lang="ts">
	import { goto } from '$app/navigation';
	import { base } from '$app/paths';
</script>

<svelte:head><title>Start a community — Open-Skill</title></svelte:head>
<div class="page-wrap"><div class="page-inner">
	<a class="btn btn-ghost" href={`${base}/communities`}>← My communities</a>
	<section class="card hero"><p class="label-caps">Community journey</p><h1>Bring your players together.</h1><p>Start a new ranking space or join an existing community with a shared invite code.</p></section>
	<section class="choice-grid">
		<button class="card choice" onclick={() => goto(`${base}/communities/create`)}><span class="choice-icon">＋</span><span class="label-caps">Create</span><h2>Start a community</h2><p>Choose the name, draw rules, and how players join.</p><span class="choice-link">Create community →</span></button>
		<div class="card choice"><span class="choice-icon">⌁</span><span class="label-caps">Join</span><h2>Enter with a code</h2><p>Use an invite code from a community owner to join directly.</p><a class="choice-link" href={`${base}/communities#join`}>Join with a code →</a></div>
	</section>
</div></div>
<style>
	.page-inner{display:flex;flex-direction:column;gap:18px;max-width:980px}.hero,.choice{padding:28px}.hero h1{font-size:clamp(32px,5vw,48px);line-height:1.05;margin-bottom:10px}.hero p:last-child,.choice p{color:var(--theme-text-muted);line-height:1.5}.choice-grid{display:grid;grid-template-columns:1fr 1fr;gap:18px}.choice{border:1px solid var(--theme-border);color:var(--theme-text);text-align:left;cursor:pointer;transition:transform .15s,border-color .15s}.choice:hover{transform:translateY(-3px);border-color:var(--theme-accent)}.choice-icon{display:grid;place-items:center;width:56px;height:56px;border-radius:16px;background:rgba(240,90,26,.12);color:var(--theme-accent);font-size:28px;margin-bottom:20px}.choice h2{font-size:24px;margin:4px 0 8px}.choice-link{display:inline-block;margin-top:24px;color:var(--theme-accent);font-weight:800;font-size:13px}.label-caps{font-size:12px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:var(--theme-accent)}@media(max-width:680px){.choice-grid{grid-template-columns:1fr}}
</style>