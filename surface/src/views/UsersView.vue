<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getSession, hasPermission } from '../auth'
import { request } from '../api'

interface Account {
  id: string
  name: string
  login: string
  permission?: string
}

/** 与后端 server/identity.go 中的权限常量保持一致。 */
const permissions = [
  { value: 'ManageUser', label: '管理用户' },
  { value: 'UploadNovel', label: '上传小说' },
  { value: 'HandleNovel', label: '处理小说' }
]

const session = getSession()
const token = session?.token ?? ''
/** 没有管理用户权限时，增删改按钮全部禁用。 */
const canManage = hasPermission(session?.permission, 'ManageUser')
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

/** 把逗号分隔的权限串拆成数组。 */
function splitPermission(permission?: string): string[] {
  return permission ? permission.split(',').filter(Boolean) : []
}

/** 把权限编码转成中文名称。 */
function labelOf(value: string): string {
  return permissions.find((item) => item.value === value)?.label ?? value
}

/** 查询用户列表。 */
async function load() {
  if (!token) {
    return
  }
  loading.value = true
  try {
    accounts.value = (await request<Account[] | null>('/account/filter', { token })) ?? []
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
    await request<string>('/account/append', {
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
    await request<string>('/account/modify', {
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
    await request<string>('/account/change', {
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
    await request<string>('/account/delete', {
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
        <el-table-column label="操作" width="210" align="right">
          <template #default="{ row }">
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
            <el-checkbox v-for="item in permissions" :key="item.value" :value="item.value">
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
            <el-checkbox v-for="item in permissions" :key="item.value" :value="item.value">
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
