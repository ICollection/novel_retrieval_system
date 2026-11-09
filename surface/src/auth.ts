import { request } from './api'

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

/** 与后端 common/enum.go 中的权限常量保持一致。 */
export const Permission = {
  /** 管理系统用户 */
  ManageUser: 'ManageUser',
  /** 管理原始文件 */
  ManageOrigin: 'ManageOrigin',
  /** 管理分析任务 */
  ManageMission: 'ManageMission'
} as const

/** 权限选项，供用户管理页面展示与勾选。 */
export const permissionOptions = [
  { value: Permission.ManageUser, label: '管理用户' },
  { value: Permission.ManageOrigin, label: '管理原始文件' },
  { value: Permission.ManageMission, label: '管理分析任务' }
]

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

/** 是否已登录 */
export function isLoggedIn(): boolean {
  return getSession() !== null
}

/** 判断权限串中是否包含指定权限，匹配规则与后端 Claims.HasPermission 一致。 */
export function hasPermission(permission: string | undefined, target: string): boolean {
  if (!permission) {
    return false
  }
  return permission.split(',').includes(target)
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

  const token = await request<string>('/account/verifyAccount', {
    query: { login: login.trim(), password }
  })
  if (typeof token !== 'string') {
    throw new Error('登录失败')
  }

  const claims = parseClaims(token)
  const session: Session = {
    token,
    name: claims?.name ?? login.trim(),
    login: claims?.login ?? login.trim(),
    permission: claims?.permission
  }
  localStorage.setItem(STORAGE_KEY, JSON.stringify(session))
  return session
}

/** 退出登录，清除本地登录态 */
export function signOut(): void {
  localStorage.removeItem(STORAGE_KEY)
}
