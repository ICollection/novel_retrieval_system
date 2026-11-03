<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { signIn } from '../auth'

const router = useRouter()
const route = useRoute()

const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

const rules: FormRules<typeof form> = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

async function submit() {
  if (!formRef.value) {
    return
  }
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) {
    return
  }

  loading.value = true
  try {
    // 演示用：模拟请求耗时，暂不调用后端
    await new Promise((resolve) => setTimeout(resolve, 350))
    const session = signIn(form.username, form.password)
    ElMessage.success(`欢迎回来，${session.username}`)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    await router.replace(redirect)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <section class="intro">
        <div class="brand">
          <div class="brand-mark">问</div>
          <div class="brand-text">
            <strong>小说问答</strong>
            <span>NOVEL Q&amp;A</span>
          </div>
        </div>

        <div class="intro-copy">
          <h2>读懂每一本小说</h2>
          <p>上传小说文件，构建专属知识库，用对话的方式检索情节、人物与细节。</p>
        </div>

        <ul class="intro-list">
          <li>对话式检索，回答附带章节出处</li>
          <li>文档解析进度实时可见</li>
          <li>本地部署，数据留在自己的服务器</li>
        </ul>
      </section>

      <section class="panel">
        <div class="panel-head">
          <h1>登录</h1>
          <p>请输入账号信息以继续使用</p>
        </div>

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-position="top"
          size="large"
          @submit.prevent="submit"
        >
          <el-form-item label="用户名" prop="username">
            <el-input v-model="form.username" placeholder="请输入用户名" autocomplete="username" />
          </el-form-item>

          <el-form-item label="密码" prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
              show-password
              autocomplete="current-password"
              @keyup.enter="submit"
            />
          </el-form-item>

          <el-button class="submit" type="primary" size="large" :loading="loading" @click="submit">
            登录
          </el-button>
        </el-form>

        <p class="hint">当前为演示环境，未连接后端，任意非空账号密码均可登录。</p>
      </section>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  display: grid;
  place-items: center;
  min-height: 100vh;
  padding: 32px 20px;
  background:
    radial-gradient(900px 420px at 12% 8%, #e9e9fb 0%, transparent 60%),
    radial-gradient(700px 380px at 92% 92%, #e6eefb 0%, transparent 62%),
    var(--bg-page);
}

.login-card {
  display: grid;
  grid-template-columns: 340px 400px;
  width: 100%;
  max-width: 740px;
  overflow: hidden;
  border-radius: 18px;
  background: var(--bg-card);
  box-shadow:
    0 24px 60px rgba(31, 36, 48, 0.12),
    0 2px 8px rgba(31, 36, 48, 0.04);
}

/* 左侧介绍 */
.intro {
  display: flex;
  flex-direction: column;
  gap: 28px;
  padding: 32px 28px;
  background: linear-gradient(160deg, #1b202c 0%, #171b25 100%);
  color: #cbd2e1;
}

.brand {
  display: flex;
  align-items: center;
  gap: 11px;
}

.brand-mark {
  display: grid;
  place-items: center;
  width: 38px;
  height: 38px;
  border-radius: 11px;
  background: linear-gradient(135deg, #6f6fe8 0%, #4a4ab8 100%);
  color: #fff;
  font-size: 18px;
  font-weight: 600;
  box-shadow: 0 6px 16px rgba(91, 91, 214, 0.35);
}

.brand-text {
  display: flex;
  flex-direction: column;
  gap: 3px;
  line-height: 1;
}

.brand-text strong {
  color: #f4f6fb;
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.brand-text span {
  color: #6d7688;
  font-size: 10px;
  letter-spacing: 1.6px;
}

.intro-copy h2 {
  margin: 0 0 10px;
  color: #f4f6fb;
  font-size: 20px;
  font-weight: 600;
  line-height: 1.4;
}

.intro-copy p {
  margin: 0;
  color: #8c95a8;
  font-size: 13px;
  line-height: 1.75;
}

.intro-list {
  margin: auto 0 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.intro-list li {
  position: relative;
  padding-left: 18px;
  color: #8c95a8;
  font-size: 12.5px;
  line-height: 1.6;
}

.intro-list li::before {
  content: '';
  position: absolute;
  left: 0;
  top: 6px;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #7c7cf0;
}

/* 右侧表单 */
.panel {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 40px 38px;
}

.panel-head h1 {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
}

.panel-head p {
  margin: 8px 0 26px;
  color: var(--text-sub);
  font-size: 13px;
}

.submit {
  width: 100%;
  margin-top: 4px;
}

.hint {
  margin: 22px 0 0;
  color: var(--text-faint);
  font-size: 12px;
  line-height: 1.7;
}

@media (max-width: 720px) {
  .login-card {
    grid-template-columns: 1fr;
    max-width: 400px;
  }

  .intro {
    gap: 18px;
    padding: 26px 24px;
  }

  .intro-list {
    display: none;
  }

  .panel {
    padding: 30px 24px 32px;
  }
}
</style>
