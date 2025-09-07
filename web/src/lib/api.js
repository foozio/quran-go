export async function api(path, init) {
  const base = import.meta.env.VITE_API_URL || '/api';
  const res = await fetch(base + path, init);
  if (!res.ok) throw new Error(await res.text());
  return await res.json();
}

