<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref } from 'vue'

interface Source {
  chapter: string
  title: string
}

interface Message {
  id: number
  role: 'user' | 'assistant'
  content: string
  time: string
  pending?: boolean
  sources?: Source[]
}

interface MockReply {
  match: string[]
  content: string
  sources: Source[]
}

const clock = () =>
  new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })

let seq = 0

const messages = ref<Message[]>([
  {
    id: ++seq,
    role: 'assistant',
    content:
      '你好，我是这部小说的阅读助手。已经入库的内容都可以问：情节走向、人物关系、设定细节，或者某个意象在原文中出现的位置。\n\n回答下方会附上参考章节，方便你回到原文核对。',
    time: clock()
  }
])

const input = ref('')
const sending = ref(false)
const scrollEl = ref<HTMLDivElement>()

const timers: number[] = []
const showGuide = computed(() => messages.value.length === 1)

const suggestions = [
  '主角是在第几章正式出场的？',
  '帮我梳理一下主要人物之间的关系',
  '故事最后的结局是怎样的？',
  '小说里有哪些反复出现的意象？'
]

const mockReplies: MockReply[] = [
  {
    match: ['人物', '关系', '角色', '谁'],
    content:
      '从已入库的章节来看，故事主要围绕三条人物线索展开。\n\n一是主角与师门之间的羁绊，前期以“授业—误解—和解”的节奏推进；二是主角与同辈伙伴的关系，他们在中段的情节转折后形成稳定的行动小组；三是反派阵营内部的裂痕，这条线在后期直接推动了终局冲突。\n\n作者很少直接交代关系，更多信息藏在对话和细节里，建议把第 3、17、42 章的相关段落放在一起读。',
    sources: [
      { chapter: '第 3 章', title: '旧院的灯' },
      { chapter: '第 17 章', title: '同行者' },
      { chapter: '第 42 章', title: '裂缝' }
    ]
  },
  {
    match: ['出场', '第几章', '第一次', '登场'],
    content:
      '主角第一次正式登场是在第 2 章的结尾：雨中有人替他撑伞，但没有给出姓名，只描写了袖口的旧纹样。\n\n到第 4 章，这个纹样在当铺里被再次提起，身份才被间接确认。真正以本名开口说话，要等到第 6 章与师父的对话。\n\n如果只统计“有名字出现”的位置，第一次在第 4 章中段。',
    sources: [
      { chapter: '第 2 章', title: '雨里的伞' },
      { chapter: '第 4 章', title: '当铺' },
      { chapter: '第 6 章', title: '拜师' }
    ]
  },
  {
    match: ['结局', '最后', '结尾', '大结局'],
    content:
      '结局处理得比较克制，没有正面写胜负，而是落在一次告别上。\n\n终章里，主角回到开篇的旧院，把一直随身携带的物件留在窗台上，然后离开。前文埋下的几处伏笔——旧纹样、那盏灯、以及第 42 章提到的裂缝——都在这一章被简短地回应。\n\n整体基调偏向“完成之后回到起点”，开放式但不算留白过多。',
    sources: [
      { chapter: '第 118 章', title: '归院' },
      { chapter: '第 119 章', title: '窗台' }
    ]
  },
  {
    match: ['意象', '象征', '反复', '细节'],
    content:
      '反复出现的意象主要有三个：灯、雨和裂缝。\n\n“灯”出现频次最高，多与人物状态相关，灯亮表示有人等候，灯灭则意味着关系断裂；“雨”几乎总出现在转折之前，起预告作用；“裂缝”出现次数不多，但每次都与阵营分裂对应。\n\n这三组意象在第 3、17、42、118 章形成了比较完整的呼应链条。',
    sources: [
      { chapter: '第 3 章', title: '旧院的灯' },
      { chapter: '第 42 章', title: '裂缝' },
      { chapter: '第 118 章', title: '归院' }
    ]
  }
]

const fallbackReply: MockReply = {
  match: [],
  content:
    '这个问题在已入库的章节里没有完全对应的段落，我按相关情节做了一个梳理。\n\n从上下文推断，作者更倾向于用细节暗示而不是直接说明，因此同一条信息往往分散在几个章节里。建议先确认你关注的是人物动机还是情节事实，我可以再缩小检索范围。\n\n（当前为界面演示，回答由模拟数据生成。）',
  sources: [
    { chapter: '第 27 章', title: '夜谈' },
    { chapter: '第 56 章', title: '回望' }
  ]
}

function pickReply(question: string): MockReply {
  return mockReplies.find((item) => item.match.some((key) => question.includes(key))) ?? fallbackReply
}

function paragraphs(content: string) {
  return content.split('\n\n').filter(Boolean)
}

function scrollDown() {
  nextTick(() => {
    const el = scrollEl.value
    if (el) el.scrollTop = el.scrollHeight
  })
}

function streamReply(reply: MockReply) {
  const message: Message = {
    id: ++seq,
    role: 'assistant',
    content: '',
    time: clock(),
    pending: true
  }
  messages.value.push(message)
  scrollDown()

  const startTimer = window.setTimeout(() => {
    message.pending = false
    let cursor = 0
    const timer = window.setInterval(() => {
      cursor = Math.min(reply.content.length, cursor + 2)
      message.content = reply.content.slice(0, cursor)
      scrollDown()
      if (cursor >= reply.content.length) {
        window.clearInterval(timer)
        message.sources = reply.sources
        sending.value = false
        scrollDown()
      }
    }, 24)
    timers.push(timer)
  }, 620)
  timers.push(startTimer)
}

function send() {
  const text = input.value.trim()
  if (!text || sending.value) return

  messages.value.push({ id: ++seq, role: 'user', content: text, time: clock() })
  input.value = ''
  sending.value = true
  scrollDown()

  streamReply(pickReply(text))
}

function useSuggestion(text: string) {
  if (sending.value) return
  input.value = text
  send()
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter' && !event.shiftKey && !event.isComposing) {
    event.preventDefault()
    send()
  }
}

onBeforeUnmount(() => {
  timers.forEach((timer) => {
    window.clearTimeout(timer)
    window.clearInterval(timer)
  })
})
</script>

<template>
  <div class="chat">
    <header class="chat-head">
      <div class="page-head">
        <div>
          <h1>对话问答</h1>
          <p>当前及历史回答按时间顺序展示在上方</p>
        </div>
        <span class="conn-badge"><i />界面演示</span>
      </div>
    </header>

    <div ref="scrollEl" class="stream">
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
            <div class="bubble">
              <span v-if="message.pending" class="typing">
                <i /><i /><i />
              </span>
              <template v-else>
                <p v-for="(para, index) in paragraphs(message.content)" :key="index">{{ para }}</p>
              </template>
            </div>
            <div v-if="message.sources && message.sources.length" class="sources">
              <span class="sources-label">参考章节</span>
              <el-tag
                v-for="source in message.sources"
                :key="source.chapter"
                size="small"
                effect="plain"
                round
              >
                {{ source.chapter }} · {{ source.title }}
              </el-tag>
            </div>
          </div>
        </div>
      </div>
    </div>

    <footer class="composer">
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
          :disabled="!input.trim() || sending"
          :loading="sending"
          @click="send"
        >
          发送
        </el-button>
      </div>
      <p class="hint">Enter 发送，Shift + Enter 换行 · 回答由模拟数据生成</p>
    </footer>
  </div>
</template>

<style scoped>
.chat {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

.chat-head {
  padding: 22px 32px 16px;
  background: var(--bg-page);
}

.conn-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 11px;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: #fff;
  color: var(--text-sub);
  font-size: 12px;
  white-space: nowrap;
}

.conn-badge i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #f0b429;
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

.sources {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin-top: 9px;
}

.sources-label {
  margin-right: 2px;
  color: var(--text-faint);
  font-size: 11.5px;
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
  padding: 14px 32px 16px;
  border-top: 1px solid var(--line);
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(8px);
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
