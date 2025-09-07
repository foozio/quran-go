<script>
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  let n = 1;
  $: n = parseInt($page.params.n || '1', 10);
  let rows = [];
  let loading = true;
  let error = '';
  onMount(async () => {
    loading = true;
    try {
      const res = await fetch(`${import.meta.env.VITE_API_URL || '/api'}/surah/`+n);
      if (!res.ok) throw new Error(await res.text());
      const data = await res.json();
      rows = data.ayah || [];
    } catch (e) {
      error = 'Failed to load surah '+n+'.';
      console.error(e);
    } finally {
      loading = false;
    }
  });
</script>

<a href="/" class="text-sm text-[#a8b3cf]">← Back</a>
<h2 class="text-xl font-semibold mt-2">Surah {n}</h2>
{#if error}
  <div class="mt-2 text-sm bg-red-500/10 border border-red-500/30 text-red-300 rounded px-3 py-2">{error}</div>
{/if}

<div class="mt-4 space-y-3">
  {#if loading}
    <p class="text-sm text-[#a8b3cf]">Loading…</p>
  {:else}
    {#each rows as a}
      <div class="card p-4">
        <div class="text-sm text-[#a8b3cf] mb-1">{n}:{a.ayah}</div>
        <div class="text-right text-2xl leading-[2.2rem] font-ar">{a.arabic}</div>
        {#if a.trans}
          <div class="mt-1">{a.trans}</div>
        {/if}
        {#if a.audio_url}
          <audio class="mt-2 w-full" controls preload="none" src={a.audio_url}></audio>
        {/if}
      </div>
    {/each}
  {/if}
  </div>
