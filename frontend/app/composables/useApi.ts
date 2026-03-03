const TOKEN_KEY = 'simplix_token'

export function useApi() {
  const config = useRuntimeConfig()
  const base = config.public.apiBase as string

  const token = useCookie<string | null>(TOKEN_KEY, { default: () => null })

  function headers(): Record<string, string> {
    return token.value ? { Authorization: `Bearer ${token.value}` } : {}
  }

  async function get<T>(path: string, params?: Record<string, string | string[]>): Promise<T> {
    const url = new URL(base + path)
    if (params) {
      Object.entries(params).forEach(([k, v]) => {
        if (Array.isArray(v)) {
          v.forEach(val => val && url.searchParams.append(k, val))
        } else if (v) {
          url.searchParams.set(k, v)
        }
      })
    }
    return $fetch(url.toString(), { headers: headers() }) as Promise<T>
  }

  async function post<T>(path: string, body?: unknown): Promise<T> {
    return $fetch(base + path, { method: 'POST', body: body as Record<string, unknown>, headers: headers() }) as Promise<T>
  }

  async function patch<T>(path: string, body?: unknown): Promise<T> {
    return $fetch(base + path, { method: 'PATCH', body: body as Record<string, unknown>, headers: headers() }) as Promise<T>
  }

  async function del(path: string): Promise<void> {
    await $fetch(base + path, { method: 'DELETE', headers: headers() })
  }

  return { get, post, patch, del, token }
}

export function useAuth() {
  const api = useApi()
  const user = useState<Record<string, unknown> | null>('auth_user', () => null)

  async function signIn(email: string, password: string) {
    const res = await api.post<{ token: string; user: Record<string, unknown> }>('/api/auth/sign_in', { email, password })
    api.token.value = res.token
    user.value = res.user
    return res
  }

  async function signUp(name: string, email: string, password: string) {
    const res = await api.post<{ token: string; user: Record<string, unknown> }>('/api/auth/sign_up', { name, email, password })
    api.token.value = res.token
    user.value = res.user
    return res
  }

  async function signOut() {
    await api.del('/api/auth/sign_out').catch(() => {})
    api.token.value = null
    user.value = null
    navigateTo('/login')
  }

  async function fetchMe() {
    if (!api.token.value) return
    try {
      user.value = await api.get('/api/auth/me')
    }
    catch {
      api.token.value = null
    }
  }

  const isAuthenticated = computed(() => !!api.token.value)

  return { user, signIn, signUp, signOut, fetchMe, isAuthenticated }
}
