import { verifyAccount } from './service/account'
import { clearSession, getSession, parseClaims, saveSession } from './session'
import type { Session } from './session'

export type { Session } from './session'
export { getSession, updateToken } from './session'

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

/** 登录，成功后把令牌和用户信息写入本地 */
export async function signIn(login: string, password: string): Promise<Session> {
  if (!login.trim() || !password) {
    throw new Error('登录名和密码不能为空')
  }

  const token = await verifyAccount(login.trim(), password)
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
  saveSession(session)
  return session
}

/** 退出登录，清除本地登录态 */
export function signOut(): void {
  clearSession()
}
