<template>
  <div class="min-h-screen overflow-auto">
    <div
      class="sticky top-0 z-10 border-b border-neutral-800/70 bg-neutral-950/85 px-8 py-5 backdrop-blur"
    >
      <div class="flex flex-col gap-4">
        <div class="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
          <div class="w-full max-w-2xl">
            <Input
              v-model="searchQuery"
              help="本地搜索：支持简繁统一、常见日文字形归一，并会尽量忽略括号、社团前缀和 C101 这类编号。"
              placeholder="搜索漫画名、作者名、编号或文件夹关键字"
              autofocus
            />
          </div>

          <div class="flex flex-wrap items-end gap-3">
            <label class="flex min-w-[10rem] flex-col gap-1 text-xs text-neutral-400">
              <span>排序</span>
              <select
                v-model="sortMode"
                class="rounded-xl border border-neutral-800 bg-neutral-900 px-3 py-2 text-sm text-neutral-100 outline-none transition-colors focus:border-sky-500"
              >
                <option value="smart">默认排序</option>
                <option value="recent-read">最近阅读</option>
                <option value="images-desc">图片数最多</option>
                <option value="name-asc">名称 A-Z</option>
                <option value="name-desc">名称 Z-A</option>
              </select>
            </label>

            <label class="flex min-w-[10rem] flex-col gap-1 text-xs text-neutral-400">
              <span>筛选</span>
              <select
                v-model="statusFilter"
                class="rounded-xl border border-neutral-800 bg-neutral-900 px-3 py-2 text-sm text-neutral-100 outline-none transition-colors focus:border-sky-500"
              >
                <option value="all">全部</option>
                <option value="pinned">只看置顶</option>
                <option value="favorite">只看收藏</option>
                <option value="read-later">只看稍后看</option>
                <option value="in-progress">继续阅读</option>
                <option value="completed">已读完</option>
                <option value="unread">未开始</option>
              </select>
            </label>

            <button
              v-if="showResetButton"
              class="cursor-pointer rounded-xl border border-neutral-700 px-3 py-2 text-sm text-neutral-200 transition-colors duration-200 hover:bg-neutral-800"
              @click="resetControls"
            >
              重置
            </button>
          </div>
        </div>

        <div class="flex flex-wrap gap-2 text-xs text-neutral-400">
          <span class="rounded-full border border-neutral-800 px-3 py-1.5">
            显示 {{ visibleMangas.length }} / {{ mangas.length }} 本
          </span>

          <button
            class="cursor-pointer rounded-full border px-3 py-1.5 transition-colors"
            :class="getQuickFilterClass('pinned')"
            @click="toggleQuickFilter('pinned')"
          >
            置顶 {{ pinnedCount }}
          </button>

          <button
            class="cursor-pointer rounded-full border px-3 py-1.5 transition-colors"
            :class="getQuickFilterClass('favorite')"
            @click="toggleQuickFilter('favorite')"
          >
            收藏 {{ favoriteCount }}
          </button>

          <button
            class="cursor-pointer rounded-full border px-3 py-1.5 transition-colors"
            :class="getQuickFilterClass('read-later')"
            @click="toggleQuickFilter('read-later')"
          >
            稍后看 {{ readLaterCount }}
          </button>

          <button
            class="cursor-pointer rounded-full border px-3 py-1.5 transition-colors"
            :class="getQuickFilterClass('in-progress')"
            @click="toggleQuickFilter('in-progress')"
          >
            继续阅读 {{ inProgressCount }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="visibleMangas.length > 0" class="flex flex-wrap gap-4 p-8 pt-6">
      <MangaCard v-for="manga in visibleMangas" :key="manga.path" :manga="manga" />
    </div>

    <div v-else class="px-8 py-16 text-center text-sm text-neutral-400">
      {{ emptyMessage }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { Input } from '@/components'
import { useLibraryMetaStore, type MangaReadingProgress } from '@/stores/libraryMetaStore'
import { buildMangaSearchIndex, splitSearchKeywords } from '@/utils/search'
import { storeToRefs } from 'pinia'
import { computed, ref } from 'vue'
import { MangaCard } from '.'
import { useHomeStore } from '../stores/homeStore'

type SortMode = 'smart' | 'recent-read' | 'images-desc' | 'name-asc' | 'name-desc'
type StatusFilter =
  | 'all'
  | 'pinned'
  | 'favorite'
  | 'read-later'
  | 'in-progress'
  | 'completed'
  | 'unread'

const homeStore = useHomeStore()
const libraryMetaStore = useLibraryMetaStore()
const { mangas } = storeToRefs(homeStore)

libraryMetaStore.cleanupOldReadingProgress()

const searchQuery = ref('')
const sortMode = ref<SortMode>('smart')
const statusFilter = ref<StatusFilter>('all')

const keywords = computed(() => splitSearchKeywords(searchQuery.value))

function isInProgress(progress: MangaReadingProgress | null | undefined) {
  return Boolean(progress && progress.progressPercent > 0 && !progress.completed)
}

function compareName(a: string, b: string) {
  return a.localeCompare(b, 'zh-Hans-CN')
}

const enrichedMangas = computed(() => {
  return mangas.value.map((manga) => ({
    manga,
    searchIndex: buildMangaSearchIndex(manga.name, manga.path),
    shelf: libraryMetaStore.getShelfState(manga.path),
    progress: libraryMetaStore.getReadingProgress(manga.path),
  }))
})

const favoriteCount = computed(() => enrichedMangas.value.filter((item) => item.shelf.favorite).length)
const pinnedCount = computed(() => enrichedMangas.value.filter((item) => item.shelf.pinned).length)
const readLaterCount = computed(() => enrichedMangas.value.filter((item) => item.shelf.readLater).length)
const inProgressCount = computed(() =>
  enrichedMangas.value.filter((item) => isInProgress(item.progress)).length,
)

const visibleMangas = computed(() => {
  const filtered = enrichedMangas.value.filter((item) => {
    const matchesKeywords =
      keywords.value.length === 0 || keywords.value.every((keyword) => item.searchIndex.includes(keyword))

    if (!matchesKeywords) {
      return false
    }

    switch (statusFilter.value) {
      case 'pinned':
        return item.shelf.pinned
      case 'favorite':
        return item.shelf.favorite
      case 'read-later':
        return item.shelf.readLater
      case 'in-progress':
        return isInProgress(item.progress)
      case 'completed':
        return Boolean(item.progress?.completed)
      case 'unread':
        return !item.progress || item.progress.progressPercent <= 0
      default:
        return true
    }
  })

  const sorted = [...filtered].sort((left, right) => {
    switch (sortMode.value) {
      case 'recent-read':
        return (
          (right.progress?.timestamp ?? 0) - (left.progress?.timestamp ?? 0) ||
          compareName(left.manga.name, right.manga.name)
        )
      case 'images-desc':
        return right.manga.imagesCount - left.manga.imagesCount || compareName(left.manga.name, right.manga.name)
      case 'name-asc':
        return compareName(left.manga.name, right.manga.name)
      case 'name-desc':
        return compareName(right.manga.name, left.manga.name)
      case 'smart':
      default: {
        const pinnedDiff = Number(right.shelf.pinned) - Number(left.shelf.pinned)
        if (pinnedDiff !== 0) return pinnedDiff

        const inProgressDiff = Number(isInProgress(right.progress)) - Number(isInProgress(left.progress))
        if (inProgressDiff !== 0) return inProgressDiff

        const recentReadDiff = (right.progress?.timestamp ?? 0) - (left.progress?.timestamp ?? 0)
        if (recentReadDiff !== 0) return recentReadDiff

        return compareName(left.manga.name, right.manga.name)
      }
    }
  })

  return sorted.map(({ manga }) => manga)
})

const showResetButton = computed(() => {
  return Boolean(searchQuery.value) || sortMode.value !== 'smart' || statusFilter.value !== 'all'
})

const emptyMessage = computed(() => {
  if (mangas.value.length === 0) {
    return '当前漫画库里还没有内容。'
  }

  if (statusFilter.value !== 'all' && !searchQuery.value) {
    return '当前筛选条件下没有找到符合条件的漫画。'
  }

  return '没有找到匹配的漫画，试试标题、作者名、编号或更短的关键字。'
})

function resetControls() {
  searchQuery.value = ''
  sortMode.value = 'smart'
  statusFilter.value = 'all'
}

function toggleQuickFilter(filter: Exclude<StatusFilter, 'all' | 'completed' | 'unread'>) {
  statusFilter.value = statusFilter.value === filter ? 'all' : filter
}

function getQuickFilterClass(filter: Exclude<StatusFilter, 'all' | 'completed' | 'unread'>) {
  return statusFilter.value === filter
    ? 'border-sky-400 bg-sky-500/10 text-sky-200'
    : 'border-neutral-800 text-neutral-400 hover:bg-neutral-800 hover:text-neutral-200'
}
</script>

<style scoped></style>
