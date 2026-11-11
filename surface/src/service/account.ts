import { request } from '../api'
import { getToken } from '../session'
import type { Consume } from './type'

/** 系统用户，字段与后端 account 表一致。 */
export interface Account {
  id: string
  name: string
  login: string
  /** 逗号分隔的权限串。 */
  permission?: string
}

/** 校验登录名与密码，返回登录令牌。 */
export async function verifyAccount(login: string, password: string): Promise<string> {
  return request<string>('/account/verifyAccount', { query: { login, password } })
}

/** 查询用户列表。 */
export async function filterAccount(): Promise<Account[] | null> {
  return request<Account[] | null>('/account/filterAccount', { token: getToken() })
}

/** 添加用户，permission 为空表示不授予任何权限。 */
export async function appendAccount(
  name: string,
  login: string,
  password: string,
  permission?: string
): Promise<void> {
  await request<string>('/account/appendAccount', {
    method: 'POST',
    token: getToken(),
    query: { name, login, password, permission: permission || undefined }
  })
}

/** 修改用户名称与权限。 */
export async function modifyAccount(id: string, name: string, permission?: string): Promise<void> {
  await request<string>('/account/modifyAccount', {
    method: 'PATCH',
    token: getToken(),
    query: { id, name, permission: permission || undefined }
  })
}

/** 修改用户密码。 */
export async function changeAccount(id: string, password: string): Promise<void> {
  await request<string>('/account/changeAccount', {
    method: 'PATCH',
    token: getToken(),
    query: { id, password }
  })
}

/** 删除用户。 */
export async function deleteAccount(id: string): Promise<void> {
  await request<string>('/account/deleteAccount', {
    method: 'DELETE',
    token: getToken(),
    query: { id }
  })
}

/** 统计用户的词元消费。 */
export async function consumeAccount(id: string): Promise<Consume | null> {
  return request<Consume | null>('/account/consumeAccount', { token: getToken(), query: { id } })
}
