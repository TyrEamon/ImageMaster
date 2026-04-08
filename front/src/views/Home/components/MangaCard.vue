<template>
  <article class="xl:w-64 w-36 transition-transform duration-300 hover:-translate-y-1">
    <div
      class="overflow-hidden rounded-2xl border border-neutral-800 bg-neutral-900 text-left transition-colors duration-300 hover:border-neutral-700"
    >
      <div
        class="cursor-pointer"
        role="button"
        tabindex="0"
        @click="toMangaDetail"
        @keydown.enter.prevent="toMangaDetail"
        @keydown.space.prevent="toMangaDetail"
      >
        <div class="h-48 overflow-hidden bg-neutral-950">
          <img :src="mangaImageSrc" :alt="manga.name" class="h-full w-full object-cover" />
        </div>

        <div class="flex flex-col gap-2 bg-neutral-900 p-3 xl:min-h-[10.25rem] min-h-[10rem]">
          <h3 class="line-clamp-2 text-xs font-bold leading-5 text-white xl:text-sm">
            {{ manga.name }}
          </h3>

          <div v-if="statusBadges.length > 0" class="flex flex-wrap gap-1">
            <span
              v-for="badge in statusBadges"
              :key="badge"
              class="rounded-full border border-neutral-700 px-2 py-1 text-[10px] text-neutral-300"
            >
              {{ badge }}
            </span>
          </div>

          <div class="flex items-center justify-between gap-2 text-[11px] text-neutral-400">
            <span>{{ manga.imagesCount }} 张</span>
            <span class="truncate text-right">{{ progressLabel }}</span>
          </div>
        </div>
      </div>

      <div class="border-t border-neutral-800 px-3 py-2">
        <div class="mb-2 flex items-center justify-end gap-2">
          <button
            class="rounded-lg border p-2 text-neutral-300 transition-colors hover:bg-neutral-800"
            :class="shelfState.favorite ? 'border-amber-400/60 bg-amber-500/10 text-amber-200' : 'border-neutral-700'"
            title="收藏"
            @click.stop="toggleFavorite"
          >
            <Heart :size="14" class="stroke-current" :fill="shelfState.favorite ? 'currentColor' : 'none'" />
          </button>

          <button
            class="rounded-lg border p-2 text-neutral-300 transition-colors hover:bg-neutral-800"
            :class="shelfState.pinned ? 'border-sky-400/60 bg-sky-500/10 text-sky-200' : 'border-neutral-700'"
            title="置顶"
            @click.stop="togglePinned"
          >
            <Pin :size="14" class="stroke-current" />
          </button>

          <button
            class="rounded-lg border p-2 text-neutral-300 transition-colors hover:bg-neutral-800"
            :class="shelfState.readLater ? 'border-emerald-400/60 bg-emerald-500/10 text-emerald-200' : 'border-neutral-700'"
            title="稍后看"
            @click.stop="toggleReadLater"
          >
            <Clock3 :size="14" class="stroke-current" />
          </button>
        </div>

        <div class="flex flex-col gap-1.5">
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
import { Clock3, Heart, Pin } from 'lucide-vue-next'
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
