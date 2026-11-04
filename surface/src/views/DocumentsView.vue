<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { UploadFile } from 'element-plus'
import { request } from '../api'
import { getSession } from '../auth'

type Status = 'uploading' | 'done' | 'failed'
type TagType = 'info' | 'success' | 'danger' | undefined

interface DocFile {
  id: number
  name: string
  size: number
  status: Status
  time: string
  /** 后端返回的原始文件编码 */
  origin?: string
  error?: string
}

const statusMeta: Record<Status, { label: string; type: TagType }> = {
  uploading: { label: '上传中', type: 'info' },
  done: { label: '已上传', type: 'success' },
  failed: { label: '失败', type: 'danger' }
}

const files = ref<DocFile[]>([])
/** 保存本地文件句柄，用于失败重传（不放入响应式数据，避免代理干扰 FormData） */
const raws = new Map<number, File>()
let seq = 0

const now = () =>
  new Date().toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })

const stats = computed(() => ({
  total: files.value.length,
  done: files.value.filter((file) => file.status === 'done').length,
  failed: files.value.filter((file) => file.status === 'failed').length
}))

function formatSize(bytes: number) {
  if (bytes >= 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`
  return `${Math.max(1, Math.round(bytes / 1024))} KB`
}

/** 提交文件到后端原始文件接口。 */
async function upload(doc: DocFile, raw: File) {
  const session = getSession()
  if (!session) {
    doc.status = 'failed'
    doc.error = '登录已失效，请重新登录。'
    return
  }

  doc.status = 'uploading'
  doc.error = undefined

  const form = new FormData()
  form.append('file', raw, raw.name)

  try {
    const origin = await request<string>('/origin/upload', {
      method: 'POST',
      token: session.token,
      body: form
    })
    doc.status = 'done'
    doc.origin = origin
    ElMessage.success(`上传成功：${doc.name}`)
  } catch (error) {
    doc.status = 'failed'
    doc.error = error instanceof Error ? error.message : '上传失败'
    ElMessage.error(`${doc.name} 上传失败：${doc.error}`)
  }
}

function handleChange(uploadFile: UploadFile) {
  const raw = uploadFile.raw
  if (!raw) return

  const doc = reactive<DocFile>({
    id: ++seq,
    name: raw.name,
    size: raw.size,
    status: 'uploading',
    time: now()
  })
  files.value.unshift(doc)
  raws.set(doc.id, raw)
  void upload(doc, raw)
}

function retry(file: DocFile) {
  const raw = raws.get(file.id)
  if (!raw) {
    ElMessage.warning('本地文件已释放，请重新选择文件上传。')
    return
  }
  void upload(file, raw)
}

function remove(file: DocFile) {
  raws.delete(file.id)
  files.value = files.value.filter((item) => item.id !== file.id)
  ElMessage.info(`已移除：${file.name}`)
}
</script>

<template>
  <div class="docs">
    <header class="page-head docs-head">
      <div>
        <h1>文档处理</h1>
        <p>上传小说 txt，文件写入对象存储并登记处理记录</p>
      </div>
      <span class="conn-badge"><i />已连接后端</span>
    </header>

    <section class="stats">
      <div class="stat">
        <span class="stat-label">文件总数</span>
        <strong>{{ stats.total }}</strong>
      </div>
      <div class="stat">
        <span class="stat-label">上传成功</span>
        <strong class="ok">{{ stats.done }}</strong>
      </div>
      <div class="stat">
        <span class="stat-label">上传失败</span>
        <strong class="busy">{{ stats.failed }}</strong>
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
        <span>上传流程</span>
        文件存入 MinIO 的 Origin 目录 → 登记 origin / novel / progress 记录
      </p>
    </section>

    <section class="card">
      <div class="card-head">
        <div>
          <h2>上传记录</h2>
          <p>共 {{ files.length }} 个文件，失败的文件可重新上传</p>
        </div>
      </div>

      <el-table v-if="files.length" :data="files" row-key="id" class="file-table">
        <el-table-column label="文件" min-width="260">
          <template #default="{ row }">
            <div class="file-cell">
              <div class="file-icon">TXT</div>
              <div class="file-copy">
                <span class="file-name">{{ row.name }}</span>
                <span class="file-sub">
                  {{ row.time }} · {{ formatSize(row.size) }}
                  <template v-if="row.error"> · {{ row.error }}</template>
                </span>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusMeta[row.status as Status].type" effect="light" round size="small">
              {{ statusMeta[row.status as Status].label }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="原始文件编码" min-width="300">
          <template #default="{ row }">
            <span class="origin-id">{{ row.origin ?? '—' }}</span>
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
            <el-button link type="danger" @click="remove(row as DocFile)">移除</el-button>
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
  background: #16a34a;
}

/* 统计 */
.stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
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
  color: #dc2626;
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

.origin-id {
  display: block;
  font-family: ui-monospace, SFMono-Regular, Consolas, monospace;
  font-size: 12px;
  color: var(--text-sub);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
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
