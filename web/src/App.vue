<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'

interface SystemInfo {
  name: string
  status: 'scaffold'
  api_version: string
}

const { data, isPending, isError } = useQuery<SystemInfo>({
  queryKey: ['system', 'info'],
  queryFn: async () => {
    const response = await fetch('/api/v1/system/info')
    if (!response.ok) {
      throw new Error('Backend returned HTTP ' + response.status)
    }
    return (await response.json()) as SystemInfo
  },
  retry: false
})

const modules = [
  { label: '协议适配', description: 'OpenAI、Anthropic、Gemini 原生入口与流式透传', stage: 'S1–S3' },
  { label: '账号调度', description: '模型映射、权重路由、租约、限流和故障切换', stage: 'S1–S2' },
  { label: '额度计费', description: '可审计的预留、结算与账本幂等性', stage: 'S1–S2' },
  { label: '管理控制台', description: '渠道、模型、用户、密钥与真实用量统计', stage: 'S1–S2' }
]
</script>

<template>
  <div class="min-h-screen bg-slate-950 text-slate-100">
    <header class="border-b border-white/10">
      <div class="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-4 px-6 py-5">
        <div class="flex items-center gap-3">
          <span class="flex h-10 w-10 items-center justify-center rounded-xl bg-indigo-500 text-lg font-bold text-white">L</span>
          <div>
            <p class="text-lg font-semibold tracking-tight">lite-api</p>
            <p class="text-xs text-slate-400">AI Gateway · Initial Scaffold</p>
          </div>
        </div>
        <span class="rounded-full border border-white/15 px-3 py-1.5 text-xs text-slate-300">工程初始化 · S0</span>
      </div>
    </header>

    <main class="mx-auto max-w-6xl px-6 py-14">
      <section class="max-w-3xl">
        <p class="mb-4 text-sm font-medium tracking-widest text-indigo-300">LITE-API / PROJECT STATUS</p>
        <h1 class="text-4xl font-semibold tracking-tight sm:text-5xl">统一模型入口，从可靠的网关基础开始。</h1>
        <p class="mt-6 text-base leading-8 text-slate-400">
          当前仓库已经具备前后端工程骨架。模型转发、账号池、密钥管理和计费仍在规划与实现阶段；
          本页面只展示真实的工程状态，不模拟尚未接入的业务数据。
        </p>
      </section>

      <section class="mt-12 grid gap-4 sm:grid-cols-2">
        <div class="rounded-2xl border border-white/10 bg-white/5 p-6">
          <p class="text-sm text-slate-400">后端连接</p>
          <div class="mt-4 flex items-center gap-3">
            <span
              class="h-2.5 w-2.5 rounded-full"
              :class="isPending ? 'bg-amber-400' : isError ? 'bg-rose-400' : 'bg-emerald-400'"
            />
            <p class="text-xl font-semibold">
              {{ isPending ? '正在检测' : isError ? '无法连接' : 'HTTP 服务在线' }}
            </p>
          </div>
          <p class="mt-3 text-sm text-slate-400">
            {{ isError ? '检查 Go 服务是否运行在 127.0.0.1:8080。' : '检查接口：/api/v1/system/info' }}
          </p>
        </div>
        <div class="rounded-2xl border border-white/10 bg-white/5 p-6">
          <p class="text-sm text-slate-400">当前交付阶段</p>
          <p class="mt-4 text-xl font-semibold">S0 · 基础工程</p>
          <p class="mt-3 text-sm text-slate-400">
            {{ data ? '接口版本：' + data.api_version + ' · 状态：' + data.status : 'Go + Vue 3 + PostgreSQL + Redis' }}
          </p>
        </div>
      </section>

      <div class="mt-14 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h2 class="text-2xl font-semibold">后续能力</h2>
          <p class="mt-2 text-sm text-slate-400">按需求文档分阶段交付，尚未实现的功能不会显示为可用。</p>
        </div>
        <a
          class="text-sm font-medium text-indigo-300 underline underline-offset-4 hover:text-indigo-200"
          href="https://github.com/CareyChi/lite-api/blob/main/docs/requirements.md"
          target="_blank"
          rel="noopener noreferrer"
        >查看需求文档</a>
      </div>

      <section class="mt-6 grid gap-4 sm:grid-cols-2">
        <article v-for="item in modules" :key="item.label" class="rounded-2xl border border-white/10 p-6">
          <div class="flex items-start justify-between gap-3">
            <h3 class="text-lg font-medium">{{ item.label }}</h3>
            <span class="rounded-md bg-white/10 px-2 py-1 text-xs text-slate-300">{{ item.stage }}</span>
          </div>
          <p class="mt-3 text-sm leading-6 text-slate-400">{{ item.description }}</p>
          <p class="mt-5 text-xs text-amber-200/80">尚未实现</p>
        </article>
      </section>
    </main>
  </div>
</template>
