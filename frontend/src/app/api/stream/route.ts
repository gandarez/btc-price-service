// app/api/stream/route.ts
import { NextRequest } from 'next/server';

export const runtime = 'nodejs'; // Required to use streaming

export async function GET(req: NextRequest) {
  const { readable, writable } = new TransformStream();

  const query = req.nextUrl.searchParams.toString()
  const url = `http://localhost:17020/v1/price-stream${query ? '?' + query : ''}`

  // Create AbortController to handle cleanup
  const controller = new AbortController();

  const response = await fetch(url, {
    method: 'GET',
    headers: {
      Accept: 'text/event-stream',
    },
    signal: controller.signal, // Pass abort signal to fetch
  });

  if (!response.body) {
    return new Response('No stream', { status: 502 });
  }

  // Handle client disconnect
  req.signal.addEventListener('abort', () => {
    controller.abort(); // Abort the upstream connection
    writable.close(); // Close the writable stream
  });

  // Pipe with error handling
  response.body.pipeTo(writable).catch(() => {
    // Handle any pipe errors silently
    controller.abort();
  });

  return new Response(readable, {
    headers: {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache',
      Connection: 'keep-alive',
    },
  });
}
