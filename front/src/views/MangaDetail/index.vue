<template>
  <div class="flex h-full flex-col">
    <Header :manga-service="mangaService" />

    <Loading v-if="loading" />

    <div v-else-if="selectedImages.length === 0" class="flex h-full flex-grow flex-col items-center justify-center">
      <p class="mb-5 text-gray-100">未找到图片</p>
    </div>

    <main
      v-else
      ref="scrollContainer"
      class="flex flex-1 flex-grow flex-col items-center gap-5 overflow-y-auto p-5"
      @scroll="scrollService.debounceSaveProgress"
    >
      <div class="sticky top-0 z-20 flex w-full max-w-[1200px] justify-end">
        <div class="relative">
          <button
            class="flex h-11 cursor-pointer items-center gap-2 rounded-2xl border border-neutral-800 bg-neutral-950/85 px-4 text-xs text-neutral-200 shadow-lg shadow-black/20 backdrop-blur transition-colors hover:border-sky-500/70 hover:bg-neutral-900 focus:outline-none focus:ring-2 focus:ring-sky-500/40"
            title="阅读显示设置"
            aria-label="阅读显示设置"
            :aria-expanded="showReaderSettings"
            @click="showReaderSettings = !showReaderSettings"
          >
            <SlidersHorizontal :size="16" />
            <span>{{ readerModeLabel }}</span>
          </button>

          <div
            v-if="showReaderSettings"
            class="absolute right-0 top-14 z-30 w-72 rounded-2xl border border-neutral-800 bg-neutral-950/95 p-4 text-xs text-neutral-400 shadow-2xl shadow-black/40 backdrop-blur"
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
                  readerFitMode === option.value
                    ? 'border-sky-500 bg-sky-500/10 text-sky-200'
                    : 'border-neutral-800 bg-neutral-900/80 text-neutral-300 hover:border-neutral-600 hover:text-neutral-100'
                "
                @click="readerFitMode = option.value"
              >
                {{ option.label }}
              </button>
            </div>

            <label v-if="readerFitMode === 'custom'" class="mt-5 block">
              <div class="mb-3 flex items-center justify-between">
                <span>图片宽度</span>
                <span class="text-neutral-100">{{ readerWidthPercent }}%</span>
              </div>
              <input
                v-model.number="readerWidthPercent"
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
      </div>

      <div v-for="(image, index) in selectedImages" :key="index" class="w-full">
        <div class="mx-auto flex w-full justify-center">
          <img
            :src="image"
            :alt="`Manga page ${index + 1}`"
            class="block rounded"
            :class="readerImageClass"
            :style="readerImageStyle"
          />
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { SlidersHorizontal } from 'lucide-vue-next'
import { Loading } from '../../components'
import { Header } from './components'
import { MangaService, ScrollService } from './services'
import { useMangaStore } from './stores'

const scrollContainer = ref<HTMLElement | null>(null)
const mangaStore = useMangaStore()
const { loading, selectedImages } = storeToRefs(mangaStore)
const route = useRoute()

const scrollService = new ScrollService(scrollContainer, mangaStore)
const mangaService = new MangaService(scrollService)
const READER_FIT_MODE_KEY = 'imagemaster:reader-fit-mode'
const READER_WIDTH_PERCENT_KEY = 'imagemaster:reader-width-percent'
type ReaderFitMode = 'width' | 'height' | 'custom'

const showReaderSettings = ref(false)
const readerFitOptions: Array<{ label: string; value: ReaderFitMode }> = [
  { label: '适应宽度', value: 'width' },
  { label: '适应高度', value: 'height' },
  { label: '自定义', value: 'custom' },
]
const storedFitMode = window.localStorage.getItem(READER_FIT_MODE_KEY)
const readerFitMode = ref<ReaderFitMode>(
  storedFitMode === 'height' || storedFitMode === 'custom' || storedFitMode === 'width' ? storedFitMode : 'width',
)
const storedWidthPercent = Number(window.localStorage.getItem(READER_WIDTH_PERCENT_KEY))
const readerWidthPercent = ref(
  Number.isFinite(storedWidthPercent) ? Math.min(Math.max(storedWidthPercent, 40), 120) : 80,
)

const readerImageClass = computed(() => {
  if (readerFitMode.value === 'height') {
    return 'h-[calc(100vh-13rem)] w-auto max-w-full object-contain'
  }

  if (readerFitMode.value === 'custom') {
    return 'h-auto max-w-none'
  }

  return 'h-auto w-full max-w-[1200px]'
})

const readerImageStyle = computed(() => {
  if (readerFitMode.value !== 'custom') {
    return undefined
  }

  return {
    width: `${readerWidthPercent.value}%`,
  }
})

const readerModeLabel = computed(
  () => readerFitOptions.find((option) => option.value === readerFitMode.value)?.label ?? '显示设置',
)

let unregisterKeyboardEvents: (() => void) | undefined

onMounted(() => {
  init()
  unregisterKeyboardEvents = scrollService.registerEvent()
})

onUnmounted(() => {
  unregisterKeyboardEvents?.()
})

watch(
  () => route.params.path,
  (newPath) => {
    if (newPath) {
      init()
    }
  },
)

watch(readerFitMode, (value) => {
  window.localStorage.setItem(READER_FIT_MODE_KEY, value)
})

watch(readerWidthPercent, (value) => {
  window.localStorage.setItem(READER_WIDTH_PERCENT_KEY, String(value))
})

function init() {
  mangaService.loadManga(route.params.path as string)
}
</script>

<style scoped></style>
