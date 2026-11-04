const STORAGE_KEY = 'novel-qa-session'

/** 后端接口地址，端口与 service/internal/config/config.yaml 中的 web.port 保持一致。 */
const API_BASE_URL = `${window.location.protocol}//${window.location.hostname}:8048/service`

export interface Session {
  token: string
  name: string
  login: string
}

/** 后端统一返回结构。 */
interface Result {
  success: boolean
  failure: number
  message?: string
  content?: unknown
}

/** 令牌中的用户信息。 */
interface Claims {
  name?: string
  login?: string
  sub?: string
}

/** 读取本地登录态 */
export function getSession(): Session | null {
  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) {
    return null
  }
  try {
    const session = JSON.parse(raw) as Session
    // 缺少令牌或用户信息的旧数据视为未登录
    return session && session.token && session.name ? session : null
  } catch {
    return null
  }
}

/** 是否已登录 */
export function isLoggedIn(): boolean {
  return getSession() !== null
}

/**
 * 解析令牌载荷，仅用于展示用户信息，不做签名校验。
 * 载荷为 Base64URL 编码的 JSON。
 */
function parseClaims(token: string): Claims | null {
  const part = token.split('.')[1]
  if (!part) {
    return null
  }
  try {
    const base64 = part.replace(/-/g, '+').replace(/_/g, '/')
    const binary = atob(base64.padEnd(Math.ceil(base64.length / 4) * 4, '='))
    const bytes = Uint8Array.from(binary, (char) => char.charCodeAt(0))
    return JSON.parse(new TextDecoder().decode(bytes)) as Claims
  } catch {
    return null
  }
}

/** 登录，成功后把令牌和用户信息写入本地 */
export async function signIn(login: string, password: string): Promise<Session> {
  if (!login.trim() || !password) {
    throw new Error('登录名和密码不能为空')
  }

  const query = new URLSearchParams({ login: login.trim(), password })
  const response = await fetch(`${API_BASE_URL}/account/verify?${query.toString()}`)
  if (!response.ok) {
    throw new Error(`登录请求失败（HTTP ${response.status}）`)
  }
  const result = (await response.json()) as Result
  if (!result.success || typeof result.content !== 'string') {
    throw new Error(result.message ?? '登录失败')
  }

  const token = result.content
  const claims = parseClaims(token)
  const session: Session = {
    token,
    name: claims?.name ?? login.trim(),
    login: claims?.login ?? login.trim()
  }
  localStorage.setItem(STORAGE_KEY, JSON.stringify(session))
  return session
}

/** 退出登录，清除本地登录态 */
export function signOut(): void {
  localStorage.removeItem(STORAGE_KEY)
}
