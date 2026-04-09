<template>
  <div class="min-h-screen overflow-auto px-8 py-6">
    <section class="mb-6 rounded-3xl border border-neutral-800 bg-neutral-900/80 p-6">
      <div class="flex flex-col gap-6 xl:flex-row xl:items-end xl:justify-between">
        <div class="max-w-3xl">
          <p class="mb-2 text-xs uppercase tracking-[0.3em] text-sky-300/80">Online Sources</p>
          <h1 class="text-2xl font-semibold text-white">在线漫画库</h1>
          <p class="mt-2 text-sm leading-6 text-neutral-400">
            这是 ImageMaster 在线源系统的第一版样板。当前先做内置源，跑通“选源 -> 搜索 -> 详情”的闭环，
            后面再把源逐步外置成独立源文件。
          </p>
        </div>

        <div class="w-full max-w-xl">
          <Input
            v-model="searchKeyword"
            help="当前版本先支持搜索和包子漫画详情。后续会继续接章节阅读和下载到本地漫画库。"
            placeholder="输入关键字搜索在线漫画源"
            @keydown.enter="runSearch(1)"
          />
        </div>
      </div>

      <div class="mt-5 flex flex-wrap gap-3">
        <button
          v-for="source in sources"
          :key="source.id"
          class="cursor-pointer rounded-2xl border px-4 py-3 text-left transition-colors"
          :class="
            selectedSourceId === source.id
              ? 'border-sky-400/70 bg-sky-500/10 text-sky-100'
              : 'border-neutral-800 bg-neutral-950/60 text-neutral-300 hover:bg-neutral-800'
          "
          @click="selectedSourceId = source.id"
        >
          <div class="font-medium">{{ source.name }}</div>
          <div class="mt-1 text-xs opacity-80">{{ source.language }} / {{ source.type }}</div>
        </button>
      </div>

      <div class="mt-5 flex flex-wrap gap-3">
        <button
          class="cursor-pointer rounded-xl border border-sky-500/60 bg-sky-500/10 px-4 py-2 text-sm text-sky-100 transition-colors hover:bg-sky-500/20 disabled:cursor-not-allowed disabled:opacity-50"
          :disabled="searching || !selectedSourceId"
          @click="runSearch(1)"
        >
          {{ searching ? '搜索中...' : '开始搜索' }}
        </button>

        <button
          v-if="result.items.length > 0"
          class="cursor-pointer rounded-xl border border-neutral-700 px-4 py-2 text-sm text-neutral-200 transition-colors hover:bg-neutral-800"
          @click="clearResult"
        >
          清空结果
        </button>
      </div>
    </section>

    <section class="mb-6 grid gap-4 lg:grid-cols-[2fr_1fr]">
      <div class="rounded-3xl border border-neutral-800 bg-neutral-900/80 p-6">
        <div class="mb-4 flex items-center justify-between gap-4">
          <div>
            <h2 class="text-lg font-semibold text-white">搜索结果</h2>
            <p class="mt-1 text-sm text-neutral-400">{{ resultCaption }}</p>
          </div>

          <button
            v-if="result.hasMore"
            class="cursor-pointer rounded-xl border border-neutral-700 px-4 py-2 text-sm text-neutral-200 transition-colors hover:bg-neutral-800 disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="searching"
            @click="runSearch(result.page + 1)"
          >
            下一页
          </button>
        </div>

        <div
          v-if="errorMessage"
          class="rounded-2xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-200"
        >
          {{ errorMessage }}
        </div>

        <div
          v-else-if="result.items.length === 0"
          class="rounded-2xl border border-dashed border-neutral-800 px-6 py-12 text-center text-sm text-neutral-500"
        >
          先选择一个源并搜索关键字。当前版本先展示搜索结果，后续会继续接详情、章节与下载。
        </div>

        <div v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <article
            v-for="item in result.items"
            :key="item.id"
            class="overflow-hidden rounded-2xl border border-neutral-800 bg-neutral-950/60"
          >
            <div class="aspect-[3/4] overflow-hidden bg-neutral-900">
              <img
                v-if="item.cover"
                :src="item.cover"
                :alt="item.title"
                class="h-full w-full object-cover"
                loading="lazy"
              />
              <div v-else class="flex h-full items-center justify-center text-xs text-neutral-600">
                No Cover
              </div>
            </div>
            <div class="flex flex-col gap-2 p-4">
              <h3 class="line-clamp-2 text-sm font-semibold text-white">{{ item.title }}</h3>
              <p class="text-xs text-neutral-400">{{ item.primaryLabel }}</p>
              <p v-if="item.secondaryLabel" class="text-xs text-neutral-500">{{ item.secondaryLabel }}</p>
              <p class="line-clamp-3 text-xs leading-5 text-neutral-400">{{ item.summary }}</p>
              <div class="mt-2 flex flex-wrap gap-2">
                <button
                  class="cursor-pointer rounded-lg border border-sky-500/50 bg-sky-500/10 px-3 py-2 text-xs text-sky-100 transition-colors hover:bg-sky-500/20"
                  @click="openDetail(item)"
                >
                  在软件内查看
                </button>
                <a
                  v-if="item.detailUrl"
                  :href="item.detailUrl"
                  target="_blank"
                  rel="noreferrer"
                  class="rounded-lg border border-neutral-700 px-3 py-2 text-xs text-neutral-300 transition-colors hover:bg-neutral-800"
                >
                  打开源站
                </a>
              </div>
            </div>
          </article>
        </div>
      </div>

      <aside class="rounded-3xl border border-neutral-800 bg-neutral-900/80 p-6">
        <h2 class="text-lg font-semibold text-white">当前方案</h2>
        <ul class="mt-4 space-y-3 text-sm leading-6 text-neutral-400">
          <li>先内置少量源，验证在线源架构是否顺手。</li>
          <li>下载逻辑继续留在软件本体，源只负责解析。</li>
          <li>等搜索、详情、章节稳定后，再外置为独立源文件。</li>
          <li>后续如果你满意，再做“安装源 / 更新源”这一步。</li>
        </ul>

        <div class="mt-6 rounded-2xl border border-neutral-800 bg-neutral-950/70 p-4">
          <div class="text-xs uppercase tracking-[0.28em] text-neutral-500">Selected Source</div>
          <div class="mt-2 text-base font-medium text-white">{{ selectedSource?.name ?? '未选择' }}</div>
          <p class="mt-2 text-sm text-neutral-400">
            {{ selectedSource?.description ?? '请选择一个源开始。' }}
          </p>
          <a
            v-if="selectedSource?.website"
            :href="selectedSource.website"
            target="_blank"
            rel="noreferrer"
            class="mt-3 inline-block text-xs text-sky-300 hover:text-sky-200"
          >
            访问源站
          </a>
        </div>
      </aside>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Input } from '@/components'
import { ListSources, SearchSources } from '../../../wailsjs/go/source/API'

interface SourceSummary {
  id: string
  name: string
  type: string
  language: string
  website: string
  description: string
}

interface SearchItem {
  id: string
  title: string
  cover: string
  summary: string
  primaryLabel: string
  secondaryLabel: string
  detailUrl: string
}

interface SearchResult {
  source: SourceSummary
  query: string
  page: number
  hasMore: boolean
  total: number
  items: SearchItem[]
}

const router = useRouter()
const sources = ref<SourceSummary[]>([])
const selectedSourceId = ref('')
const searchKeyword = ref('')
const searching = ref(false)
const errorMessage = ref('')
const result = ref<SearchResult>({
  source: {
    id: '',
    name: '',
    type: '',
    language: '',
    website: '',
    description: '',
  },
  query: '',
  page: 1,
  hasMore: false,
  total: 0,
  items: [],
})

const selectedSource = computed(
  () => sources.value.find((source) => source.id === selectedSourceId.value) ?? null,
)

const resultCaption = computed(() => {
  if (!result.value.query) {
    return '当前还没有搜索结果。'
  }

  return `源：${result.value.source.name} / 关键字：${result.value.query} / 共 ${result.value.total} 条`
})

onMounted(async () => {
  sources.value = await ListSources()
  selectedSourceId.value = sources.value[0]?.id ?? ''
})

async function runSearch(page: number) {
  if (!selectedSourceId.value || !searchKeyword.value.trim()) {
    return
  }

  searching.value = true
  errorMessage.value = ''

  try {
    result.value = await SearchSources(selectedSourceId.value, searchKeyword.value.trim(), page)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '搜索失败，请稍后再试。'
  } finally {
    searching.value = false
  }
}

function openDetail(item: SearchItem) {
  router.push({
    path: '/online/detail',
    query: {
      source: result.value.source.id,
      id: item.id,
    },
  })
}

function clearResult() {
  errorMessage.value = ''
  result.value = {
    source: {
      id: '',
      name: '',
      type: '',
      language: '',
      website: '',
      description: '',
    },
    query: '',
    page: 1,
    hasMore: false,
    total: 0,
    items: [],
  }
}
</script>

<style scoped></style>
