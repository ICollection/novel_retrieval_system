<script setup lang="ts">
/**
 * 文档处理页面。
 *
 * 页面分三块：上传文件、任务列表（控制启动/取消/重试）、小说列表（内容信息）。
 * 小说来自 POST /novel/filter，任务来自 POST /mission/filter，前端按 novel 合并后展示。
 */
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadFile } from 'element-plus'
import { request } from '../api'
import { getSession, hasPermission, Permission } from '../auth'

/** 后端返回的小说信息。 */
interface NovelItem {
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
interface MissionItem {
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

type TagType = 'success' | 'info' | 'warning' | 'danger' | undefined

/** 任务状态展示信息。 */
const statusMeta: Record<string, { label: string; type: TagType }> = {
  '01': { label: '待启动', type: 'info' },
  '02': { label: '已启动', type: 'info' },
  '03': { label: '解析中', type: 'warning' },
  '04': { label: '切分中', type: 'warning' },
  '05': { label: '向量中', type: 'warning' },
  '06': { label: '已完成', type: 'success' },
  '07': { label: '已失败', type: 'danger' },
  '08': { label: '已取消', type: 'info' }
}

/** 正在处理中的状态。 */
const runningStatus = ['02', '03', '04', '05']

/** 本地上传条目，只保留上传中和失败的记录。 */
interface LocalFile {
  key: number
  name: string
  size: number
  status: 'uploading' | 'failed'
  error?: string
}

const files = ref<LocalFile[]>([])
/** 本地文件句柄，用于失败重传（不放入响应式数据，避免代理干扰 FormData） */
const raws = new Map<number, File>()
let seq = 0

const novels = ref<NovelItem[]>([])
const missions = ref<MissionItem[]>([])
const loading = ref(false)
const keyword = ref('')
const tab = ref<'task' | 'novel'>('task')

const session = getSession()
const canUpload = hasPermission(session?.permission, Permission.ManageOrigin)
const canHandle = hasPermission(session?.permission, Permission.ManageMission)
const canModify = hasPermission(session?.permission, Permission.ManageOrigin)

/** 当前筛选后的小说。 */
const filtered = computed(() => {
  const text = keyword.value.trim()
  if (!text) return novels.value
  return novels.value.filter((item) => (item.title ?? '').includes(text))
})

/** 任务视图：显示全部任务。 */
const tasks = computed(() => missions.value)

/** 是否有正在处理的任务，决定是否轮询。 */
const hasRunning = computed(() => missions.value.some((item) => runningStatus.includes(item.status ?? '')))

function formatSize(bytes: number) {
  if (bytes >= 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`
  return `${Math.max(1, Math.round(bytes / 1024))} KB`
}

/** 时间格式化为 yyyy-MM-dd HH:mm:ss，按浏览器本地时区显示。 */
function formatTime(value?: string) {
  if (!value) return '—'
  // 后端统一返回 UTC 时间，可能省略时区后缀，这里补上 Z 再交给 Date 解析，
  // 由浏览器自动换算成本地时区。
  const normalized = value.trim().replace(' ', 'T')
  const withZone = /(Z|[+-]\d{2}:?\d{2})$/.test(normalized) ? normalized : `${normalized}Z`
  const moment = new Date(withZone)
  if (Number.isNaN(moment.getTime())) return value
  const pad = (input: number) => String(input).padStart(2, '0')
  return (
    `${moment.getFullYear()}-${pad(moment.getMonth() + 1)}-${pad(moment.getDate())}` +
    ` ${pad(moment.getHours())}:${pad(moment.getMinutes())}:${pad(moment.getSeconds())}`
  )
}

function meta(status?: string) {
  return statusMeta[status ?? ''] ?? { label: '未知', type: undefined }
}

/** 入库进度百分比，保留两位小数。 */
function percent(item: MissionItem) {
  if (!item.total_count || item.total_count <= 0) return 0
  return Math.min(100, Number((((item.embed_count ?? 0) / item.total_count) * 100).toFixed(2)))
}

/** 进度文本，保留两位小数。 */
function formatPercent(value: number) {
  return `${value.toFixed(2)}%`
}

/** 拉取小说与任务列表。 */
async function refresh(silent = false) {
  if (!session) {
    return
  }
  if (!silent) loading.value = true
  try {
    const [novelContent, missionContent] = await Promise.all([
      request<NovelItem[]>('/novel/filter', { method: 'POST', token: session.token }),
      request<MissionItem[]>('/mission/filter', { method: 'POST', token: session.token })
    ])
    novels.value = Array.isArray(novelContent) ? novelContent : []
    missions.value = Array.isArray(missionContent) ? missionContent : []
  } catch (error) {
    if (!silent) ElMessage.error(error instanceof Error ? error.message : '加载失败')
  } finally {
    loading.value = false
    ensurePolling()
  }
}

/** 提交文件到后端原始文件接口，成功后刷新列表。 */
async function upload(file: LocalFile, raw: File) {
  if (!session) {
    file.status = 'failed'
    file.error = '登录已失效，请重新登录。'
    return
  }

  file.status = 'uploading'
  file.error = undefined

  const form = new FormData()
  form.append('file', raw, raw.name)

  try {
    await request<string>('/origin/upload', {
      method: 'POST',
      token: session.token,
      body: form
    })
    files.value = files.value.filter((item) => item.key !== file.key)
    raws.delete(file.key)
    ElMessage.success(`上传成功：${file.name}`)
    await refresh(true)
  } catch (error) {
    file.status = 'failed'
    file.error = error instanceof Error ? error.message : '上传失败'
    ElMessage.error(`${file.name} 上传失败：${file.error}`)
  }
}

function handleChange(uploadFile: UploadFile) {
  const raw = uploadFile.raw
  if (!raw) return

  const file = reactive<LocalFile>({
    key: ++seq,
    name: raw.name,
    size: raw.size,
    status: 'uploading'
  })
  files.value.unshift(file)
  raws.set(file.key, raw)
  void upload(file, raw)
}

function retryUpload(file: LocalFile) {
  const raw = raws.get(file.key)
  if (!raw) {
    ElMessage.warning('本地文件已释放，请重新选择文件上传。')
    return
  }
  void upload(file, raw)
}

function removeUpload(file: LocalFile) {
  raws.delete(file.key)
  files.value = files.value.filter((item) => item.key !== file.key)
}

/** 修改任务状态，后端以任务标识 id 作为参数。 */
async function act(path: string, mission: MissionItem, message: string) {
  if (!session) return
  try {
    await request<unknown>(path, {
      method: 'POST',
      token: session.token,
      query: { id: mission.id }
    })
    ElMessage.success(message)
    await refresh(true)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '操作失败')
  }
}

function launch(mission: MissionItem) {
  void act('/mission/launch', mission, `已启动：${mission.novel_name ?? mission.id}`)
}

async function cancel(mission: MissionItem) {
  try {
    await ElMessageBox.confirm(
      `取消后当前处理会停止，已完成的解析与切片会保留，下次启动可继续。确定取消《${mission.novel_name ?? mission.id}》吗？`,
      '取消任务',
      { type: 'warning', confirmButtonText: '确定取消', cancelButtonText: '返回' }
    )
  } catch {
    return
  }
  void act('/mission/cancel', mission, `已取消：${mission.novel_name ?? mission.id}`)
}

/** 修改小说名称的对话框。 */
const editVisible = ref(false)
const editing = ref<NovelItem | null>(null)
const editTitle = ref('')
const saving = ref(false)

/** 打开修改名称对话框。 */
function openEdit(novel: NovelItem) {
  editing.value = novel
  editTitle.value = novel.title ?? ''
  editVisible.value = true
}

/** 保存小说名称。 */
async function saveTitle() {
  if (!session || !editing.value) return
  const title = editTitle.value.trim()
  if (!title) {
    ElMessage.warning('请输入小说名称。')
    return
  }
  saving.value = true
  try {
    await request<string>('/novel/modify', {
      method: 'POST',
      token: session.token,
      query: { id: editing.value.id, title }
    })
    ElMessage.success('修改成功')
    editVisible.value = false
    await refresh(true)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '修改失败')
  } finally {
    saving.value = false
  }
}

/** 轮询：只在进行中任务存在时启动，页面隐藏时暂停。 */
let timer: number | undefined

function stopPolling() {
  if (timer !== undefined) {
    window.clearInterval(timer)
    timer = undefined
  }
}

function ensurePolling() {
  if (hasRunning.value && document.visibilityState === 'visible') {
    if (timer === undefined) {
      timer = window.setInterval(() => void refresh(true), 3000)
    }
  } else {
    stopPolling()
  }
}

function onVisibilityChange() {
  ensurePolling()
}

onMounted(() => {
  void refresh()
  document.addEventListener('visibilitychange', onVisibilityChange)
})

onBeforeUnmount(() => {
  stopPolling()
  document.removeEventListener('visibilitychange', onVisibilityChange)
})
</script>

<template>
  <div class="docs">
    <header class="page-head docs-head">
      <div>
        <h1>文档处理</h1>
        <p>上传小说文件，启动处理后可在任务与小说列表中查看进度</p>
      </div>
      <span class="conn-badge"><i />已连接后端</span>
    </header>

    <section class="card">
      <div class="card-head">
        <div>
          <h2>上传文件</h2>
          <p>拖拽或点击选择，支持一次选择多个文件</p>
        </div>
      </div>

      <el-upload
        class="uploader"
        drag
        multiple
        accept=".txt,.md"
        :disabled="!canUpload"
        :auto-upload="false"
        :show-file-list="false"
        :on-change="handleChange"
      >
        <div class="upload-inner">
          <svg class="upload-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4">
            <path d="M12 16V5" stroke-linecap="round" />
            <path d="m7.5 9.5 4.5-4.5 4.5 4.5" stroke-linecap="round" stroke-linejoin="round" />
            <path d="M4.5 15v2.5A2.5 2.5 0 0 0 7 20h10a2.5 2.5 0 0 0 2.5-2.5V15" stroke-linecap="round" />
          </svg>
          <p class="upload-title">拖拽文件到此处，或<em>点击选择</em></p>
          <p class="upload-sub">
            支持 .txt、.md，建议使用 UTF-8 编码，单个文件不超过 50MB
          </p>
        </div>
      </el-upload>

      <div v-if="files.length" class="upload-list">
        <div v-for="file in files" :key="file.key" class="upload-item">
          <span class="upload-name">{{ file.name }}</span>
          <span class="upload-size">{{ formatSize(file.size) }}</span>
          <el-tag v-if="file.status === 'uploading'" type="info" effect="light" size="small" round>上传中</el-tag>
          <el-tooltip v-else :content="file.error ?? '上传失败'" placement="top">
            <el-tag type="danger" effect="light" size="small" round>失败</el-tag>
          </el-tooltip>
          <el-button v-if="file.status === 'failed'" link type="primary" @click="retryUpload(file)">重试</el-button>
          <el-button link type="danger" @click="removeUpload(file)">移除</el-button>
        </div>
      </div>
    </section>

    <section class="card">
      <div class="list-head">
        <el-tabs v-model="tab" class="list-tabs">
          <el-tab-pane :label="`任务（${tasks.length}）`" name="task" />
          <el-tab-pane :label="`小说（${filtered.length}）`" name="novel" />
        </el-tabs>
        <div class="list-tools">
          <el-input
            v-if="tab === 'novel'"
            v-model="keyword"
            placeholder="搜索书名"
            clearable
            size="small"
            class="search"
          />
          <el-button size="small" :loading="loading" @click="refresh()">刷新</el-button>
        </div>
      </div>

      <!-- 任务视图：数据来自任务查询接口 -->
      <el-table v-if="tab === 'task'" v-loading="loading" :data="tasks" row-key="id" class="doc-table">
        <el-table-column label="小说名称" min-width="120">
          <template #default="{ row }">
            <span class="doc-title">{{ row.novel_name || row.novel_code || '—' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="任务状态" width="86">
          <template #default="{ row }">
            <el-tooltip v-if="row.status === '07'" placement="top">
              <span class="doc-sub">{{ meta(row.status).label }}</span>
              <template #content>
                <div>重试次数：{{ row.retry_count ?? 0 }}</div>
                <div>异常消息：{{ row.message || '无' }}</div>
              </template>
            </el-tooltip>
            <span v-else class="doc-sub">{{ meta(row.status).label }}</span>
          </template>
        </el-table-column>

        <el-table-column label="分片总数" width="82" align="right">
          <template #default="{ row }">
            <span class="doc-sub">{{ row.total_count ?? 0 }}</span>
          </template>
        </el-table-column>

        <el-table-column label="入库数量" width="82" align="right">
          <template #default="{ row }">
            <span class="doc-sub">{{ row.embed_count ?? 0 }}</span>
          </template>
        </el-table-column>

        <el-table-column label="入库进度" width="112" align="right">
          <template #default="{ row }">
            <span class="doc-sub">{{ formatPercent(percent(row)) }}</span>
          </template>
        </el-table-column>

        <el-table-column label="启动时间" width="145">
          <template #default="{ row }">
            <span class="doc-sub">{{ row.launch_time ? formatTime(row.launch_time) : '' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="完成时间" width="145">
          <template #default="{ row }">
            <span class="doc-sub">{{ row.finish_time ? formatTime(row.finish_time) : '' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="更新时间" width="145">
          <template #default="{ row }">
            <span class="doc-sub">{{ formatTime(row.update_time) }}</span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="92" align="right">
          <template #default="{ row }">
            <el-button
              v-if="['01', '07', '08'].includes(row.status)"
              link
              type="primary"
              :disabled="!canHandle"
              @click="launch(row)"
            >
              启动
            </el-button>
            <el-button
              v-if="runningStatus.includes(row.status)"
              link
              type="danger"
              :disabled="!canHandle"
              @click="cancel(row)"
            >
              取消
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 小说视图：数据来自小说查询接口 -->
      <el-table v-else v-loading="loading" :data="filtered" row-key="id" class="doc-table">
        <el-table-column label="小说名称" min-width="150">
          <template #default="{ row }">
            <span class="doc-title">{{ row.title || '未命名' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="小说状态" width="80">
          <template #default="{ row }">
            <span class="doc-sub">{{ row.status === '06' ? '就绪' : '等待' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="总字数" width="80" align="right">
          <template #default="{ row }">
            <span class="doc-sub">{{ row.character_count ?? 0 }}</span>
          </template>
        </el-table-column>

        <el-table-column label="章节数" width="80" align="right">
          <template #default="{ row }">
            <span class="doc-sub">{{ row.chapter_count ?? 0 }}</span>
          </template>
        </el-table-column>

        <el-table-column label="创建人" width="100">
          <template #default="{ row }">
            <span class="doc-sub">{{ row.create_name ?? '' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="创建时间" width="145">
          <template #default="{ row }">
            <span class="doc-sub">{{ formatTime(row.create_time) }}</span>
          </template>
        </el-table-column>

        <el-table-column label="更新时间" width="145">
          <template #default="{ row }">
            <span class="doc-sub">{{ formatTime(row.update_time) }}</span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="90" align="right">
          <template #default="{ row }">
            <el-button link type="primary" :disabled="!canModify" @click="openEdit(row)">修改</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty
        v-if="!loading && (tab === 'task' ? tasks.length === 0 : filtered.length === 0)"
        :description="tab === 'task' ? '暂无待处理任务' : '还没有小说，先上传一个文件吧'"
        :image-size="90"
      />
    </section>

    <el-dialog v-model="editVisible" title="修改小说名称" width="400px" :close-on-click-modal="false">
      <el-input
        v-model="editTitle"
        maxlength="200"
        show-word-limit
        placeholder="请输入小说名称"
        @keyup.enter="saveTitle"
      />
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveTitle">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.docs {
  height: 100%;
  overflow-y: auto;
  padding: 22px 32px 40px;
}

.docs-head {
  margin-bottom: 18px;
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
  background: #16a34a;
}

/* 卡片 */
.card {
  max-width: 1080px;
  margin: 0 auto 16px;
  padding: 20px 22px 22px;
  border: 1px solid var(--line);
  border-radius: 16px;
  background: #fff;
}

.card-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 16px;
}

.card-head h2 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
}

.card-head p {
  margin: 5px 0 0;
  color: var(--text-sub);
  font-size: 12.5px;
}

/* 上传区 */
.uploader {
  width: 100%;
}

.uploader :deep(.el-upload) {
  display: block;
  width: 100%;
}

.uploader :deep(.el-upload-dragger) {
  width: 100%;
  padding: 26px 20px;
  border: 1.5px dashed var(--line-strong);
  border-radius: 14px;
  background: #fbfbfe;
  transition: border-color 0.18s ease, background 0.18s ease;
}

.uploader :deep(.el-upload-dragger:hover) {
  border-color: var(--el-color-primary-light-5);
  background: #f7f7fe;
}

.upload-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

.upload-icon {
  width: 34px;
  height: 34px;
  margin-bottom: 4px;
  color: var(--brand);
}

.upload-title {
  margin: 0;
  font-size: 13.5px;
  color: var(--text-main);
}

.upload-title em {
  color: var(--brand);
  font-style: normal;
}

.upload-sub {
  margin: 0;
  color: var(--text-faint);
  font-size: 12px;
}

/* 本地上传条目 */
.upload-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 14px;
}

.upload-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 12px;
  border: 1px solid var(--line);
  border-radius: 10px;
  background: #fbfbfe;
}

.upload-name {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.upload-size {
  color: var(--text-faint);
  font-size: 11.5px;
  white-space: nowrap;
}

/* 列表头部 */
.list-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 8px;
}

.list-tabs {
  flex: 1;
  min-width: 0;
}

.list-tabs :deep(.el-tabs__header) {
  margin: 0;
}

.list-tools {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 0 0 auto;
}

.search {
  width: 180px;
}

/* 表格 */
.doc-table {
  width: 100%;
}

.doc-table :deep(th.el-table__cell) {
  background: #fafbfd;
  font-weight: 500;
  color: var(--text-sub);
}

.doc-table :deep(td.el-table__cell) {
  padding: 12px 0;
}

.doc-title {
  display: block;
  font-size: 13.5px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.doc-sub {
  display: block;
  margin-top: 3px;
  color: var(--text-faint);
  font-size: 11.5px;
  white-space: nowrap;
}

@media (max-width: 900px) {
  .docs {
    padding: 18px 16px 32px;
  }

  .list-head {
    flex-direction: column;
    align-items: stretch;
  }

  .search {
    width: 100%;
  }
}
</style>
