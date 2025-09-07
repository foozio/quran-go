<script>
  import { onMount } from 'svelte';
  import { api } from '$lib/api';
  let q = '';
  let list = [];
  let results = [];
  let loading = true;
  let error = '';

  onMount(async () => {
    try {
      list = await api('/surah');
    } catch (e) {
      error = 'Failed to load surah list. Check API connectivity.';
      console.error(e);
    } finally {
      loading = false;
    }
  });

  let timer;
  async function doSearch() {
    if (!q.trim()) { results = []; return; }
    try {
      const res = await fetch(`${import.meta.env.VITE_API_URL || '/api'}/search?q=`+encodeURIComponent(q));
      if (res.ok) {
        const data = await res.json();
        results = data.hits || [];
      } else {
        console.warn('Search failed', await res.text());
      }
    } catch (e) {
      console.error('Search error', e);
    }
  }
  function onType(){ clearTimeout(timer); timer = setTimeout(doSearch, 300); }
</script>

<div class="grid md:grid-cols-[280px_1fr] gap-6">
  <aside class="card p-4">
    <div class="flex items-center justify-between mb-2">
      <h3 class="font-semibold">Surah</h3>
      <span class="text-xs text-[#a8b3cf]">{list.length}</span>
    </div>
    {#if error}
      <div class="mb-2 text-sm bg-red-500/10 border border-red-500/30 text-red-300 rounded px-3 py-2">{error}</div>
    {/if}
    {#if loading}
      <p class="text-sm text-[#a8b3cf]">Loading…</p>
    {:else}
      <div class="space-y-1 max-h-[70vh] overflow-auto pr-1">
      {#each list as s}
        <a class="block px-2 py-1 rounded hover:bg-[#1b2030]" href={'/s/'+s.number}>
          [{s.number}] {s.name_ar}
        </a>
      {/each}
      </div>
    {/if}
  </aside>

  <section class="space-y-4">
    <div class="card p-4">
      <input class="w-full bg-[#0f131c] border border-[#222a3d] rounded-lg px-3 py-2"
             placeholder="Search Arabic or translation…" bind:value={q} on:input={onType} />
    </div>

    <div class="card p-4">
      <h3 class="font-semibold mb-2">Results</h3>
      {#if results.length === 0}
        <p class="text-sm text-[#a8b3cf]">Type to search…</p>
      {:else}
        <div class="space-y-2">
          {#each results as h}
            <div class="font-ar">
              <a class="text-[#a8b3cf]" href={'/s/'+h.surah}>Surah {h.surah}:{h.number}</a> —
              <span>{@html h.snip}</span>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  </section>
</div>
