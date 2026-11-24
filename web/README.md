# lite-api Web

根目录 web/ 是项目唯一的前端工程，使用 Vue 3 + TypeScript + Vite、Tailwind CSS、Pinia 与 TanStack Vue Query。

## 开发

    npm install
    npm run dev
    npm run typecheck
    npm run build

开发服务器默认监听 http://127.0.0.1:5173，并将 /api 请求代理到 http://127.0.0.1:8080。

目前只有工程状态页；真实用户、渠道与计费界面按 docs/requirements.md 的 S1–S4 阶段建设。shadcn-vue 组件与 Apache ECharts 在业务页面需要时按需引入。
