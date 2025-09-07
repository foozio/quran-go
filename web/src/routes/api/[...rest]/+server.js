export async function GET({ request, url, params }) {
  const rest = params.rest || '';
  const target = new URL(`http://app:8080/${rest}`);
  target.search = url.search;
  const upstream = await fetch(target, { headers: request.headers });
  const headers = new Headers(upstream.headers);
  return new Response(upstream.body, { status: upstream.status, headers });
}

