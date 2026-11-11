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

/** 登录失效时的处理函数，由 main.ts 注册。 */
let unauthorizedHandler: (() => void) | undefined

/** 注册登录失效时的处理函数。 */
export function onUnauthorized(handler: () => void): void {
  unauthorizedHandler = handler
}

/** 后端返回 401 时通知登录已失效。 */
export function checkUnauthorized(failure: number): void {
  if (failure === 401) {
    unauthorizedHandler?.()
  }
}

/** 令牌续签时的处理函数，由 main.ts 注册。 */
let renewedHandler: ((token: string) => void) | undefined

/** 注册令牌续签时的处理函数。 */
export function onTokenRenewed(handler: (token: string) => void): void {
  renewedHandler = handler
}

/** 响应头带有续签令牌时，通知外部替换本地令牌。 */
export function renewToken(response: Response): void {
  const token = response.headers.get('X-Token')
  if (token) {
    renewedHandler?.(token)
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
  // 令牌临近过期时后端会续签，并通过响应头返回新令牌
  renewToken(response)
  if (!response.ok) {
    throw new ApiError('请求失败，请稍后重试。', response.status)
  }
  const result = (await response.json()) as Result<T>
  if (!result.success) {
    // 令牌过期或无效时跳转登录页
    checkUnauthorized(result.failure)
    throw new ApiError(result.message ?? '请求失败', result.failure)
  }
  return result.content as T
}
