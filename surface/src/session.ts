/** 本地登录态的存储键。 */
const STORAGE_KEY = 'novel-qa-session'

export interface Session {
  token: string
  name: string
  login: string
  /** 逗号分隔的权限串，与后端令牌中的 permission 一致。 */
  permission?: string
}

/** 令牌中的用户信息。 */
interface Claims {
  name?: string
  login?: string
  permission?: string
  sub?: string
}

/**
 * 解析令牌载荷，仅用于展示用户信息，不做签名校验。
 * 载荷为 Base64URL 编码的 JSON。
 */
export function parseClaims(token: string): Claims | null {
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

/** 读取本地登录态 */
export function getSession(): Session | null {
  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) {
    return null
  }
  try {
    const session = JSON.parse(raw) as Session
    // 缺少令牌或用户信息的旧数据视为未登录
    if (!session || !session.token || !session.name) {
      return null
    }
    // 旧会话没有权限字段时，从令牌中补全
    if (session.permission === undefined) {
      session.permission = parseClaims(session.token)?.permission
    }
    return session
  } catch {
    return null
  }
}

/** 每次读取当前令牌，令牌续签后自动取到新值 */
export function getToken(): string {
  return getSession()?.token ?? ''
}

/** 保存登录态 */
export function saveSession(session: Session): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(session))
}

/** 用续签的令牌替换本地令牌，其余登录信息保持不变 */
export function updateToken(token: string): void {
  const session = getSession()
  if (!session) {
    return
  }
  localStorage.setItem(STORAGE_KEY, JSON.stringify({ ...session, token }))
}

/** 清除本地登录态 */
export function clearSession(): void {
  localStorage.removeItem(STORAGE_KEY)
}
