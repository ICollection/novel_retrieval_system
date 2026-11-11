import './assets/main.css'
import 'element-plus/dist/index.css'

import { createApp } from 'vue'
import App from './App.vue'
import { onTokenRenewed, onUnauthorized } from './api'
import { signOut, updateToken } from './auth'
import router from './router'

// 后端续签令牌时替换本地令牌
onTokenRenewed(updateToken)

// 令牌过期或失效时清除本地登录态，跳转登录页并记录来源
onUnauthorized(() => {
  signOut()
  if (router.currentRoute.value.name === 'login') {
    return
  }
  const current = router.currentRoute.value.fullPath
  router.replace({ name: 'login', query: current === '/' ? {} : { redirect: current } })
})

const app = createApp(App)

app.use(router)

app.mount('#app')
