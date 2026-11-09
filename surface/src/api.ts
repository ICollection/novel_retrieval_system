/** 后端接口地址，取自 .env.development / .env.production 中的 VITE_API_BASE_URL。 */
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL

/** 对象存储访问地址（含存储桶名称），取自 .env.* 中的 VITE_OSS_BASE_URL。 */
export const OSS_BASE_URL = import.meta.env.VITE_OSS_BASE_URL

/** 后端统一返回结构。 */
export interface Result<T> {
  success: boolean
  failure: number
  message?: string
  content?: T
}

/** 接口调用失败，failure 为后端返回的错误码。 */
export class ApiError extends Error {
  failure: number

  constructor(message: string, failure: number) {
    super(message)
    this.failure = failure
  }
}

export interface RequestOptions {
  method?: 'GET' | 'POST' | 'PATCH' | 'DELETE'
  /** 登录令牌，传入后自动加上 Authorization 请求头。 */
  token?: string
  /** 查询参数，值为 undefined 时忽略。 */
  query?: Record<string, string | undefined>
  /** 请求体，传 FormData 时由浏览器自动生成 multipart 边界。 */
  body?: BodyInit
}

/** 调用后端接口，成功返回 content，失败抛出 ApiError。 */
export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const url = new URL(`${API_BASE_URL}${path}`, window.location.origin)
  Object.entries(options.query ?? {}).forEach(([key, value]) => {
    if (value !== undefined) {
      url.searchParams.set(key, value)
    }
  })

  const headers: Record<string, string> = {}
  if (options.token) {
    headers.Authorization = `Bearer ${options.token}`
  }

  const response = await fetch(url, { method: options.method ?? 'GET', headers, body: options.body })
  if (!response.ok) {
    throw new ApiError('请求失败，请稍后重试。', response.status)
  }
  const result = (await response.json()) as Result<T>
  if (!result.success) {
    throw new ApiError(result.message ?? '请求失败', result.failure)
  }
  return result.content as T
}
