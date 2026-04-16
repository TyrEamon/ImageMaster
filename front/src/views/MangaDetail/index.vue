<template>
  <div class="flex h-full flex-col">
    <Header
      :manga-service="mangaService"
      v-model:reader-fit-mode="readerFitMode"
      v-model:reader-width-percent="readerWidthPercent"
      :reader-fit-options="readerFitOptions"
      :reader-mode-label="readerModeLabel"
    />

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
