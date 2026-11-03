<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { UploadFile } from 'element-plus'

type Status = 'pending' | 'parsing' | 'chunking' | 'embedding' | 'done' | 'failed'
type TagType = 'info' | 'warning' | 'success' | 'danger' | undefined

interface DocFile {
  id: number
  name: string
  size: number
  chapters: number
  segments: number
  embedded: number
  percent: number
  status: Status
  time: string
}

const statusMeta: Record<Status, { label: string; type: TagType }> = {
  pending: { label: '待处理', type: 'info' },
  parsing: { label: '解析中', type: 'warning' },
  chunking: { label: '切片中', type: 'warning' },
  embedding: { label: '向量化中', type: undefined },
  done: { label: '已完成', type: 'success' },
  failed: { label: '失败', type: 'danger' }
}

const timers = new Map<number, number>()
let seq = 0

const now = () =>
  new Date().toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })

function createDoc(name: string, size: number): DocFile {
  return {
    id: ++seq,
    name,
    size,
    chapters: 0,
    segments: 0,
    embedded: 0,
    percent: 0,
    status: 'pending',
    time: now()
  }
}

const files = ref<DocFile[]>([
  {
    id: ++seq,
    name: '长夜行·全本.txt',
    size: 3.4 * 1024 * 1024,
    chapters: 128,
    segments: 764,
    embedded: 764,
    percent: 100,
    status: 'done',
    time: '11月02日 21:14'
  },
  {
    id: ++seq,
    name: '雾中灯塔.txt',
    size: 1.8 * 1024 * 1024,
    chapters: 74,
    segments: 442,
    embedded: 274,
    percent: 62,
    status: 'embedding',
    time: '11月03日 09:36'
  },
  {
    id: ++seq,
    name: '旧城纪事.txt',
    size: 0.9 * 1024 * 1024,
    chapters: 12,
    segments: 68,
    embedded: 0,
    percent: 18,
    status: 'failed',
    time: '11月03日 10:02'
  }
])

const stats = computed(() => {
  const inProgress = files.value.filter((file) =>
    ['pending', 'parsing', 'chunking', 'embedding'].includes(file.status)
  ).length
  return {
    total: files.value.length,
    done: files.value.filter((file) => file.status === 'done').length,
    inProgress,
    segments: files.value.reduce((sum, file) => sum + file.segments, 0)
  }
})

function formatSize(bytes: number) {
  if (bytes >= 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`
  return `${Math.max(1, Math.round(bytes / 1024))} KB`
}

function stopPipeline(id: number) {
  const timer = timers.get(id)
  if (timer !== undefined) {
    window.clearInterval(timer)
    timers.delete(id)
  }
}

function runPipeline(file: DocFile) {
  stopPipeline(file.id)
  file.status = 'parsing'
  file.percent = 2

  const timer = window.setInterval(() => {
    file.percent = Math.min(100, file.percent + Math.random() * 8 + 3)

    if (file.percent < 30) {
      file.status = 'parsing'
    } else if (file.percent < 55) {
      if (!file.chapters) {
        file.chapters = 40 + Math.round(Math.random() * 80)
        file.segments = file.chapters * (5 + Math.round(Math.random() * 3))
      }
      file.status = 'chunking'
    } else if (file.percent < 100) {
      file.status = 'embedding'
      file.embedded = Math.round((file.segments * (file.percent - 55)) / 45)
    } else {
      file.percent = 100
      file.embedded = file.segments
      file.status = 'done'
      stopPipeline(file.id)
      ElMessage.success(`${file.name} 处理完成`)
    }
  }, 700)

  timers.set(file.id, timer)
}

function handleChange(uploadFile: UploadFile) {
  const raw = uploadFile.raw
  if (!raw) return

  const doc = createDoc(raw.name, raw.size)
  files.value.unshift(doc)
  runPipeline(doc)
  ElMessage.success(`已加入处理队列：${raw.name}`)
}

function retry(file: DocFile) {
  file.embedded = 0
  file.percent = 0
  file.status = 'pending'
  runPipeline(file)
}

function remove(file: DocFile) {
  stopPipeline(file.id)
  files.value = files.value.filter((item) => item.id !== file.id)
  ElMessage.info(`已移除：${file.name}`)
}

function progressStatus(file: DocFile): '' | 'success' | 'exception' {
  if (file.status === 'done') return 'success'
  if (file.status === 'failed') return 'exception'
  return ''
}

onBeforeUnmount(() => {
  timers.forEach((timer) => window.clearInterval(timer))
  timers.clear()
})
</script>

<template>
  <div class="docs">
    <header class="page-head docs-head">
      <div>
        <h1>文档处理</h1>
        <p>上传小说 txt，查看解析、切片与向量化的处理状态</p>
      </div>
      <span class="conn-badge"><i />界面演示</span>
    </header>

    <section class="stats">
      <div class="stat">
        <span class="stat-label">文件总数</span>
        <strong>{{ stats.total }}</strong>
      </div>
      <div class="stat">
        <span class="stat-label">已完成</span>
        <strong class="ok">{{ stats.done }}</strong>
      </div>
      <div class="stat">
        <span class="stat-label">处理中</span>
        <strong class="busy">{{ stats.inProgress }}</strong>
      </div>
      <div class="stat">
        <span class="stat-label">切片总量</span>
        <strong>{{ stats.segments }}</strong>
      </div>
    </section>

    <section class="card">
      <div class="card-head">
        <div>
          <h2>上传文件</h2>
          <p>拖拽或点击选择，支持一次选择多个 txt 文件</p>
        </div>
      </div>

      <el-upload
        class="uploader"
        drag
        multiple
        accept=".txt"
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
          <p class="upload-sub">仅支持 .txt，建议使用 UTF-8 编码，单个文件不超过 50MB</p>
        </div>
      </el-upload>

      <p class="flow-hint">
        <span>处理流程</span>
        解析章节 → 文本切片 → 向量化入库，状态与进度会自动刷新
      </p>
    </section>

    <section class="card">
      <div class="card-head">
        <div>
          <h2>文件列表</h2>
          <p>共 {{ files.length }} 个文件，失败文件可重试</p>
        </div>
      </div>

      <el-table v-if="files.length" :data="files" row-key="id" class="file-table">
        <el-table-column label="文件" min-width="260">
          <template #default="{ row }">
            <div class="file-cell">
              <div class="file-icon">TXT</div>
              <div class="file-copy">
                <span class="file-name">{{ row.name }}</span>
                <span class="file-sub">{{ row.time }} · {{ formatSize(row.size) }}</span>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="内容规模" min-width="150">
          <template #default="{ row }">
            <div class="scale-cell">
              <span>{{ row.chapters ? `${row.chapters} 章` : '—' }}</span>
              <i>·</i>
              <span>{{ row.segments ? `${row.segments} 切片` : '待切片' }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusMeta[row.status as Status].type" effect="light" round size="small">
              {{ statusMeta[row.status as Status].label }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="进度" min-width="180">
          <template #default="{ row }">
            <div class="progress-cell">
              <el-progress
                :percentage="Math.round(row.percent)"
                :status="progressStatus(row as DocFile)"
                :stroke-width="6"
                :show-text="false"
              />
              <span class="progress-text">
                {{ Math.round(row.percent) }}%
                <template v-if="row.status === 'embedding'">
                  ({{ row.embedded }}/{{ row.segments }})
                </template>
              </span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="130" align="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'failed'"
              link
              type="primary"
              @click="retry(row as DocFile)"
            >
              重试
            </el-button>
            <el-button link type="danger" @click="remove(row as DocFile)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-else description="还没有上传文件，先选择一个 txt 小说吧" :image-size="90" />
    </section>
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
  background: #f0b429;
}

/* 统计 */
.stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
  max-width: 1080px;
  margin: 0 auto 16px;
}

.stat {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 16px 18px;
  border: 1px solid var(--line);
  border-radius: 14px;
  background: #fff;
}

.stat-label {
  color: var(--text-sub);
  font-size: 12px;
}

.stat strong {
  font-size: 24px;
  font-weight: 600;
  letter-spacing: 0.4px;
}

.stat strong.ok {
  color: #16a34a;
}

.stat strong.busy {
  color: var(--brand);
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

.flow-hint {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 14px 0 0;
  color: var(--text-sub);
  font-size: 12px;
}

.flow-hint span {
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--brand-soft);
  color: var(--brand);
  font-size: 11.5px;
}

/* 文件表格 */
.file-table {
  width: 100%;
}

.file-table :deep(th.el-table__cell) {
  background: #fafbfd;
  font-weight: 500;
  color: var(--text-sub);
}

.file-table :deep(td.el-table__cell) {
  padding: 13px 0;
}

.file-cell {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.file-icon {
  display: grid;
  place-items: center;
  width: 34px;
  height: 34px;
  flex: 0 0 34px;
  border-radius: 9px;
  background: var(--brand-soft);
  color: var(--brand);
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.4px;
}

.file-copy {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.file-name {
  font-size: 13.5px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-sub {
  color: var(--text-faint);
  font-size: 11.5px;
}

.scale-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--text-sub);
  font-size: 12.5px;
}

.scale-cell i {
  color: var(--line-strong);
  font-style: normal;
}

.progress-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.progress-cell :deep(.el-progress) {
  flex: 1;
  min-width: 60px;
}

.progress-text {
  width: 104px;
  flex: 0 0 auto;
  color: var(--text-sub);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  text-align: right;
  white-space: nowrap;
}

@media (max-width: 900px) {
  .stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .docs {
    padding: 18px 16px 32px;
  }
}
</style>
