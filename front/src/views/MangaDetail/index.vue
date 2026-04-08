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
      <div v-for="(image, index) in selectedImages" :key="index" class="w-full">
        <div class="mx-auto w-full max-w-[1200px]">
          <img :src="image" :alt="`Manga page ${index + 1}`" class="block h-auto w-full rounded" />
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { onMounted, onUnmounted, ref, watch } from 'vue'
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

function init() {
  mangaService.loadManga(route.params.path as string)
}
</script>

<style scoped></style>
