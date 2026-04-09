<template>
  <div class="min-h-screen overflow-y-auto px-8 py-6">
    <section class="mb-6 rounded-3xl border border-neutral-800 bg-neutral-900/80 p-6">
      <div class="flex flex-col gap-6 xl:flex-row xl:items-end xl:justify-between">
        <div class="max-w-3xl">
          <p class="mb-2 text-xs uppercase tracking-[0.3em] text-sky-300/80">Online Sources</p>
          <h1 class="text-2xl font-semibold text-white">在线漫画库</h1>
          <p class="mt-2 text-sm leading-6 text-neutral-400">
            这里是 ImageMaster 的在线源入口。下载页仍然保留原来的链接下载能力；在线源页负责搜索、榜单、详情和在线阅读，后面再逐步把更多站点接进来。
          </p>
        </div>

        <div class="w-full max-w-xl">
          <Input
            v-model="searchKeyword"
            help="输入关键词搜索当前选中的源；支持榜单的源会优先显示推荐内容。"
            placeholder="输入关键词搜索在线漫画源"
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
          :disabled="searching || !selectedSourceId || !searchKeyword.trim()"
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
      <div class="space-y-4">
        <div
          v-if="supportsRanking && !isSearchMode"
          class="rounded-3xl border border-neutral-800 bg-neutral-900/80 p-6"
        >
          <div class="mb-4 flex flex-wrap items-center justify-between gap-4">
            <div>
              <h2 class="text-lg font-semibold text-white">推荐内容</h2>
              <p class="mt-1 text-sm text-neutral-400">
                {{ selectedSource?.name }} 支持榜单推荐，可以先从热门内容里挑，再决定要不要搜索。
              </p>
            </div>

            <div class="flex flex-wrap gap-2">
              <button
                v-for="kind in rankingKinds"
                :key="kind"
                class="cursor-pointer rounded-full border px-3 py-1 text-xs transition-colors"
                :class="
                  featuredKind === kind
                    ? 'border-sky-400/60 bg-sky-500/10 text-sky-100'
                    : 'border-neutral-700 text-neutral-300 hover:bg-neutral-800'
                "
                @click="loadRanking(kind)"
              >
                {{ rankingLabel(kind) }}
              </button>
            </div>
          </div>

          <div class="mb-4 flex items-center justify-between gap-3 text-xs text-neutral-500">
            <span>当前显示 {{ featuredVisibleItems.length }} / {{ featuredResult.items.length }}</span>
            <span v-if="featuredLoadingMore" class="text-sky-300">正在加载更多...</span>
          </div>

          <div
            v-if="featuredError"
            class="rounded-2xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-200"
          >
            {{ featuredError }}
          </div>

          <div
            v-else-if="rankingLoading"
            class="rounded-2xl border border-dashed border-neutral-800 px-6 py-12 text-center text-sm text-neutral-500"
          >
            正在加载推荐内容...
          </div>

          <div
            v-else-if="featuredResult.items.length === 0"
            class="rounded-2xl border border-dashed border-neutral-800 px-6 py-12 text-center text-sm text-neutral-500"
          >
            当前榜单还没有可展示的内容。
          </div>

          <div v-else>
            <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              <article
                v-for="item in featuredVisibleItems"
                :key="`featured-${featuredResult.kind}-${item.id}`"
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
                  <div
                    v-else
                    class="flex h-full flex-col items-center justify-center gap-2 px-6 text-center text-xs text-neutral-600"
                  >
                    <span>封面待补</span>
                    <span class="leading-5">点进详情后会继续尝试补图。</span>
                  </div>
                </div>

                <div class="flex flex-col gap-2 p-4">
                  <h3 class="line-clamp-2 text-sm font-semibold text-white">{{ item.title }}</h3>
                  <p class="text-xs text-neutral-400">{{ item.primaryLabel }}</p>
                  <p v-if="item.secondaryLabel" class="text-xs text-neutral-500">{{ item.secondaryLabel }}</p>
                  <p class="line-clamp-3 text-xs leading-5 text-neutral-400">
                    {{ item.summary || '暂无简介' }}
                  </p>

                  <div class="mt-2 flex flex-wrap gap-2">
                    <button
                      class="cursor-pointer rounded-lg border border-sky-500/50 bg-sky-500/10 px-3 py-2 text-xs text-sky-100 transition-colors hover:bg-sky-500/20"
                      @click="openDetail(selectedSourceId, item.id)"
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

            <div
              v-if="featuredResult.items.length > FEATURED_INITIAL_COUNT"
              class="mt-5 rounded-2xl border border-dashed border-neutral-800 px-6 py-5 text-center text-sm text-neutral-500"
            >
              <button
                v-if="canLoadMoreFeatured"
                class="cursor-pointer rounded-xl border border-neutral-700 px-4 py-2 text-sm text-neutral-200 transition-colors hover:bg-neutral-800 disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="featuredLoadingMore"
                @click="loadMoreFeatured"
              >
                {{ featuredLoadingMore ? '正在加载更多...' : '加载更多' }}
              </button>

              <div v-else class="text-neutral-500">已全部加载完成</div>
            </div>
          </div>
        </div>

        <div class="rounded-3xl border border-neutral-800 bg-neutral-900/80 p-6">
          <div class="mb-4 flex items-center justify-between gap-4">
            <div>
              <div class="mb-1 flex flex-wrap items-center gap-2">
                <span
                  v-if="isSearchMode"
                  class="rounded-full border border-sky-500/40 bg-sky-500/10 px-3 py-1 text-[11px] text-sky-100"
                >
                  搜索结果视图
                </span>
                <h2 class="text-lg font-semibold text-white">搜索结果</h2>
              </div>
              <p class="mt-1 text-sm text-neutral-400">{{ resultCaption }}</p>
            </div>

            <div class="flex flex-wrap gap-2">
              <button
                v-if="isSearchMode && supportsRanking"
                class="cursor-pointer rounded-xl border border-neutral-700 px-4 py-2 text-sm text-neutral-200 transition-colors hover:bg-neutral-800"
                @click="clearResult"
              >
                返回推荐区
              </button>

              <button
                v-if="result.hasMore"
                class="cursor-pointer rounded-xl border border-neutral-700 px-4 py-2 text-sm text-neutral-200 transition-colors hover:bg-neutral-800 disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="searching"
                @click="runSearch(result.page + 1)"
              >
                下一页
              </button>
            </div>
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
              <div>当前还没有搜索结果。</div>
              <div class="mt-2">
                {{
                  supportsRanking
                    ? '可以返回推荐区看看榜单，或者换一个关键词再试。'
                    : '先选择一个源，再输入关键词开始搜索。'
                }}
              </div>
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
                <p class="line-clamp-3 text-xs leading-5 text-neutral-400">
                  {{ item.summary || '暂无简介' }}
                </p>

                <div class="mt-2 flex flex-wrap gap-2">
                  <button
                    class="cursor-pointer rounded-lg border border-sky-500/50 bg-sky-500/10 px-3 py-2 text-xs text-sky-100 transition-colors hover:bg-sky-500/20"
                    @click="openDetail(result.source.id, item.id)"
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
      </div>

      <aside class="space-y-4">
        <div class="rounded-3xl border border-neutral-800 bg-neutral-900/80 p-6">
          <h2 class="text-lg font-semibold text-white">当前源</h2>
          <div class="mt-4 rounded-2xl border border-neutral-800 bg-neutral-950/70 p-4">
            <div class="text-base font-medium text-white">{{ selectedSource?.name ?? '未选择' }}</div>
            <p class="mt-2 text-sm text-neutral-400">
              {{ selectedSource?.description ?? '请选择一个在线源。' }}
            </p>

            <div v-if="selectedSource?.capabilities?.length" class="mt-4 flex flex-wrap gap-2">
              <span
                v-for="capability in selectedSource.capabilities"
                :key="capability"
                class="rounded-full border border-neutral-700 px-3 py-1 text-xs text-neutral-300"
              >
                {{ capabilityLabel(capability) }}
              </span>
            </div>

            <a
              v-if="selectedSource?.website"
              :href="selectedSource.website"
              target="_blank"
              rel="noreferrer"
              class="mt-4 inline-block text-xs text-sky-300 hover:text-sky-200"
            >
              访问源站
            </a>
          </div>
        </div>

        <div class="rounded-3xl border border-neutral-800 bg-neutral-900/80 p-6">
          <div class="flex items-center justify-between gap-3">
            <h2 class="text-lg font-semibold text-white">在线书架</h2>
            <span class="rounded-full border border-neutral-700 px-3 py-1 text-xs text-neutral-300">
              {{ onlineShelfItems.length }} 本
            </span>
          </div>

          <p class="mt-2 text-sm leading-6 text-neutral-400">
            不想下载到本地时，可以先加入在线书架，后面直接从软件里回到作品详情或章节。
          </p>

          <div v-if="onlineShelfItems.length === 0" class="mt-5 text-sm text-neutral-500">
            还没有收藏的在线漫画。进入详情页后点“加入在线书架”就行。
          </div>

          <div v-else class="mt-5 space-y-3">
            <button
              v-for="item in onlineShelfItems"
              :key="item.key"
              class="flex w-full cursor-pointer gap-3 rounded-2xl border border-neutral-800 bg-neutral-950/70 p-3 text-left transition-colors hover:bg-neutral-800/70"
              @click="openDetail(item.sourceId, item.itemId)"
            >
              <div class="h-20 w-14 shrink-0 overflow-hidden rounded-xl bg-neutral-900">
                <img
                  v-if="item.cover"
                  :src="item.cover"
                  :alt="item.title"
                  class="h-full w-full object-cover"
                  loading="lazy"
                />
              </div>

              <div class="min-w-0 flex-1">
                <div class="line-clamp-2 text-sm font-medium text-white">{{ item.title }}</div>
                <div class="mt-1 text-xs text-neutral-400">{{ item.sourceName }}</div>
                <div class="mt-2 line-clamp-2 text-xs leading-5 text-neutral-500">
                  {{ item.summary || item.author || '在线书架条目' }}
                </div>
              </div>
            </button>
          </div>
        </div>
      </aside>
    </section>
  </div>
</template>

<script setup lang="ts">
import { Input } from '@/components'
import { useLibraryMetaStore } from '@/stores/libraryMetaStore'
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { GetSourceRanking, ListSources, SearchSources } from '../../../wailsjs/go/source/API'

interface SourceSummary {
  id: string
  name: string
  type: string
  language: string
  website: string
  version?: string
  builtIn?: boolean
  capabilities?: string[]
  rankingKinds?: string[]
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

interface RankingResult {
  source: SourceSummary
  kind: string
  page: number
  total: number
  items: SearchItem[]
}

const FEATURED_INITIAL_COUNT = 9
const FEATURED_LOAD_MORE_COUNT = 6

const router = useRouter()
const libraryMetaStore = useLibraryMetaStore()

const sources = ref<SourceSummary[]>([])
const selectedSourceId = ref('')
const searchKeyword = ref('')
const searching = ref(false)
const errorMessage = ref('')
const rankingLoading = ref(false)
const featuredError = ref('')
const featuredKind = ref('')
const featuredVisibleCount = ref(FEATURED_INITIAL_COUNT)
const featuredLoadingMore = ref(false)

function createEmptySourceSummary(): SourceSummary {
  return {
    id: '',
    name: '',
    type: '',
    language: '',
    website: '',
    description: '',
  }
}

function createEmptySearchResult(): SearchResult {
  return {
    source: createEmptySourceSummary(),
    query: '',
    page: 1,
    hasMore: false,
    total: 0,
    items: [],
  }
}

function createEmptyRankingResult(): RankingResult {
  return {
    source: createEmptySourceSummary(),
    kind: '',
    page: 1,
    total: 0,
    items: [],
  }
}

const result = ref<SearchResult>(createEmptySearchResult())
const featuredResult = ref<RankingResult>(createEmptyRankingResult())

const selectedSource = computed(
  () => sources.value.find((source) => source.id === selectedSourceId.value) ?? null,
)
const onlineShelfItems = computed(() => libraryMetaStore.onlineShelfItems)

const supportsRanking = computed(() =>
  Boolean(
    selectedSource.value?.capabilities?.includes('ranking') &&
      selectedSource.value?.rankingKinds?.length,
  ),
)

const rankingKinds = computed(() => selectedSource.value?.rankingKinds ?? [])

const featuredVisibleItems = computed(() =>
  featuredResult.value.items.slice(0, featuredVisibleCount.value),
)

const canLoadMoreFeatured = computed(
  () => featuredVisibleCount.value < featuredResult.value.items.length,
)

const isSearchMode = computed(
  () => searching.value || Boolean(result.value.query) || Boolean(errorMessage.value),
)

const resultCaption = computed(() => {
  if (!result.value.query) {
    return '当前还没有搜索结果。'
  }

  return `来源：${result.value.source.name} / 关键词：${result.value.query} / 共 ${result.value.total} 条`
})

onMounted(async () => {
  sources.value = await ListSources()
  selectedSourceId.value = sources.value[0]?.id ?? ''
})

watch(
  () => selectedSourceId.value,
  async () => {
    clearResult()
    featuredError.value = ''
    featuredResult.value = createEmptyRankingResult()
    featuredVisibleCount.value = FEATURED_INITIAL_COUNT
    featuredLoadingMore.value = false

    if (supportsRanking.value && rankingKinds.value.length > 0) {
      await loadRanking(rankingKinds.value[0]!)
    }
  },
  { immediate: false },
)

watch(
  () => [featuredResult.value.kind, featuredResult.value.items.length],
  () => {
    featuredVisibleCount.value = FEATURED_INITIAL_COUNT
    featuredLoadingMore.value = false
  },
)

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

async function loadRanking(kind: string) {
  if (!selectedSourceId.value) {
    return
  }

  featuredKind.value = kind
  rankingLoading.value = true
  featuredError.value = ''
  featuredVisibleCount.value = FEATURED_INITIAL_COUNT
  featuredLoadingMore.value = false

  try {
    featuredResult.value = await GetSourceRanking(selectedSourceId.value, kind, 1)
  } catch (error) {
    featuredError.value = error instanceof Error ? error.message : '加载推荐失败，请稍后再试。'
    featuredResult.value = createEmptyRankingResult()
  } finally {
    rankingLoading.value = false
  }
}

function loadMoreFeatured() {
  if (featuredLoadingMore.value || !canLoadMoreFeatured.value) {
    return
  }

  featuredLoadingMore.value = true
  window.setTimeout(() => {
    featuredVisibleCount.value = Math.min(
      featuredVisibleCount.value + FEATURED_LOAD_MORE_COUNT,
      featuredResult.value.items.length,
    )
    featuredLoadingMore.value = false
  }, 280)
}

function openDetail(source: string, id: string) {
  router.push({
    path: '/online/detail',
    query: { source, id },
  })
}

function clearResult() {
  errorMessage.value = ''
  result.value = createEmptySearchResult()
}

function rankingLabel(kind: string) {
  switch (kind) {
    case 'day':
      return '日榜'
    case 'week':
      return '周榜'
    case 'month':
      return '月榜'
    case 'year':
      return '年榜'
    default:
      return kind
  }
}

function capabilityLabel(capability: string) {
  switch (capability) {
    case 'search':
      return '搜索'
    case 'detail':
      return '详情'
    case 'read':
      return '阅读'
    case 'ranking':
      return '榜单'
    default:
      return capability
  }
}
</script>

<style scoped></style>
