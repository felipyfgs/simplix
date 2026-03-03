export function useSSE() {
  const api = useApi()
  const config = useRuntimeConfig()
  const base = config.public.apiBase as string
  const listeners = new Map<string, ((payload: unknown) => void)[]>()
  let source: EventSource | null = null
  let closed = false

  function on(eventType: string, handler: (payload: unknown) => void) {
    if (!listeners.has(eventType)) listeners.set(eventType, [])
    listeners.get(eventType)!.push(handler)
  }

  function off(eventType: string, handler: (payload: unknown) => void) {
    const arr = listeners.get(eventType)
    if (!arr) return
    const idx = arr.indexOf(handler)
    if (idx >= 0) arr.splice(idx, 1)
  }

  function connect() {
    if (!api.token.value || typeof window === 'undefined') return
    closed = false
    const url = `${base}/api/events?token=${encodeURIComponent(api.token.value)}`
    source = new EventSource(url)

    source.addEventListener('message', (e) => {
      try {
        const evt = JSON.parse(e.data) as { type: string; payload: unknown }
        const handlers = listeners.get(evt.type) ?? []
        handlers.forEach(h => h(evt.payload))
      }
      catch { /* ignore */ }
    })

    // Handle named events
    ;[
      'conversation.created', 'conversation.updated', 'conversation.resolved',
      'message.created', 'job.updated', 'contact.enriched', 'note.created'
    ].forEach(evtType => {
      source!.addEventListener(evtType, (e: MessageEvent) => {
        try {
          const payload = JSON.parse(e.data)
          const handlers = listeners.get(evtType) ?? []
          handlers.forEach(h => h(payload))
        }
        catch { /* ignore */ }
      })
    })

    source.onerror = () => {
      if (closed) return
      source?.close()
      // Reconnect after 5 seconds
      setTimeout(() => { if (!closed) connect() }, 5000)
    }
  }

  function disconnect() {
    closed = true
    source?.close()
    source = null
  }

  return { on, off, connect, disconnect }
}
