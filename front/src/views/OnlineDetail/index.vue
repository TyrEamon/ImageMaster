<template>
  <div class="min-h-screen overflow-y-auto px-8 py-6">
    <section class="mb-6 rounded-3xl border border-neutral-800 bg-neutral-900/80 p-6">
      <div class="mb-6 flex flex-wrap items-center gap-3">
        <button
          class="cursor-pointer rounded-xl border border-neutral-700 px-4 py-2 text-sm text-neutral-200 transition-colors hover:bg-neutral-800"
          @click="router.back()"
        >
          返回
        </button>
        <div class="text-sm text-neutral-400">
          在线详情 / {{ detail.source.name || sourceId || '未选择源' }}
        </div>
      </div>

      <div v-if="loading" class="py-16 text-center text-sm text-neutral-400">正在加载详情...</div>

      <div
        v-else-if="errorMessage"
        class="rounded-2xl border border-red-500/30 bg-red-500/10 px-4 py-4 text-sm text-red-200"
      >
        {{ errorMessage }}
      </div>

      <div v-else class="grid gap-6 xl:grid-cols-[320px_1fr]">
        <div class="overflow-hidden rounded-3xl border border-neutral-800 bg-neutral-950/70">
          <div class="aspect-[3/4] overflow-hidden bg-neutral-900">
            <img
              v-if="detail.item.cover"
              :src="detail.item.cover"
              :alt="detail.item.title"
              class="h-full w-full object-cover"
            />
            <div v-else class="flex h-full items-center justify-center text-sm text-neutral-600">
              No Cover
            </div>
          </div>
        </div>

        <div class="space-y-5">
          <div>
            <p class="mb-2 text-xs uppercase tracking-[0.3em] text-sky-300/80">Comic Detail</p>
            <h1 class="text-3xl font-semibold text-white">{{ detail.item.title }}</h1>
            <div class="mt-3 flex flex-wrap gap-2">
              <span class="rounded-full border border-neutral-700 px-3 py-1 text-xs text-neutral-200">
                作者：{{ detail.item.author || '未知' }}
              </span>
              <span class="rounded-full border border-neutral-700 px-3 py-1 text-xs text-neutral-200">
                状态：{{ detail.item.status || '未知' }}
              </span>
              <span
                v-for="tag in detail.item.tags"
                :key="tag"
                class="rounded-full border border-sky-500/30 bg-sky-500/10 px-3 py-1 text-xs text-sky-100"
              >
                {{ tag }}
              </span>
            </div>
          </div>

          <p class="rounded-2xl border border-neutral-800 bg-neutral-950/70 p-4 text-sm leading-7 text-neutral-300">
            {{ detail.item.summary || '暂无简介' }}
          </p>

          <div class="flex flex-wrap gap-3">
            <button
              class="cursor-pointer rounded-xl border px-4 py-2 text-sm transition-colors"
              :class="
                inOnlineShelf
                  ? 'border-amber-400/50 bg-amber-500/10 text-amber-100 hover:bg-amber-500/20'
                  : 'border-neutral-700 text-neutral-200 hover:bg-neutral-800'
              "
              @click="toggleOnlineShelf"
            >
              {{ inOnlineShelf ? '移出在线书架' : '加入在线书架' }}
            </button>

            <a
              v-if="detail.item.detailUrl"
              :href="detail.item.detailUrl"
              target="_blank"
              rel="noreferrer"
              class="rounded-xl border border-neutral-700 px-4 py-2 text-sm text-neutral-300 transition-colors hover:bg-neutral-800"
            >
              打开源站作品页
            </a>
          </div>
        </div>
      </div>
    </section>

    <section class="rounded-3xl border border-neutral-800 bg-neutral-900/80 p-6">
      <div class="mb-4 flex items-center justify-between gap-4">
        <div>
          <h2 class="text-lg font-semibold text-white">章节列表</h2>
          <p class="mt-1 text-sm text-neutral-400">
            点章节会进入软件内的在线阅读页，不需要先下载到本地。
          </p>
        </div>
        <div class="rounded-full border border-neutral-700 px-3 py-1 text-xs text-neutral-300">
          共 {{ detail.item.chapters.length }} 章
        </div>
      </div>

      <div
        v-if="detail.item.chapters.length === 0"
        class="rounded-2xl border border-dashed border-neutral-800 px-6 py-12 text-center text-sm text-neutral-500"
      >
        当前这个源还没有返回章节列表。
      </div>

      <div v-else class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
        <button
          v-for="chapter in detail.item.chapters"
          :key="chapter.id"
          class="cursor-pointer rounded-2xl border border-neutral-800 bg-neutral-950/60 px-4 py-4 text-left transition-colors hover:bg-neutral-800/70"
          @click="openReader(chapter)"
        >
          <div class="text-sm font-medium text-white">{{ chapter.name }}</div>
          <div v-if="chapter.updatedLabel" class="mt-2 text-xs text-neutral-500">
            {{ chapter.updatedLabel }}
          </div>
          <div class="mt-3 text-xs text-sky-300">在软件内阅读</div>
        </button>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { useLibraryMetaStore } from '@/stores/libraryMetaStore'
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { GetSourceDetail } from '../../../wailsjs/go/source/API'

interface SourceSummary {
  id: string
  name: string
  type: string
  language: string
  website: string
  version?: string
  builtIn?: boolean
  capabilities?: string[]
  description: string
}

interface ChapterItem {
  id: string
  name: string
  url: string
  index?: number
  updatedLabel: string
}

interface DetailItem {
  id: string
  title: string
  cover: string
  summary: string
  author: string
  status: string
  tags: string[]
  detailUrl: string
  chapters: ChapterItem[]
}

interface DetailResult {
  source: SourceSummary
  item: DetailItem
}

function createEmptyDetailResult(): DetailResult {
  return {
    source: {
      id: '',
      name: '',
      type: '',
      language: '',
      website: '',
      description: '',
    },
    item: {
      id: '',
      title: '',
      cover: '',
      summary: '',
      author: '',
      status: '',
      tags: [],
      detailUrl: '',
      chapters: [],
    },
  }
}

const route = useRoute()
const router = useRouter()
const libraryMetaStore = useLibraryMetaStore()
const loading = ref(false)
const errorMessage = ref('')
const detail = ref<DetailResult>(createEmptyDetailResult())

const sourceId = computed(() => String(route.query.source ?? '').trim())
const itemId = computed(() => String(route.query.id ?? '').trim())
const inOnlineShelf = computed(() => {
  if (!sourceId.value || !itemId.value) {
    return false
  }

  return libraryMetaStore.isOnlineShelfSaved(sourceId.value, itemId.value)
})

onMounted(() => {
  loadDetail()
})

watch(
  () => [sourceId.value, itemId.value],
  () => {
    loadDetail()
  },
)

async function loadDetail() {
  if (!sourceId.value || !itemId.value) {
    errorMessage.value = '缺少在线源参数，无法加载详情。'
    return
  }

  loading.value = true
  errorMessage.value = ''

  try {
    detail.value = await GetSourceDetail(sourceId.value, itemId.value)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '加载详情失败，请稍后再试。'
  } finally {
    loading.value = false
  }
}

function toggleOnlineShelf() {
  const saved = libraryMetaStore.toggleOnlineShelfItem({
    sourceId: detail.value.source.id,
    sourceName: detail.value.source.name,
    itemId: detail.value.item.id,
    title: detail.value.item.title,
    cover: detail.value.item.cover,
    summary: detail.value.item.summary,
    author: detail.value.item.author,
    status: detail.value.item.status,
    detailUrl: detail.value.item.detailUrl,
  })

  toast.success(saved ? '已加入在线书架' : '已移出在线书架', {
    description: detail.value.item.title,
  })
}

function openReader(chapter: ChapterItem) {
  router.push({
    path: '/online/read',
    query: {
      source: detail.value.source.id,
      chapter: chapter.id,
    },
  })
}
</script>

<style scoped></style>
