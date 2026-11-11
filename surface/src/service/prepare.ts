import { request } from '../api'
import { getToken } from '../session'

/** 后端返回的小说信息。 */
export interface NovelItem {
  /** 小说标识 */
  id: string
  /** 原始文件标识 */
  origin?: string
  /** 书名 */
  title?: string
  /** 字符数 */
  character_count?: number
  /** 章节数 */
  chapter_count?: number
  /** 任务状态，06 表示已就绪 */
  status?: string
  /** 创建人编码 */
  create_code?: string
  /** 创建人名称 */
  create_name?: string
  /** 创建时间 */
  create_time?: string
  /** 更新时间 */
  update_time?: string
}

/** 后端返回的分析任务信息。 */
export interface MissionItem {
  /** 任务标识 */
  id: string
  /** 所属小说标识 */
  novel_code?: string
  /** 所属小说名称 */
  novel_name?: string
  /** 任务状态，01 待启动、02 已启动、03 解析中、04 切分中、05 向量中、06 已完成、07 已失败、08 已取消 */
  status?: string
  /** 分片总数 */
  total_count?: number
  /** 已向量化数量 */
  embed_count?: number
  /** 启动时间 */
  launch_time?: string
  /** 结束时间 */
  finish_time?: string
  /** 更新时间 */
  update_time?: string
  /** 重试次数 */
  retry_count?: number
  /** 失败原因 */
  message?: string
}

/** 上传原始文件。 */
export async function uploadOrigin(form: FormData): Promise<void> {
  await request<string>('/prepare/uploadOrigin', { method: 'POST', token: getToken(), body: form })
}

/** 查询分析任务列表。 */
export async function filterMission(): Promise<MissionItem[] | null> {
  return request<MissionItem[] | null>('/prepare/filterMission', { token: getToken() })
}

/** 启动任务。 */
export async function launchMission(id: string): Promise<void> {
  await request<unknown>('/prepare/launchMission', {
    method: 'PATCH',
    token: getToken(),
    query: { id }
  })
}

/** 取消任务。 */
export async function cancelMission(id: string): Promise<void> {
  await request<unknown>('/prepare/cancelMission', {
    method: 'PATCH',
    token: getToken(),
    query: { id }
  })
}

/** 查询小说列表。 */
export async function filterNovel(): Promise<NovelItem[] | null> {
  return request<NovelItem[] | null>('/prepare/filterNovel', { token: getToken() })
}

/** 修改小说名称。 */
export async function modifyNovel(id: string, title: string): Promise<void> {
  await request<string>('/prepare/modifyNovel', {
    method: 'PATCH',
    token: getToken(),
    query: { id, title }
  })
}

/** 删除小说，连同任务、章节、分片、向量与原始文件一起清除。 */
export async function deleteNovel(id: string): Promise<void> {
  await request<string>('/prepare/deleteNovel', {
    method: 'DELETE',
    token: getToken(),
    query: { id }
  })
}
