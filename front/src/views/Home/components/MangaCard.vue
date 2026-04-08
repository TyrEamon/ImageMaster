<template>
  <article class="group relative xl:w-64 w-36">
    <div class="absolute left-2 top-2 z-10 flex max-w-[70%] flex-wrap gap-1">
      <span
        v-for="badge in statusBadges"
        :key="badge"
        class="rounded-full border border-neutral-700/80 bg-neutral-950/85 px-2 py-1 text-[10px] font-semibold tracking-wide text-neutral-100 backdrop-blur"
      >
        {{ badge }}
      </span>
    </div>

    <div class="absolute right-2 top-2 z-10 flex gap-1">
      <button
        class="rounded-full border border-neutral-700/80 bg-neutral-950/85 px-2 py-1 text-[10px] text-neutral-200 backdrop-blur transition-colors hover:border-amber-400 hover:text-amber-200"
        :class="shelfState.favorite ? 'border-amber-400 text-amber-200' : ''"
        title="收藏"
        @click.stop="toggleFavorite"
      >
        藏
      </button>
      <button
        class="rounded-full border border-neutral-700/80 bg-neutral-950/85 px-2 py-1 text-[10px] text-neutral-200 backdrop-blur transition-colors hover:border-sky-400 hover:text-sky-200"
        :class="shelfState.pinned ? 'border-sky-400 text-sky-200' : ''"
        title="置顶"
        @click.stop="togglePinned"
      >
        顶
      </button>
      <button
        class="rounded-full border border-neutral-700/80 bg-neutral-950/85 px-2 py-1 text-[10px] text-neutral-200 backdrop-blur transition-colors hover:border-emerald-400 hover:text-emerald-200"
        :class="shelfState.readLater ? 'border-emerald-400 text-emerald-200' : ''"
        title="稍后看"
        @click.stop="toggleReadLater"
      >
        后
      </button>
    </div>

    <div
      class="cursor-pointer overflow-hidden rounded-2xl border border-neutral-800 bg-neutral-900 text-left transition-transform duration-300 hover:-translate-y-1 hover:border-neutral-700"
      role="button"
      tabindex="0"
      @click="toMangaDetail"
      @keydown.enter.prevent="toMangaDetail"
      @keydown.space.prevent="toMangaDetail"
    >
      <div class="h-48 overflow-hidden bg-neutral-950">
        <img :src="mangaImageSrc" :alt="manga.name" class="h-full w-full object-cover" />
      </div>

      <div class="flex flex-col gap-2 bg-neutral-900 p-3 xl:min-h-[8.75rem] min-h-[8.5rem]">
        <h3 class="line-clamp-2 text-xs font-bold leading-5 text-white xl:text-sm">
          {{ manga.name }}
        </h3>

        <div class="flex items-center justify-between gap-2 text-[11px] text-neutral-400">
          <span>{{ manga.imagesCount }} 张</span>
          <span class="truncate text-right">{{ progressLabel }}</span>
        </div>

        <div class="mt-auto flex flex-col gap-1.5">
          <div class="h-1.5 overflow-hidden rounded-full bg-neutral-800">
            <div
              class="h-full rounded-full bg-sky-500 transition-all duration-300"
              :style="{ width: `${progressWidth}%` }"
            />
          </div>
          <div class="text-[10px] text-neutral-500">
            {{ secondaryLabel }}
          </div>
        </div>
      </div>
    </div>
  </article>
</template>

<script setup lang="ts">
import { useLibraryMetaStore } from '@/stores/libraryMetaStore'
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { UrlEncode } from '../../../utils'
import { MangaService } from '../services'
import type { Manga } from '../stores/homeStore'

const props = defineProps<{
  manga: Manga
}>()

const mangaService = new MangaService()
const router = useRouter()
const libraryMetaStore = useLibraryMetaStore()

const mangaImageSrc = computed(() => mangaService.getMangaImage(props.manga.previewImg))
const shelfState = computed(() => libraryMetaStore.getShelfState(props.manga.path))
const readingProgress = computed(() => libraryMetaStore.getReadingProgress(props.manga.path))

const statusBadges = computed(() => {
  const badges: string[] = []

  if (shelfState.value.pinned) badges.push('置顶')
  if (shelfState.value.favorite) badges.push('收藏')
  if (shelfState.value.readLater) badges.push('稍后看')
  if (readingProgress.value?.completed) badges.push('已读完')

  return badges
})

const progressWidth = computed(() => Math.round((readingProgress.value?.progressPercent ?? 0) * 100))

const progressLabel = computed(() => {
  const progress = readingProgress.value
  if (!progress) {
    return '未开始'
  }

  if (progress.completed) {
    return '已读完'
  }

  return `继续阅读 ${Math.round(progress.progressPercent * 100)}%`
})

const secondaryLabel = computed(() => {
  const progress = readingProgress.value
  if (!progress) {
    return '点击进入阅读页'
  }

  if (progress.completed) {
    return `上次看到第 ${progress.totalImages}/${progress.totalImages} 张`
  }

  return `上次看到第 ${progress.lastReadImage}/${progress.totalImages} 张`
})

function toggleFavorite() {
  libraryMetaStore.toggleFavorite(props.manga.path)
}

function togglePinned() {
  libraryMetaStore.togglePinned(props.manga.path)
}

function toggleReadLater() {
  libraryMetaStore.toggleReadLater(props.manga.path)
}

function toMangaDetail() {
  router.push(`/manga/${UrlEncode(props.manga.path)}`)
}
</script>

<style scoped></style>
