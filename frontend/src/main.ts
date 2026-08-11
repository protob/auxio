import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'

import '@unocss/reset/tailwind-compat.css'
import 'virtual:uno.css'

import '@/styles/tokens.css'
import '@/styles/base.css'

createApp(App).use(createPinia()).use(router).mount('#app')
