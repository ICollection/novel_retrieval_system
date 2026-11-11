<script setup lang="ts">
/**
 * 文档处理页面。
 *
 * 页面分三块：上传文件、任务列表（控制启动/取消/重试）、小说列表（内容信息）。
 * 小说来自 GET /prepare/filterNovel，任务来自 GET /prepare/filterMission，前端按 novel 合并后展示。
 */
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadFile } from 'element-plus'
import { getSession, hasPermission, Permission } from '../auth'
import {
  cancelMission,
  deleteNovel,
  filterMission,
  filterNovel,
  launchMission,
  modifyNovel,
  uploadOrigin
} from '../service/prepare'
import type { MissionItem, NovelItem } from '../service/prepare'

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

/** 允许删除的状态：待启动、已完成、已失败、已取消。 */
const deletableStatus = ['01', '06', '07', '08']

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

/** 任务状态底色：处理中、已完成、已失败。 */
const statusTone: Record<string, string> = {
  '02': 'status-running',
  '03': 'status-running',
  '04': 'status-running',
  '05': 'status-running',
  '06': 'status-done',
  '07': 'status-failed'
}

/** 获取任务状态底色样式。 */
function tone(status?: string) {
  return statusTone[status ?? ''] ?? ''
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
    const [novelContent, missionContent] = await Promise.all([filterNovel(), filterMission()])
    novels.value = Array.isArray(novelContent) ? novelContent : []
    missions.value = Array.isArray(missionContent) ? missionContent : []
  } catch (error) {
    if (!silent) ElMessage.error(error instanceof Error ? error.message : '加载失败')
  } finally {
    loading.value = false
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
    await uploadOrigin(form)
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

/** 修改任务状态，action 为启动或取消接口。 */
async function act(action: (id: string) => Promise<void>, mission: MissionItem, message: string) {
  try {
    await action(mission.id)
    ElMessage.success(message)
    await refresh(true)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '操作失败')
  }
}

function launch(mission: MissionItem) {
  void act(launchMission, mission, `已启动：${mission.novel_name ?? mission.id}`)
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
  void act(cancelMission, mission, `已取消：${mission.novel_name ?? mission.id}`)
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
    await modifyNovel(editing.value.id, title)
    ElMessage.success('修改成功')
    editVisible.value = false
    await refresh(true)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '修改失败')
  } finally {
    saving.value = false
  }
}

/** 正在删除的小说标识，避免重复提交。 */
const deleting = ref<string | null>(null)

/** 删除小说，连同任务、章节、分片、向量与原始文件一起清除。 */
async function removeNovel(novel: NovelItem) {
  if (!session) return
  if (!deletableStatus.includes(novel.status ?? '')) {
    ElMessage.warning('任务正在处理，无法删除。')
    return
  }
  try {
    await ElMessageBox.confirm(
      `删除《${novel.title || '未命名'}》会同时清除其任务、章节、分片、向量与原始文件，且不可恢复。确定删除吗？`,
      '删除小说',
      { type: 'warning', confirmButtonText: '确定删除', cancelButtonText: '返回' }
    )
  } catch {
    return
  }
  deleting.value = novel.id
  try {
    await deleteNovel(novel.id)
    ElMessage.success(`已删除：${novel.title || novel.id}`)
    await refresh(true)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '删除失败')
  } finally {
    deleting.value = null
  }
}

/** 轮询：每 1 分钟刷新一次任务列表与小说列表。 */
let timer: number | undefined

onMounted(() => {
  void refresh()
  // 直接启动轮询，不依赖首次请求是否成功
  timer = window.setInterval(() => void refresh(true), 60000)
})

onBeforeUnmount(() => {
  if (timer !== undefined) {
    window.clearInterval(timer)
    timer = undefined
  }
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
        <el-table-column label="小说" min-width="110">
          <template #default="{ row }">
            <span class="doc-title">{{ row.novel_name || row.novel_code || '—' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="状态" min-width="66">
          <template #default="{ row }">
            <el-tooltip v-if="row.status === '07'" placement="top">
              <span :class="['doc-sub', tone(row.status)]">{{ meta(row.status).label }}</span>
              <template #content>
                <div>重试次数：{{ row.retry_count ?? 0 }}</div>
                <div>异常消息：{{ row.message || '无' }}</div>
              </template>
            </el-tooltip>
            <span v-else :class="['doc-sub', tone(row.status)]">{{ meta(row.status).label }}</span>
          </template>
        </el-table-column>

        <el-table-column label="总片数" min-width="66" align="right">
          <template #default="{ row }">
            <span class="doc-sub">{{ row.total_count ?? 0 }}</span>
          </template>
        </el-table-column>

        <el-table-column label="入库数" min-width="66" align="right">
          <template #default="{ row }">
            <span class="doc-sub">{{ row.embed_count ?? 0 }}</span>
          </template>
        </el-table-column>

        <el-table-column label="进度" min-width="76" align="right">
          <template #default="{ row }">
            <span class="doc-sub">{{ formatPercent(percent(row)) }}</span>
          </template>
        </el-table-column>

        <el-table-column label="启动时间" min-width="100">
          <template #default="{ row }">
            <el-tooltip v-if="row.launch_time" placement="top" :content="formatTime(row.launch_time)">
              <span class="doc-sub">{{ formatTime(row.launch_time) }}</span>
            </el-tooltip>
            <span v-else class="doc-sub" />
          </template>
        </el-table-column>

        <el-table-column label="完成时间" min-width="100">
          <template #default="{ row }">
            <el-tooltip v-if="row.finish_time" placement="top" :content="formatTime(row.finish_time)">
              <span class="doc-sub">{{ formatTime(row.finish_time) }}</span>
            </el-tooltip>
            <span v-else class="doc-sub" />
          </template>
        </el-table-column>

        <el-table-column label="更新时间" min-width="100">
          <template #default="{ row }">
            <el-tooltip placement="top" :content="formatTime(row.update_time)">
              <span class="doc-sub">{{ formatTime(row.update_time) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>

        <el-table-column label="操作" min-width="68" align="right">
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
        <el-table-column label="小说" min-width="140">
          <template #default="{ row }">
            <span class="doc-title">{{ row.title || '未命名' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="状态" min-width="70">
          <template #default="{ row }">
            <span class="doc-sub">{{ row.status === '06' ? '就绪' : '等待' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="总字数" min-width="70" align="right">
          <template #default="{ row }">
            <span class="doc-sub">{{ row.character_count ?? 0 }}</span>
          </template>
        </el-table-column>

        <el-table-column label="章节数" min-width="70" align="right">
          <template #default="{ row }">
            <span class="doc-sub">{{ row.chapter_count ?? 0 }}</span>
          </template>
        </el-table-column>

        <el-table-column label="创建用户" min-width="84">
          <template #default="{ row }">
            <span class="doc-sub">{{ row.create_name ?? '' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="创建时间" min-width="100">
          <template #default="{ row }">
            <el-tooltip placement="top" :content="formatTime(row.create_time)">
              <span class="doc-sub">{{ formatTime(row.create_time) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>

        <el-table-column label="更新时间" min-width="100">
          <template #default="{ row }">
            <el-tooltip placement="top" :content="formatTime(row.update_time)">
              <span class="doc-sub">{{ formatTime(row.update_time) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>

        <el-table-column label="操作" min-width="124" align="right">
          <template #default="{ row }">
            <el-button link type="primary" :disabled="!canModify" @click="openEdit(row)">修改</el-button>
            <el-tooltip
              v-if="!deletableStatus.includes(row.status ?? '')"
              content="任务正在处理，无法删除"
              placement="top"
            >
              <span class="action-disabled">
                <el-button link type="danger" disabled>删除</el-button>
              </span>
            </el-tooltip>
            <el-button
              v-else
              link
              type="danger"
              :disabled="!canModify"
              :loading="deleting === row.id"
              @click="removeNovel(row)"
            >
              删除
            </el-button>
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
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 任务状态底色：处理中淡蓝、已完成淡绿、已失败淡红 */
.doc-sub.status-running {
  display: inline-block;
  padding: 1px 6px;
  border-radius: 4px;
  background: #eaf2fd;
  color: #3568b0;
}

.doc-sub.status-done {
  display: inline-block;
  padding: 1px 6px;
  border-radius: 4px;
  background: #e9f7ee;
  color: #2f7a4a;
}

.doc-sub.status-failed {
  display: inline-block;
  padding: 1px 6px;
  border-radius: 4px;
  background: #fdecec;
  color: #b23b3b;
}

/* 列宽自适应，窗口再窄也不显示横向滚动条 */
.doc-table :deep(.el-scrollbar__bar.is-horizontal) {
  display: none;
}

/* 禁用态的删除按钮包裹层，保持与前一按钮的间距 */
.action-disabled {
  display: inline-block;
  margin-left: 12px;
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
