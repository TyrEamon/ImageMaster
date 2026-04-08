<template>
  <div class="h-screen overflow-auto">
    <div
      class="sticky top-0 z-10 border-b border-neutral-800/70 bg-neutral-950/85 px-8 py-5 backdrop-blur"
    >
      <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div class="w-full max-w-2xl">
          <Input
            v-model="searchQuery"
            help="本地搜索，仅过滤当前漫画库；支持简繁统一和常见日文字形归一"
            placeholder="搜索漫画名、作者名、编号或文件夹关键字"
            autofocus
          />
        </div>

        <div class="flex items-center gap-3 text-xs text-neutral-400">
          <div>显示 {{ filteredMangas.length }} / {{ mangas.length }} 本</div>
          <button
            v-if="searchQuery"
            class="cursor-pointer rounded-xl border border-neutral-700 px-3 py-1.5 text-neutral-200 transition-colors duration-200 hover:bg-neutral-800"
            @click="clearSearch"
          >
            清空
          </button>
        </div>
      </div>
    </div>

    <div v-if="filteredMangas.length > 0" class="flex flex-wrap gap-4 p-8 pt-6">
      <MangaCard v-for="manga in filteredMangas" :key="manga.path" :manga="manga" />
    </div>

    <div v-else class="px-8 py-16 text-center text-sm text-neutral-400">
      没找到匹配的漫画，试试标题、作者名、编号或文件夹关键字。
    </div>
  </div>
</template>

<script setup lang="ts">
import { Input } from '@/components'
import { buildMangaSearchIndex, splitSearchKeywords } from '@/utils/search'
import { storeToRefs } from 'pinia'
import { computed, ref } from 'vue'
import { MangaCard } from '.'
import { useHomeStore } from '../stores/homeStore'

const homeStore = useHomeStore()
const { mangas } = storeToRefs(homeStore)

const searchQuery = ref('')

const keywords = computed(() => {
  return splitSearchKeywords(searchQuery.value)
})

const indexedMangas = computed(() => {
  return mangas.value.map((manga) => ({
    manga,
    searchIndex: buildMangaSearchIndex(manga.name, manga.path),
  }))
})

const filteredMangas = computed(() => {
  if (keywords.value.length === 0) {
    return mangas.value
  }

  return indexedMangas.value
    .filter(({ searchIndex }) => {
      return keywords.value.every((keyword) => searchIndex.includes(keyword))
    })
    .map(({ manga }) => manga)
})

function clearSearch() {
  searchQuery.value = ''
}
</script>

<style scoped></style>
