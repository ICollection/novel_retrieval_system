/// <reference types="vite/client" />

/** 环境变量。 */
interface ImportMetaEnv {
  /** 后端接口地址。 */
  readonly VITE_API_BASE_URL: string
  /** 对象存储访问地址（含存储桶名称）。 */
  readonly VITE_OSS_BASE_URL: string
}
