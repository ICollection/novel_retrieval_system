/** 词元消费，字段与后端 consume 表一致，用户统计与会话统计共用。 */
export interface Consume {
  llm_prompt?: number
  llm_answer?: number
  slm_prompt?: number
  slm_answer?: number
  vem_prompt?: number
}
