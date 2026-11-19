<script lang="ts">
	import { tick } from 'svelte';

	let {
		label, value, options, allLabel, allowClear = true, disabled = false, hint = '',
		emptyText = 'No options found.', onchange
	}: {
		label: string;
		value: string;
		options: { value: string; label: string; detail?: string }[];
		allLabel: string;
		allowClear?: boolean;
		disabled?: boolean;
		hint?: string;
		emptyText?: string;
		onchange: (value: string) => void | Promise<void>;
	} = $props();
	const id = $props.id();
	let trigger: HTMLButtonElement;
	let dialog: HTMLDialogElement;
	let searchInput: HTMLInputElement;
	let search = $state('');
	const selected = $derived(options.find((option) => option.value === value));
	const filtered = $derived(options.filter((option) =>
		`${option.label} ${option.detail ?? ''}`.toLocaleLowerCase().includes(search.trim().toLocaleLowerCase())
	));

	function open() {
		search = '';
		trigger.focus();
		dialog.showModal();
		searchInput.focus();
	}

	async function select(next: string) {
		dialog.close();
		if (next !== value) await onchange(next);
		await tick();
		trigger.focus();
	}
</script>

<button bind:this={trigger} class="picker" {disabled} onclick={open} aria-haspopup="dialog" aria-controls={id}>
	<span>{label}</span>
	<strong>{selected?.label || allLabel}</strong>
	<span class="change">Change</span>
</button>

<dialog bind:this={dialog} {id} aria-labelledby={`${id}-title`}>
	<div class="heading">
		<h2 id={`${id}-title`}>{label}</h2>
		<button class="btn btn-ghost" onclick={() => dialog.close()} aria-label={`Close ${label.toLowerCase()} picker`}>Close</button>
	</div>
	<label for={`${id}-search`}>Search {label.toLowerCase()}</label>
	<input bind:this={searchInput} id={`${id}-search`} class="text-input" type="search" bind:value={search} autocomplete="off" />
	{#if hint}<p class="hint">{hint}</p>{/if}
	<ul>
		<li><button class:selected={!value} aria-pressed={!value} disabled={disabled || !allowClear} onclick={() => select('')}>
			<span>{allLabel}</span>{#if !allowClear}<small>Keep a player or community selected</small>{/if}
		</button></li>
		{#each filtered as option (option.value)}
			<li><button class:selected={value === option.value} aria-pressed={value === option.value} {disabled} onclick={() => select(option.value)}>
				<span>{option.label}</span>{#if option.detail}<small>{option.detail}</small>{/if}
			</button></li>
		{/each}
		{#if !filtered.length}<li class="empty">{search ? 'No options match your search.' : emptyText}</li>{/if}
	</ul>
</dialog>

<style>
	.picker{display:grid;grid-template-columns:minmax(0,1fr) auto;gap:4px 12px;width:100%;padding:14px 16px;text-align:left;background:rgba(255,255,255,.04);border:1px solid var(--theme-border);border-radius:14px;color:var(--theme-text);cursor:pointer}
	.picker>span:first-child{font-size:12px;color:var(--theme-text-muted)}.picker strong{grid-column:1;overflow-wrap:anywhere}.change{grid-column:2;grid-row:1/3;align-self:center;font-size:12px;color:var(--theme-accent)}
	dialog{box-sizing:border-box;position:fixed;inset:auto 0 0;margin:0;width:100%;max-width:none;max-height:85dvh;padding:20px 20px max(20px,env(safe-area-inset-bottom));border:1px solid var(--theme-border);border-radius:20px 20px 0 0;background:var(--theme-surface);color:var(--theme-text)}
	dialog[open]{display:flex;flex-direction:column;gap:12px}dialog::backdrop{background:rgba(0,0,0,.65)}.heading{display:flex;align-items:center;justify-content:space-between;gap:12px}.heading h2{font-size:22px}
	label,.hint{font-size:13px}.hint,small,.empty{color:var(--theme-text-muted)}input{width:100%;flex-shrink:0}ul{list-style:none;padding:0;overflow-y:auto;min-height:0}li+li{margin-top:6px}
	li button{display:flex;flex-direction:column;gap:4px;width:100%;min-height:48px;padding:12px;text-align:left;overflow-wrap:anywhere;border:1px solid var(--theme-border);border-radius:10px;background:transparent;color:var(--theme-text);cursor:pointer}li button.selected{border-color:var(--theme-accent);background:rgba(240,90,26,.12)}small{font-size:12px}.empty{padding:16px 12px}button:disabled{opacity:.5;cursor:default}
	@media(min-width:640px){dialog{inset:0;margin:auto;width:min(480px,calc(100% - 40px));border-radius:20px;max-height:80dvh}}
</style>
