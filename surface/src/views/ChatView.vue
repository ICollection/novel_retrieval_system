<script setup lang="ts">
/**
 * 对话问答页面。
 *
 * 会话来自 GET /discuss/filterSession，历史消息来自 GET /discuss/filterMessage，
 * 提问通过 POST /discuss/appendMessage 以流式响应（text/event-stream）返回。
 */
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { API_BASE_URL, request } from '../api'
import { getSession } from '../auth'

/** 会话，字段与后端 service/discuss_service.go 中的 session 结构一致。 */
interface SessionItem {
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
interface MessageItem {
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

/** 小说，字段与后端 service/novel.go 中的 novel 结构一致。 */
interface NovelItem {
  id: string
  title?: string
  status?: string
}

/** 页面消息。 */
interface ChatMessage {
  id: number
  role: 'user' | 'assistant'
  content: string
  thinking: string
  time: string
  pending?: boolean
  error?: string
}

/** 流式事件。 */
interface StreamEvent {
  type?: string
  content?: string
}

const token = getSession()?.token ?? ''

const sessions = ref<SessionItem[]>([])
const currentSession = ref<string | null>(null)
const messages = ref<ChatMessage[]>([])
const novels = ref<NovelItem[]>([])
const selectedNovels = ref<string[]>([])

const input = ref('')
const sending = ref(false)
const loadingSessions = ref(false)
const loadingMessages = ref(false)
const scrollEl = ref<HTMLDivElement>()

let sequence = 0
let controller: AbortController | null = null

const currentTitle = computed(() => {
  const target = sessions.value.find((item) => item.id === currentSession.value)
  return target?.title || '新会话'
})

const showGuide = computed(() => messages.value.length === 0 && !loadingMessages.value)

const suggestions = [
  '主角是在第几章正式出场的？',
  '帮我梳理一下主要人物之间的关系',
  '故事最后的结局是怎样的？',
  '小说里有哪些反复出现的意象？'
]

/** 当前时刻，用于本地消息。 */
function clock(): string {
  return new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

/** 格式化后端时间，按浏览器本地时区显示。 */
function formatTime(value?: string | null): string {
  if (!value) {
    return ''
  }
  let text = value.trim().replace(' ', 'T')
  if (!/(Z|[+-]\d{2}:\d{2})$/.test(text)) {
    text += 'Z'
  }
  const target = new Date(text)
  if (Number.isNaN(target.getTime())) {
    return ''
  }
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${target.getFullYear()}-${pad(target.getMonth() + 1)}-${pad(target.getDate())} ${pad(target.getHours())}:${pad(target.getMinutes())}`
}

/** 按空行拆分段落。 */
function paragraphs(content: string): string[] {
  return content.split(/\n{2,}/).filter((item) => item.trim() !== '')
}

function scrollDown() {
  nextTick(() => {
    const element = scrollEl.value
    if (element) {
      element.scrollTop = element.scrollHeight
    }
  })
}

/** 加载会话列表。 */
async function loadSessions() {
  if (!token) {
    return
  }
  loadingSessions.value = true
  try {
    const content = await request<SessionItem[]>('/discuss/filterSession', {
      token
    })
    sessions.value = Array.isArray(content) ? content : []
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '加载会话失败')
  } finally {
    loadingSessions.value = false
  }
}

/** 加载就绪的小说，作为本轮问答的书单候选。 */
async function loadNovels() {
  if (!token) {
    return
  }
  try {
    const content = await request<NovelItem[]>('/prepare/filterNovel', { token })
    novels.value = (Array.isArray(content) ? content : []).filter((item) => item.status === '06')
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '加载小说失败')
  }
}

/** 进入历史会话。 */
async function openSession(id: string) {
  if (sending.value || currentSession.value === id) {
    return
  }
  currentSession.value = id
  messages.value = []
  loadingMessages.value = true
  try {
    const content = await request<MessageItem[]>('/discuss/filterMessage', {
      token,
      query: { session: id }
    })
    messages.value = (Array.isArray(content) ? content : []).map((item) => ({
      id: ++sequence,
      role: item.human ? 'user' : 'assistant',
      content: item.content ?? '',
      thinking: '',
      time: formatTime(item.create_time) || clock()
    }))
    scrollDown()
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '加载会话消息失败')
  } finally {
    loadingMessages.value = false
  }
}

/** 新建会话。 */
function newSession() {
  if (sending.value) {
    return
  }
  currentSession.value = null
  messages.value = []
}

/** 处理一个流式事件。 */
function applyEvent(raw: string, reply: ChatMessage) {
  const line = raw.split('\n').find((item) => item.startsWith('data:'))
  if (!line) {
    return
  }
  let payload: StreamEvent
  try {
    payload = JSON.parse(line.slice(5).trim()) as StreamEvent
  } catch {
    return
  }
  switch (payload.type) {
    case 'think':
      reply.thinking += payload.content ?? ''
      scrollDown()
      break
    case 'answer':
      reply.content += payload.content ?? ''
      scrollDown()
      break
    case 'finish':
      if (payload.content) {
        currentSession.value = payload.content
      }
      void loadSessions()
      break
    case 'error':
      reply.error = payload.content || '回答失败，请稍后重试。'
      break
  }
}

/** 读取流式响应。 */
async function consume(body: ReadableStream<Uint8Array>, reply: ChatMessage) {
  const reader = body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  while (true) {
    const { done, value } = await reader.read()
    if (done) {
      break
    }
    buffer += decoder.decode(value, { stream: true })
    let index = buffer.indexOf('\n\n')
    while (index >= 0) {
      applyEvent(buffer.slice(0, index), reply)
      buffer = buffer.slice(index + 2)
      index = buffer.indexOf('\n\n')
    }
  }
}

/** 发送提问。 */
async function send() {
  const text = input.value.trim()
  if (!text || sending.value) {
    return
  }
  if (selectedNovels.value.length === 0) {
    ElMessage.warning('请先选择参考小说。')
    return
  }

  messages.value.push({ id: ++sequence, role: 'user', content: text, thinking: '', time: clock() })
  const reply = reactive<ChatMessage>({
    id: ++sequence,
    role: 'assistant',
    content: '',
    thinking: '',
    time: clock(),
    pending: true
  })
  messages.value.push(reply)
  input.value = ''
  sending.value = true
  scrollDown()

  const form = new FormData()
  if (currentSession.value) {
    form.append('session', currentSession.value)
  }
  form.append('content', text)
  form.append('novels', selectedNovels.value.join(','))

  controller = new AbortController()
  try {
    const response = await fetch(`${API_BASE_URL}/discuss/appendMessage`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
      body: form,
      signal: controller.signal
    })
    if (!response.ok || !response.body) {
      throw new Error('请求失败，请稍后重试。')
    }
    await consume(response.body, reply)
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      return
    }
    reply.error = error instanceof Error ? error.message : '请求失败，请稍后重试。'
    ElMessage.error(reply.error)
  } finally {
    reply.pending = false
    sending.value = false
    scrollDown()
  }
}

/** 修改会话名称。 */
async function renameSession(item: SessionItem) {
  if (!item.id || sending.value) {
    return
  }
  let title: string
  try {
    const result = await ElMessageBox.prompt('请输入新的会话名称', '重命名会话', {
      inputValue: item.title ?? '',
      inputValidator: (value) => (value && value.trim() !== '' ? true : '会话名称不能为空'),
      confirmButtonText: '保存',
      cancelButtonText: '取消'
    })
    title = result.value.trim()
  } catch {
    return
  }
  try {
    const form = new FormData()
    form.append('id', item.id)
    form.append('title', title)
    await request<string>('/discuss/modifySession', { method: 'POST', token, body: form })
    const target = sessions.value.find((session) => session.id === item.id)
    if (target) {
      target.title = title
    }
    ElMessage.success('会话名称已修改')
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '修改会话失败')
  }
}

/** 删除会话。 */
async function deleteSession(item: SessionItem) {
  if (!item.id || sending.value) {
    return
  }
  try {
    await ElMessageBox.confirm(
      `确定删除会话「${item.title || '未命名会话'}」吗？会话中的消息会一并删除。`,
      '删除会话',
      {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消'
      }
    )
  } catch {
    return
  }
  try {
    await request<string>('/discuss/deleteSession', {
      method: 'DELETE',
      token,
      query: { id: item.id }
    })
    if (currentSession.value === item.id) {
      currentSession.value = null
      messages.value = []
    }
    sessions.value = sessions.value.filter((session) => session.id !== item.id)
    ElMessage.success('会话已删除')
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '删除会话失败')
  }
}

/** 使用推荐问题。 */
function useSuggestion(text: string) {
  if (sending.value) {
    return
  }
  input.value = text
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter' && !event.shiftKey && !event.isComposing) {
    event.preventDefault()
    void send()
  }
}

onMounted(() => {
  void Promise.all([loadSessions(), loadNovels()])
})

onBeforeUnmount(() => {
  controller?.abort()
})
</script>

<template>
  <div class="chat">
    <main class="chat-main">
      <header class="chat-head">
        <div class="page-head">
          <div>
            <h1>对话问答</h1>
            <p>{{ currentTitle }}</p>
          </div>
          <el-button :disabled="sending" @click="newSession">新建会话</el-button>
        </div>
      </header>

      <div ref="scrollEl" class="stream" v-loading="loadingMessages">
        <div class="stream-inner">
          <div v-if="showGuide" class="guide">
            <p class="guide-title">试试这样问</p>
            <div class="chips">
              <button v-for="item in suggestions" :key="item" class="chip" @click="useSuggestion(item)">
                {{ item }}
              </button>
            </div>
          </div>

          <div v-for="message in messages" :key="message.id" class="row" :class="message.role">
            <div class="avatar">{{ message.role === 'user' ? '我' : 'AI' }}</div>
            <div class="bubble-wrap">
              <div class="meta">
                {{ message.role === 'user' ? '你' : '阅读助手' }} · {{ message.time }}
              </div>
              <details v-if="message.thinking" class="thinking" open>
                <summary>思考过程</summary>
                <p>{{ message.thinking }}</p>
              </details>
              <div class="bubble">
                <span v-if="message.pending && !message.content" class="typing">
                  <i /><i /><i />
                </span>
                <template v-else>
                  <p v-for="(para, index) in paragraphs(message.content)" :key="index">{{ para }}</p>
                  <span v-if="message.pending" class="cursor" />
                </template>
              </div>
              <p v-if="message.error" class="error-line">{{ message.error }}</p>
            </div>
          </div>
        </div>
      </div>

      <footer class="composer">
        <div class="composer-tools">
          <span class="tools-label">小说范围</span>
          <el-select
            v-model="selectedNovels"
            multiple
            size="small"
            placeholder="请选择小说范围"
            :disabled="sending"
            class="novel-select"
          >
            <el-option
              v-for="item in novels"
              :key="item.id"
              :label="item.title || item.id"
              :value="item.id"
            />
          </el-select>
          <span v-if="novels.length === 0" class="tools-empty">暂无就绪小说</span>
        </div>

        <div class="composer-inner">
          <el-input
            v-model="input"
            type="textarea"
            resize="none"
            :autosize="{ minRows: 1, maxRows: 5 }"
            placeholder="输入你的问题，例如：主角的师父是谁？"
            @keydown="onKeydown"
          />
          <el-button
            type="primary"
            class="send-btn"
            :disabled="!input.trim() || sending || selectedNovels.length === 0"
            :loading="sending"
            @click="send"
          >
            发送
          </el-button>
        </div>
        <p class="hint">Enter 发送，Shift + Enter 换行</p>
      </footer>
    </main>

    <aside class="session-panel">
      <p class="session-panel-title">历史会话</p>
      <div class="session-list" v-loading="loadingSessions">
        <div v-for="item in sessions" :key="item.id ?? item.title ?? ''" class="session-item">
          <button
            type="button"
            class="session-chip"
            :class="{ 'is-active': item.id === currentSession }"
            :disabled="sending || loadingMessages"
            @click="item.id && openSession(item.id)"
          >
            {{ item.title || '未命名会话' }}
          </button>
          <span v-if="item.id" class="session-actions">
            <button type="button" class="session-action" :disabled="sending" @click.stop="renameSession(item)">
              重命名
            </button>
            <button type="button" class="session-action" :disabled="sending" @click.stop="deleteSession(item)">
              删除
            </button>
          </span>
        </div>
        <span v-if="!loadingSessions && sessions.length === 0" class="session-empty">
          还没有历史会话
        </span>
      </div>
    </aside>
  </div>
</template>

<style scoped>
.chat {
  display: flex;
  height: 100%;
  min-height: 0;
}

.chat-main {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
}

.chat-head {
  padding: 22px 32px 0;
  background: var(--bg-page);
}

/* 会话列表 */
.session-panel {
  display: flex;
  flex: 0 0 232px;
  flex-direction: column;
  min-height: 0;
  padding: 22px 16px 16px;
  border-left: 1px solid var(--line);
  background: var(--bg-page);
}

.session-panel-title {
  margin: 0 0 12px;
  color: var(--text-main);
  font-size: 13px;
  font-weight: 600;
}

.session-list {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 6px;
  min-height: 0;
  overflow-y: auto;
  scrollbar-width: thin;
}

.session-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.session-actions {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  gap: 2px;
}

.session-action {
  padding: 2px 4px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--text-sub);
  font-family: inherit;
  font-size: 11.5px;
  cursor: pointer;
  transition: all 0.16s ease;
}

.session-action:hover:not(:disabled) {
  background: var(--brand-soft);
  color: var(--brand);
}

.session-action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.session-chip {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid var(--line);
  border-radius: 10px;
  background: #fff;
  color: var(--text-sub);
  font-family: inherit;
  font-size: 12.5px;
  text-align: left;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  cursor: pointer;
  transition: all 0.16s ease;
}

.session-chip:hover:not(:disabled) {
  border-color: var(--el-color-primary-light-5);
  color: var(--brand);
  background: var(--brand-soft);
}

.session-chip.is-active {
  border-color: var(--brand);
  background: var(--brand-soft);
  color: var(--brand);
}

.session-chip:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.session-empty {
  color: var(--text-faint);
  font-size: 12.5px;
}

/* 消息流水 */
.stream {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 8px 32px 24px;
  scrollbar-width: thin;
}

.stream-inner {
  max-width: 820px;
  margin: 0 auto;
}

.guide {
  margin: 18px 0 26px 46px;
}

.guide-title {
  margin: 0 0 10px;
  color: var(--text-faint);
  font-size: 12px;
}

.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.chip {
  padding: 8px 14px;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: #fff;
  color: var(--text-sub);
  font-family: inherit;
  font-size: 12.5px;
  cursor: pointer;
  transition: all 0.16s ease;
}

.chip:hover {
  border-color: var(--el-color-primary-light-5);
  color: var(--brand);
  background: var(--brand-soft);
}

.row {
  display: flex;
  gap: 12px;
  margin-bottom: 22px;
}

.row.user {
  flex-direction: row-reverse;
}

.avatar {
  display: grid;
  place-items: center;
  width: 34px;
  height: 34px;
  flex: 0 0 34px;
  border-radius: 10px;
  background: #eef0f6;
  color: #7a8296;
  font-size: 12px;
  font-weight: 600;
}

.row.user .avatar {
  background: linear-gradient(135deg, #6f6fe8 0%, #4a4ab8 100%);
  color: #fff;
}

.bubble-wrap {
  min-width: 0;
  max-width: 78%;
}

.row.user .bubble-wrap {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

.meta {
  margin-bottom: 6px;
  color: var(--text-faint);
  font-size: 11.5px;
}

.thinking {
  margin-bottom: 8px;
  padding: 9px 12px;
  border: 1px dashed var(--line);
  border-radius: 10px;
  background: #fafafd;
  color: var(--text-sub);
  font-size: 12.5px;
  line-height: 1.7;
}

.thinking summary {
  color: var(--text-faint);
  cursor: pointer;
}

.thinking p {
  margin: 8px 0 0;
  max-height: 240px;
  overflow-y: auto;
  white-space: pre-wrap;
}

.bubble {
  padding: 13px 16px;
  border: 1px solid var(--line);
  border-radius: 4px 14px 14px 14px;
  background: var(--bg-card);
  font-size: 14px;
  line-height: 1.75;
  box-shadow: 0 2px 10px rgba(31, 36, 48, 0.035);
}

.bubble p {
  margin: 0 0 10px;
  white-space: pre-wrap;
}

.bubble p:last-child {
  margin-bottom: 0;
}

.row.user .bubble {
  border-color: transparent;
  border-radius: 14px 4px 14px 14px;
  background: var(--brand);
  color: #fff;
  box-shadow: 0 6px 18px rgba(91, 91, 214, 0.22);
}

.cursor {
  display: inline-block;
  width: 6px;
  height: 14px;
  margin-left: 2px;
  vertical-align: -2px;
  background: var(--brand);
  animation: blink 1s infinite;
}

.error-line {
  margin: 8px 0 0;
  color: var(--el-color-danger);
  font-size: 12px;
}

/* 打字指示 */
.typing {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  height: 22px;
}

.typing i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #c3c8d6;
  animation: blink 1.2s infinite ease-in-out;
}

.typing i:nth-child(2) {
  animation-delay: 0.16s;
}

.typing i:nth-child(3) {
  animation-delay: 0.32s;
}

@keyframes blink {
  0%,
  80%,
  100% {
    opacity: 0.35;
    transform: translateY(0);
  }
  40% {
    opacity: 1;
    transform: translateY(-3px);
  }
}

/* 输入区 */
.composer {
  flex: 0 0 auto;
  padding: 12px 32px 16px;
  border-top: 1px solid var(--line);
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(8px);
}

.composer-tools {
  display: flex;
  align-items: center;
  gap: 8px;
  max-width: 820px;
  margin: 0 auto 8px;
}

.tools-label {
  color: var(--text-faint);
  font-size: 12px;
  white-space: nowrap;
}

.novel-select {
  flex: 1;
  min-width: 0;
}

.novel-select :deep(.el-select__wrapper) {
  height: auto;
  min-height: 24px;
  padding-top: 3px;
  padding-bottom: 3px;
}

.novel-select :deep(.el-select__tags) {
  flex-wrap: wrap;
  row-gap: 4px;
}

.tools-empty {
  color: var(--text-faint);
  font-size: 12px;
}

.composer-inner {
  display: flex;
  align-items: flex-end;
  gap: 10px;
  max-width: 820px;
  margin: 0 auto;
}

.composer-inner :deep(.el-textarea__inner) {
  padding: 11px 14px;
  min-height: 44px !important;
  border-radius: 12px;
  box-shadow: 0 0 0 1px var(--line) inset;
  font-size: 14px;
  line-height: 1.6;
  background: #fbfbfe;
}

.composer-inner :deep(.el-textarea__inner:focus) {
  box-shadow: 0 0 0 1px var(--brand) inset;
  background: #fff;
}

.send-btn {
  height: 44px;
  padding: 0 22px;
  border-radius: 12px;
  font-weight: 500;
}

.hint {
  margin: 8px 0 0;
  text-align: center;
  color: var(--text-faint);
  font-size: 11.5px;
}
</style>
