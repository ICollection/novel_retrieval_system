<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'

const route = useRoute()
const activeKey = computed(() => (route.path.startsWith('/documents') ? 'documents' : 'chat'))
</script>

<template>
  <div class="shell">
    <aside class="sidebar">
      <div class="brand">
        <div class="brand-mark">问</div>
        <div class="brand-text">
          <strong>小说问答</strong>
          <span>NOVEL Q&amp;A</span>
        </div>
      </div>

      <nav class="nav">
        <p class="nav-label">工作区</p>

        <RouterLink to="/" class="nav-item" :class="{ 'is-active': activeKey === 'chat' }">
          <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6">
            <path
              d="M20 12a7.5 7.5 0 0 1-7.5 7.5H9L4.5 22v-4.2A7.5 7.5 0 0 1 4 12a7.5 7.5 0 0 1 7.5-7.5H12A7.5 7.5 0 0 1 20 12Z"
              stroke-linejoin="round"
            />
            <path d="M9 11h7M9 14.5h4.5" stroke-linecap="round" />
          </svg>
          <span class="nav-copy">
            <b>对话问答</b>
            <i>基于小说内容检索回答</i>
          </span>
        </RouterLink>

        <RouterLink
          to="/documents"
          class="nav-item"
          :class="{ 'is-active': activeKey === 'documents' }"
        >
          <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6">
            <path d="M13 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V9l-6-6Z" stroke-linejoin="round" />
            <path d="M13 3v6h6" stroke-linejoin="round" />
            <path d="M9 13.5h6M9 17h4" stroke-linecap="round" />
          </svg>
          <span class="nav-copy">
            <b>文档处理</b>
            <i>上传小说并查看处理状态</i>
          </span>
        </RouterLink>
      </nav>

      <div class="sidebar-foot">
        <span class="dot" />
        <span>本地演示 · 未连接后端</span>
      </div>
    </aside>

    <main class="main">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.shell {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background: var(--bg-page);
}

/* 左侧导航 */
.sidebar {
  position: relative;
  display: flex;
  flex-direction: column;
  width: 236px;
  flex: 0 0 236px;
  padding: 22px 16px 18px;
  background: linear-gradient(180deg, #1b202c 0%, #171b25 100%);
  color: #cbd2e1;
}

.brand {
  display: flex;
  align-items: center;
  gap: 11px;
  padding: 0 6px 22px;
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

.nav {
  flex: 1;
  min-height: 0;
}

.nav-label {
  margin: 6px 10px 10px;
  color: #5f6879;
  font-size: 11px;
  letter-spacing: 1.2px;
}

.nav-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 11px;
  padding: 11px 12px;
  margin-bottom: 6px;
  border-radius: 10px;
  color: #b6bdcc;
  text-decoration: none;
  transition: background 0.18s ease, color 0.18s ease;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.06);
  color: #e6eaf3;
}

.nav-item.is-active {
  background: rgba(255, 255, 255, 0.09);
  color: #ffffff;
}

.nav-item.is-active::before {
  content: '';
  position: absolute;
  left: -16px;
  top: 50%;
  width: 3px;
  height: 22px;
  border-radius: 0 3px 3px 0;
  background: #7c7cf0;
  transform: translateY(-50%);
}

.nav-icon {
  width: 20px;
  height: 20px;
  flex: 0 0 20px;
}

.nav-copy {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.nav-copy b {
  font-size: 13.5px;
  font-weight: 500;
}

.nav-copy i {
  color: #757f92;
  font-size: 11px;
  font-style: normal;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.nav-item.is-active .nav-copy i {
  color: #9aa3b6;
}

.sidebar-foot {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 10px 2px;
  border-top: 1px solid rgba(255, 255, 255, 0.07);
  color: #6d7688;
  font-size: 11.5px;
}

.dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #4ade80;
  box-shadow: 0 0 0 3px rgba(74, 222, 128, 0.15);
}

/* 右侧内容区 */
.main {
  flex: 1;
  min-width: 0;
  height: 100vh;
  overflow: hidden;
}
</style>
