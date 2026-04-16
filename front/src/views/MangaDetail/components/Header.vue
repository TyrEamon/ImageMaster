<template>
  <header class="flex flex-col gap-3 border-b border-neutral-800/70 bg-transparent px-4 py-3">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div class="flex min-w-0 flex-1 items-center gap-2">
        <Button @click="router.push('/')">
          <div class="flex items-center gap-2">
            <ArrowLeft :size="16" class="text-white" />
            <span>返回</span>
          </div>
        </Button>

        <div class="min-w-0 flex-1 truncate text-sm font-bold text-white">
          {{ mangaName }}
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <Button @click="showNavigation = !showNavigation">
          <div class="flex items-center gap-2">
            <EyeClosed v-if="!showNavigation" :size="16" class="text-white" />
            <Eye v-else :size="16" class="text-white" />
            <span>{{ showNavigation ? '隐藏导航' : '显示导航' }}</span>
          </div>
        </Button>

        <Button @click="showQuickDownloadModal = true">
          <div class="flex items-center gap-2">
            <Download :size="16" class="text-white" />
            <span>快速下载</span>
          </div>
        </Button>

        <Button @click="() => mangaService.deleteAndViewNextManga()">
          <div class="flex items-center gap-2">
            <Trash :size="16" class="text-white" />
            <span>删除并看下一部</span>
          </div>
        </Button>
      </div>
    </div>

    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="flex flex-wrap items-center gap-2 text-xs text-neutral-400">
        <span class="rounded-full border border-neutral-800 px-3 py-1.5">
          {{ progressSummary }}
        </span>
        <span v-if="shelfState.pinned" class="rounded-full border border-sky-500/40 px-3 py-1.5 text-sky-200">
          置顶
        </span>
        <span
          v-if="shelfState.favorite"
          class="rounded-full border border-amber-500/40 px-3 py-1.5 text-amber-200"
        >
          收藏
        </span>
        <span
          v-if="shelfState.readLater"
          class="rounded-full border border-emerald-500/40 px-3 py-1.5 text-emerald-200"
        >
          稍后看
        </span>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <div ref="readerSettingsRef" class="relative">
          <button
            class="rounded-xl border px-3 py-1.5 text-xs transition-colors"
            :class="
              showReaderSettings
                ? 'border-sky-400 bg-sky-500/10 text-sky-200'
                : 'border-neutral-700 text-neutral-300 hover:bg-neutral-800'
            "
            title="阅读显示设置"
            aria-label="阅读显示设置"
            :aria-expanded="showReaderSettings"
            @click="showReaderSettings = !showReaderSettings"
          >
            <span class="flex items-center gap-2">
              <SlidersHorizontal :size="14" />
              <span>{{ readerModeLabel }}</span>
            </span>
          </button>

          <div
            v-if="showReaderSettings"
            class="absolute right-0 top-10 z-30 w-72 rounded-2xl border border-neutral-800 bg-neutral-950/95 p-4 text-xs text-neutral-400 shadow-2xl shadow-black/40 backdrop-blur"
            @keydown.escape="showReaderSettings = false"
          >
            <div class="mb-4 flex items-center justify-between">
              <span class="text-neutral-300">阅读显示</span>
              <span class="text-neutral-100">{{ readerModeLabel }}</span>
            </div>

            <div class="grid grid-cols-3 gap-2">
              <button
                v-for="option in readerFitOptions"
                :key="option.value"
                class="cursor-pointer rounded-xl border px-3 py-2 transition-colors"
                :class="
                  readerFitModeModel === option.value
                    ? 'border-sky-500 bg-sky-500/10 text-sky-200'
                    : 'border-neutral-800 bg-neutral-900/80 text-neutral-300 hover:border-neutral-600 hover:text-neutral-100'
                "
                @click="readerFitModeModel = option.value"
              >
                {{ option.label }}
              </button>
            </div>

            <label v-if="readerFitModeModel === 'custom'" class="mt-5 block">
              <div class="mb-3 flex items-center justify-between">
                <span>图片宽度</span>
                <span class="text-neutral-100">{{ readerWidthPercentModel }}%</span>
              </div>
              <input
                v-model.number="readerWidthPercentModel"
                class="w-full accent-sky-500"
                type="range"
                min="40"
                max="120"
                step="5"
                aria-label="阅读图片宽度比例"
              />
            </label>
          </div>
        </div>

        <button
          class="rounded-xl border px-3 py-1.5 text-xs transition-colors"
          :class="
            shelfState.favorite
              ? 'border-amber-400 bg-amber-500/10 text-amber-200'
              : 'border-neutral-700 text-neutral-300 hover:bg-neutral-800'
          "
          @click="toggleFavorite"
        >
          {{ shelfState.favorite ? '取消收藏' : '加入收藏' }}
        </button>

        <button
          class="rounded-xl border px-3 py-1.5 text-xs transition-colors"
          :class="
            shelfState.pinned
              ? 'border-sky-400 bg-sky-500/10 text-sky-200'
              : 'border-neutral-700 text-neutral-300 hover:bg-neutral-800'
          "
          @click="togglePinned"
        >
          {{ shelfState.pinned ? '取消置顶' : '设为置顶' }}
        </button>

        <button
          class="rounded-xl border px-3 py-1.5 text-xs transition-colors"
          :class="
            shelfState.readLater
              ? 'border-emerald-400 bg-emerald-500/10 text-emerald-200'
              : 'border-neutral-700 text-neutral-300 hover:bg-neutral-800'
          "
          @click="toggleReadLater"
        >
          {{ shelfState.readLater ? '取消稍后看' : '稍后看' }}
        </button>
      </div>
    </div>
  </header>

  <QuickDownloadModal v-model="showQuickDownloadModal" />

  <Transition>
    <div v-if="showNavigation" class="fixed bottom-8 right-8">
      <div class="flex flex-col gap-2">
        <button
          class="cursor-pointer rounded-full bg-neutral-900/70 p-3 hover:bg-neutral-900/90"
          @click="mangaService.navigateToNextManga()"
        >
          <ChevronRight :size="20" class="text-white" />
        </button>
        <button
          class="cursor-pointer rounded-full bg-neutral-900/70 p-3 hover:bg-neutral-900/90"
          @click="mangaService.navigateToPrevManga()"
        >
          <ChevronLeft :size="20" class="text-white" />
        </button>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { Button, QuickDownloadModal } from '@/components'
import { useLibraryMetaStore } from '@/stores/libraryMetaStore'
import { storeToRefs } from 'pinia'
import { ArrowLeft, ChevronLeft, ChevronRight, Download, Eye, EyeClosed, SlidersHorizontal, Trash } from 'lucide-vue-next'
import { computed, onMounted, onUnmounted, ref, type PropType } from 'vue'
import { useRouter } from 'vue-router'
import { MangaService } from '../services'
import { useMangaStore } from '../stores'

const router = useRouter()
const mangaStore = useMangaStore()
const libraryMetaStore = useLibraryMetaStore()
const { mangaName, mangaPath, selectedImages } = storeToRefs(mangaStore)
type ReaderFitMode = 'width' | 'height' | 'custom'
type ReaderFitOption = { label: string; value: ReaderFitMode }

const props = defineProps({
  mangaService: {
    type: Object as PropType<MangaService>,
    required: true,
  },
  readerFitMode: {
    type: String as PropType<ReaderFitMode>,
    required: true,
  },
  readerWidthPercent: {
    type: Number,
    required: true,
  },
  readerFitOptions: {
    type: Array as PropType<ReaderFitOption[]>,
    required: true,
  },
  readerModeLabel: {
    type: String,
    required: true,
  },
})

const emit = defineEmits<{
  (event: 'update:readerFitMode', value: ReaderFitMode): void
  (event: 'update:readerWidthPercent', value: number): void
}>()

const mangaService = props.mangaService
const showNavigation = ref(false)
const showQuickDownloadModal = ref(false)
const showReaderSettings = ref(false)
const readerSettingsRef = ref<HTMLElement | null>(null)
const readerFitModeModel = computed({
  get: () => props.readerFitMode,
  set: (value: ReaderFitMode) => emit('update:readerFitMode', value),
})
const readerWidthPercentModel = computed({
  get: () => props.readerWidthPercent,
  set: (value: number) => emit('update:readerWidthPercent', value),
})

const shelfState = computed(() => libraryMetaStore.getShelfState(mangaPath.value))
const readingProgress = computed(() => libraryMetaStore.getReadingProgress(mangaPath.value))

const progressSummary = computed(() => {
  const progress = readingProgress.value
  const totalImages = selectedImages.value.length || progress?.totalImages || 0

  if (!progress) {
    return totalImages > 0 ? `尚未开始阅读 · 共 ${totalImages} 张` : '尚未开始阅读'
  }

  if (progress.completed) {
    return `已读完 · ${totalImages || progress.totalImages}/${totalImages || progress.totalImages} 张`
  }

  return `已读 ${Math.round(progress.progressPercent * 100)}% · 第 ${progress.lastReadImage}/${progress.totalImages} 张`
})

function toggleFavorite() {
  if (mangaPath.value) {
    libraryMetaStore.toggleFavorite(mangaPath.value)
  }
}

function togglePinned() {
  if (mangaPath.value) {
    libraryMetaStore.togglePinned(mangaPath.value)
  }
}

function toggleReadLater() {
  if (mangaPath.value) {
    libraryMetaStore.toggleReadLater(mangaPath.value)
  }
}

function handleKeydown(event: KeyboardEvent) {
  if (event.ctrlKey && event.key === 'd') {
    showQuickDownloadModal.value = true
  }
}

function handleDocumentPointerDown(event: PointerEvent) {
  if (!showReaderSettings.value) {
    return
  }

  const target = event.target
  if (target instanceof Node && readerSettingsRef.value?.contains(target)) {
    return
  }

  showReaderSettings.value = false
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
  document.addEventListener('pointerdown', handleDocumentPointerDown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
  document.removeEventListener('pointerdown', handleDocumentPointerDown)
})
</script>

<style scoped>
.v-enter-active,
.v-leave-active {
  transition: all 0.5s ease;
}

.v-enter-from,
.v-leave-to {
  opacity: 0;
  transform: translateY(20px);
}
</style>
