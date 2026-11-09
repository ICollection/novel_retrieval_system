<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getSession, hasPermission, Permission, permissionOptions } from '../auth'
import { request } from '../api'

interface Account {
  id: string
  name: string
  login: string
  permission?: string
}

/** 词元消费，字段与后端 service/compose_service.go 中的 consume 结构一致。 */
interface Consume {
  llm_prompt?: number
  llm_answer?: number
  slm_prompt?: number
  slm_answer?: number
  vem_prompt?: number
}

const session = getSession()
const token = session?.token ?? ''
/** 没有管理用户权限时，增删改按钮全部禁用。 */
const canManage = hasPermission(session?.permission, Permission.ManageUser)
const accounts = ref<Account[]>([])
const loading = ref(false)
const submitting = ref(false)

const addVisible = ref(false)
const editVisible = ref(false)
const passwordVisible = ref(false)
const current = ref<Account | null>(null)

const addFormRef = ref<FormInstance>()
const editFormRef = ref<FormInstance>()
const passwordFormRef = ref<FormInstance>()

const addForm = reactive({ name: '', login: '', password: '', permissions: [] as string[] })
const editForm = reactive({ name: '', permissions: [] as string[] })
const passwordForm = reactive({ password: '' })

const addRules: FormRules<typeof addForm> = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  login: [{ required: true, message: '请输入登录名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const editRules: FormRules<typeof editForm> = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }]
}

const passwordRules: FormRules<typeof passwordForm> = {
  password: [{ required: true, message: '请输入新密码', trigger: 'blur' }]
}

/** 取出错误提示。 */
function messageOf(error: unknown): string {
  return error instanceof Error ? error.message : '操作失败'
}

/** 把数字格式化为带千分号的文本。 */
function thousands(value: number): string {
  return value.toLocaleString('zh-CN')
}

/** 把词元消费整理成提示内容。 */
function consumeHtml(data?: Consume | null): string {
  const llmPrompt = data?.llm_prompt ?? 0
  const llmAnswer = data?.llm_answer ?? 0
  const slmPrompt = data?.slm_prompt ?? 0
  const slmAnswer = data?.slm_answer ?? 0
  const vemPrompt = data?.vem_prompt ?? 0
  const total = llmPrompt + llmAnswer + slmPrompt + slmAnswer + vemPrompt
  return `<table style="width: 100%; border-collapse: collapse; line-height: 1.9">
    <thead>
      <tr style="color: #909399; font-size: 12px">
        <th style="text-align: left; font-weight: 400">模型</th>
        <th style="text-align: right; font-weight: 400">输入</th>
        <th style="text-align: right; font-weight: 400">输出</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td>大语言模型</td>
        <td style="text-align: right">${thousands(llmPrompt)}</td>
        <td style="text-align: right">${thousands(llmAnswer)}</td>
      </tr>
      <tr>
        <td>小语言模型</td>
        <td style="text-align: right">${thousands(slmPrompt)}</td>
        <td style="text-align: right">${thousands(slmAnswer)}</td>
      </tr>
      <tr>
        <td>向量模型</td>
        <td style="text-align: right">${thousands(vemPrompt)}</td>
        <td style="text-align: right">0</td>
      </tr>
      <tr style="border-top: 1px solid #ebeef5">
        <td style="padding-top: 8px">合计</td>
        <td colspan="2" style="padding-top: 8px; text-align: right; font-weight: 600">
          ${thousands(total)}
        </td>
      </tr>
    </tbody>
  </table>`
}

/** 枚举中存在的权限值。 */
const permissionValues = permissionOptions.map((item) => item.value)

/** 把逗号分隔的权限串拆成数组，只保留枚举中存在的权限。 */
function splitPermission(permission?: string): string[] {
  if (!permission) {
    return []
  }
  return permission
    .split(',')
    .map((item) => item.trim())
    .filter((item) => permissionValues.includes(item))
}

/** 把权限编码转成中文名称。 */
function labelOf(value: string): string {
  return permissionOptions.find((item) => item.value === value)?.label ?? value
}

/** 查询用户列表。 */
async function load() {
  if (!token) {
    return
  }
  loading.value = true
  try {
    accounts.value = (await request<Account[] | null>('/account/filterAccount', { token })) ?? []
  } catch (error) {
    ElMessage.error(messageOf(error))
  } finally {
    loading.value = false
  }
}

function openAdd() {
  addForm.name = ''
  addForm.login = ''
  addForm.password = ''
  addForm.permissions = []
  addVisible.value = true
}

function openEdit(row: Account) {
  current.value = row
  editForm.name = row.name
  editForm.permissions = splitPermission(row.permission)
  editVisible.value = true
}

function openPassword(row: Account) {
  current.value = row
  passwordForm.password = ''
  passwordVisible.value = true
}

/** 添加用户。 */
async function submitAdd() {
  const valid = await addFormRef.value?.validate().catch(() => false)
  if (!valid) {
    return
  }
  submitting.value = true
  try {
    await request<string>('/account/appendAccount', {
      method: 'POST',
      token,
      query: {
        name: addForm.name.trim(),
        login: addForm.login.trim(),
        password: addForm.password,
        permission: addForm.permissions.join(',') || undefined
      }
    })
    ElMessage.success('用户已添加')
    addVisible.value = false
    await load()
  } catch (error) {
    ElMessage.error(messageOf(error))
  } finally {
    submitting.value = false
  }
}

/** 修改用户名称和权限。 */
async function submitEdit() {
  if (!current.value) {
    return
  }
  const valid = await editFormRef.value?.validate().catch(() => false)
  if (!valid) {
    return
  }
  submitting.value = true
  try {
    await request<string>('/account/modifyAccount', {
      method: 'PATCH',
      token,
      query: {
        id: current.value.id,
        name: editForm.name.trim(),
        permission: editForm.permissions.join(',') || undefined
      }
    })
    ElMessage.success('用户已修改')
    editVisible.value = false
    await load()
  } catch (error) {
    ElMessage.error(messageOf(error))
  } finally {
    submitting.value = false
  }
}

/** 修改用户密码。 */
async function submitPassword() {
  if (!current.value) {
    return
  }
  const valid = await passwordFormRef.value?.validate().catch(() => false)
  if (!valid) {
    return
  }
  submitting.value = true
  try {
    await request<string>('/account/changeAccount', {
      method: 'PATCH',
      token,
      query: { id: current.value.id, password: passwordForm.password }
    })
    ElMessage.success('密码已修改')
    passwordVisible.value = false
  } catch (error) {
    ElMessage.error(messageOf(error))
  } finally {
    submitting.value = false
  }
}

/** 删除用户。 */
async function remove(row: Account) {
  try {
    await ElMessageBox.confirm(`确定删除用户「${row.name}」吗？`, '删除用户', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    await request<string>('/account/deleteAccount', {
      method: 'DELETE',
      token,
      query: { id: row.id }
    })
    ElMessage.success('用户已删除')
    await load()
  } catch (error) {
    ElMessage.error(messageOf(error))
  }
}

/** 统计用户的词元消费。 */
async function showConsume(row: Account) {
  try {
    const content = await request<Consume | null>('/account/consumeAccount', {
      token,
      query: { id: row.id }
    })
    ElMessageBox.alert(consumeHtml(content), `词元统计 · ${row.name}`, {
      dangerouslyUseHTMLString: true,
      confirmButtonText: '知道了',
      customClass: 'consume-message-box'
    }).catch(() => {})
  } catch (error) {
    ElMessage.error(messageOf(error))
  }
}

onMounted(load)
</script>

<template>
  <div class="users">
    <header class="page-head users-head">
      <div>
        <h1>用户管理</h1>
        <p>管理账号与权限，添加、修改、改密、删除需要「管理用户」权限</p>
      </div>
      <div class="actions">
        <el-button :loading="loading" @click="load">刷新</el-button>
        <el-button type="primary" :disabled="!canManage" @click="openAdd">添加用户</el-button>
      </div>
    </header>

    <section class="card">
      <div class="card-head">
        <div>
          <h2>用户列表</h2>
          <p>共 {{ accounts.length }} 个账号</p>
        </div>
      </div>

      <el-table v-loading="loading" :data="accounts" row-key="id" empty-text="暂无用户">
        <el-table-column prop="name" label="名称" min-width="150" />
        <el-table-column prop="login" label="登录名" min-width="150" />
        <el-table-column label="权限" min-width="220">
          <template #default="{ row }">
            <span v-if="!splitPermission(row.permission).length" class="muted">无</span>
            <span v-else class="permission-list">
              <span v-for="item in splitPermission(row.permission)" :key="item" class="permission">
                {{ labelOf(item) }}
              </span>
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="290" align="right">
          <template #default="{ row }">
            <el-button link type="primary" :disabled="!canManage" @click="showConsume(row)">
              词元统计
            </el-button>
            <el-button link type="primary" :disabled="!canManage" @click="openEdit(row)">修改</el-button>
            <el-button link type="primary" :disabled="!canManage" @click="openPassword(row)">
              改密
            </el-button>
            <el-button link type="danger" :disabled="!canManage" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </section>

    <el-dialog
      v-model="addVisible"
      title="添加用户"
      width="440px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form ref="addFormRef" :model="addForm" :rules="addRules" label-position="top">
        <el-form-item label="名称" prop="name">
          <el-input v-model="addForm.name" placeholder="请输入名称" />
        </el-form-item>
        <el-form-item label="登录名" prop="login">
          <el-input v-model="addForm.login" placeholder="请输入登录名" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="addForm.password" type="password" show-password placeholder="请输入密码" />
        </el-form-item>
        <el-form-item label="权限">
          <el-checkbox-group v-model="addForm.permissions">
            <el-checkbox v-for="item in permissionOptions" :key="item.value" :value="item.value">
              {{ item.label }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitAdd">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="editVisible"
      title="修改用户"
      width="440px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form ref="editFormRef" :model="editForm" :rules="editRules" label-position="top">
        <el-form-item label="名称" prop="name">
          <el-input v-model="editForm.name" placeholder="请输入名称" />
        </el-form-item>
        <el-form-item label="权限">
          <el-checkbox-group v-model="editForm.permissions">
            <el-checkbox v-for="item in permissionOptions" :key="item.value" :value="item.value">
              {{ item.label }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitEdit">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="passwordVisible"
      title="修改密码"
      width="440px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form ref="passwordFormRef" :model="passwordForm" :rules="passwordRules" label-position="top">
        <el-form-item label="新密码" prop="password">
          <el-input
            v-model="passwordForm.password"
            type="password"
            show-password
            placeholder="请输入新密码"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitPassword">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.users {
  height: 100%;
  overflow-y: auto;
  padding: 22px 32px 40px;
}

.users-head {
  margin-bottom: 18px;
}

.actions {
  display: flex;
  gap: 10px;
  flex: 0 0 auto;
}

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
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.card-head h2 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
}

.card-head p {
  margin: 4px 0 0;
  color: var(--text-sub);
  font-size: 12px;
}

.permission-list {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 6px;
}

.permission {
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--brand-soft);
  color: var(--brand-deep);
  font-size: 12px;
}

.muted {
  color: var(--text-faint);
}
</style>
