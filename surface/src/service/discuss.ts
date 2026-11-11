import { API_BASE_URL, checkUnauthorized, renewToken, request } from '../api'
import type { Result } from '../api'
import { getToken } from '../session'
import type { Consume } from './type'

/** 会话，字段与后端 service/discuss_service.go 中的 session 结构一致。 */
export interface SessionItem {
  id: string | null
  title: string | null
  summary: string | null
  contain: string | null
  current: number | null
  create_user: string | null
  create_time: string | null
  update_time: string | null
}

/** 会话消息，字段与后端 service/discuss_service.go 中的 message 结构一致。 */
export interface MessageItem {
  id: string | null
  session: string | null
  serial: number | null
  human: boolean | null
  content: string | null
  novels: string | null
  segments: string | null
  complete: boolean | null
  create_time: string | null
}

/** 查询当前用户的会话列表。 */
export async function filterSession(): Promise<SessionItem[] | null> {
  return request<SessionItem[] | null>('/discuss/filterSession', { token: getToken() })
}

/** 修改会话名称。 */
export async function modifySession(id: string, title: string): Promise<void> {
  const form = new FormData()
  form.append('id', id)
  form.append('title', title)
  await request<string>('/discuss/modifySession', { method: 'POST', token: getToken(), body: form })
}

/** 删除会话，会话中的消息一并删除。 */
export async function deleteSession(id: string): Promise<void> {
  await request<string>('/discuss/deleteSession', {
    method: 'DELETE',
    token: getToken(),
    query: { id }
  })
}

/** 查询会话的历史消息，recent 为最近条数。 */
export async function filterMessage(session: string, recent: number): Promise<MessageItem[] | null> {
  return request<MessageItem[] | null>('/discuss/filterMessage', {
    token: getToken(),
    query: { session, recent: String(recent) }
  })
}

/** 查询会话使用的参考小说标识。 */
export async function filterNovels(id: string): Promise<string[] | null> {
  return request<string[] | null>('/discuss/filterNovels', { token: getToken(), query: { id } })
}

/** 统计会话的词元消费。 */
export async function consumeMission(id: string): Promise<Consume | null> {
  return request<Consume | null>('/discuss/consumeMission', { token: getToken(), query: { id } })
}

/** 提问，返回事件流；令牌失效时后端返回普通 JSON，此时抛出错误。 */
export async function appendMessage(form: FormData, signal?: AbortSignal): Promise<ReadableStream<Uint8Array>> {
  const response = await fetch(`${API_BASE_URL}/discuss/appendMessage`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${getToken()}` },
    body: form,
    signal
  })
  // 令牌临近过期时后端会续签，并通过响应头返回新令牌
  renewToken(response)
  if (!response.ok || !response.body) {
    throw new Error('请求失败，请稍后重试。')
  }
  // 令牌失效时后端返回普通 JSON，而不是事件流
  if (response.headers.get('Content-Type')?.includes('application/json')) {
    const result = (await response.json()) as Result<unknown>
    checkUnauthorized(result.failure)
    throw new Error(result.message ?? '请求失败，请稍后重试。')
  }
  return response.body
}
