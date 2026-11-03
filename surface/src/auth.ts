const STORAGE_KEY = 'novel-qa-session'

export interface Session {
  token: string
  username: string
}

/** 读取本地登录态 */
export function getSession(): Session | null {
  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) {
    return null
  }
  try {
    const session = JSON.parse(raw) as Session
    return session && session.token ? session : null
  } catch {
    return null
  }
}

/** 是否已登录 */
export function isLoggedIn(): boolean {
  return getSession() !== null
}

/**
 * 登录。当前仅前端演示，未接入后端，
 * 任意非空用户名和密码都视为登录成功。
 */
export function signIn(username: string, password: string): Session {
  if (!username.trim() || !password) {
    throw new Error('用户名和密码不能为空')
  }
  const session: Session = {
    token: `mock-${Date.now()}`,
    username: username.trim()
  }
  localStorage.setItem(STORAGE_KEY, JSON.stringify(session))
  return session
}

/** 退出登录，清除本地登录态 */
export function signOut(): void {
  localStorage.removeItem(STORAGE_KEY)
}
