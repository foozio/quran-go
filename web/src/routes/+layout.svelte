<script>
  import '../app.css';
  import { onMount, onDestroy } from 'svelte';
  let status = 'checking'; // 'ok' | 'down' | 'checking'
  let latency = null; // ms
  const apiBase = import.meta.env.VITE_API_URL || '/api';

  function timeoutFetch(url, ms) {
    const c = new AbortController();
    const id = setTimeout(() => c.abort('timeout'), ms);
    return fetch(url, { signal: c.signal }).finally(() => clearTimeout(id));
  }

  async function ping() {
    const t0 = Date.now();
    try {
      const res = await timeoutFetch(`${apiBase}/healthz`, 2000);
      if (!res.ok) throw new Error('bad status');
      latency = Date.now() - t0;
      status = 'ok';
    } catch (_) {
      latency = null;
      status = 'down';
    }
  }

  let t;
  onMount(() => {
    ping();
    t = setInterval(ping, 10000);
  });
  onDestroy(() => { if (t) clearInterval(t); });
</script>

<svelte:head>
  <title>Quran Learn</title>
  <meta name="viewport" content="width=device-width,initial-scale=1" />
</svelte:head>

<div class="min-h-screen">
  <header class="max-w-6xl mx-auto px-4 py-6">
    <div class="flex items-center justify-between gap-3">
      <div class="flex items-center gap-3">
        <div class="w-8 h-8 rounded-lg bg-[#7aa2f7]"></div>
        <h1 class="text-xl font-semibold">Quran Learn</h1>
        <span class="text-sm text-[#a8b3cf]">Search • Read • Listen • Review</span>
      </div>
      <div class="flex items-center gap-2 text-sm">
        <span class="text-[#a8b3cf]">API</span>
        <span class="inline-flex items-center gap-2 px-2 py-1 rounded-full border border-[#2a3248] bg-[#151821]">
          <span class="w-2.5 h-2.5 rounded-full"
            class:bg-green-500={status==='ok'}
            class:bg-red-500={status==='down'}
            class:bg-yellow-500={status==='checking'}></span>
          {#if status === 'ok'}
            <span class="text-[#a8b3cf]">up{latency !== null ? ` ${latency}ms` : ''}</span>
          {:else if status === 'down'}
            <span class="text-[#a8b3cf]">down</span>
          {:else}
            <span class="text-[#a8b3cf]">checking…</span>
          {/if}
        </span>
      </div>
    </div>
  </header>
  <main class="max-w-6xl mx-auto px-4 pb-20">
    <slot />
  </main>
</div>
