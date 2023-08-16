import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import Axios from 'axios'
import App from './App.vue'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

import maoRouter from './mao-router'

console.log("1 ===")

const app = createApp(App);
console.log("2 ===")

for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component)
}
console.log("3 ===")

app.config.globalProperties.$http = Axios.create({
    baseUrl: "https://www.maojianwei.com/resources/",
    timeout: 3000
})
console.log("4 ===")


app.use(ElementPlus, { size: 'large' })
console.log("5 ===")

app.use(maoRouter)
console.log("6 ===")

app.mount('#app')
console.log("7 ===")

