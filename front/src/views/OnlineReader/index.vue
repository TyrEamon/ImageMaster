<template>
  <div class="flex h-full flex-col">
    <header class="border-b border-neutral-800 bg-neutral-950/80 px-6 py-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="flex flex-wrap items-center gap-3">
          <button
            class="cursor-pointer rounded-xl border border-neutral-700 px-4 py-2 text-sm text-neutral-200 transition-colors hover:bg-neutral-800"
            @click="router.back()"
          >
            返回
          </button>
          <div>
            <div class="text-sm font-medium text-white">{{ imageResult.chapterTitle || '在线章节' }}</div>
            <div class="mt-1 text-xs text-neutral-400">
              {{ imageResult.comicTitle || sourceId || '在线漫画' }}
            </div>
          </div>
        </div>

        <div class="flex flex-wrap gap-2">
          <button
            class="cursor-pointer rounded-xl border border-emerald-500/50 bg-emerald-500/10 px-4 py-2 text-sm text-emerald-100 transition-colors hover:bg-emerald-500/20 disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="loading || downloading || resolvedImages.length === 0"
            @click="downloadCurrentChapter"
          >
            {{ downloading ? '下载中...' : '下载到本地漫画库' }}
          </button>
        </div>
      </div>
    </header>

    <div v-if="loading" class="flex h-full flex-1 items-center justify-center text-sm text-neutral-400">
      正在加载图片...
    </div>

    <div
      v-else-if="errorMessage"
      class="m-6 rounded-2xl border border-red-500/30 bg-red-500/10 px-4 py-4 text-sm text-red-200"
    >
      {{ errorMessage }}
    </div>

    <main
      v-else
      class="flex flex-1 flex-col items-center gap-5 overflow-y-auto bg-neutral-950 px-4 py-5"
    >
      <div
        v-for="(image, index) in resolvedImages"
        :key="`${image}-${index}`"
        class="w-full max-w-[1200px]"
      >
        <img
          :src="image"
          :alt="`${imageResult.chapterTitle || 'Chapter'} - ${index + 1}`"
          class="block h-auto w-full rounded-xl border border-neutral-800 bg-neutral-900"
          loading="lazy"
        />
      </div>

      <div v-if="resolvedImages.length === 0" class="py-8 text-sm text-neutral-500">
        当前章节还没有返回图片。
      </div>

      <div class="mt-4 flex flex-wrap justify-center gap-3 pb-6">
        <a
          v-if="imageResult.chapterUrl"
          :href="imageResult.chapterUrl"
          target="_blank"
          rel="noreferrer"
          class="rounded-xl border border-neutral-700 px-4 py-2 text-sm text-neutral-300 transition-colors hover:bg-neutral-800"
        >
          打开章节源页
        </a>

        <button
          v-if="imageResult.hasNext && imageResult.nextUrl"
          class="cursor-pointer rounded-xl border border-sky-500/50 bg-sky-500/10 px-4 py-2 text-sm text-sky-100 transition-colors hover:bg-sky-500/20"
          @click="openNextChapter"
        >
          下一话
        </button>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { GetImageDataUrl, LoadActiveLibrary } from '../../../wailsjs/go/library/API'
import { DownloadSourceChapter, GetSourceImages } from '../../../wailsjs/go/source/API'

interface SourceSummary {
  id: string
  name: string
  type: string
  language: string
  website: string
  description: string
}

interface ImageResult {
  source: SourceSummary
  comicTitle: string
  chapterTitle: string
  chapterUrl: string
  images: string[]
  hasNext: boolean
  nextUrl: string
}

interface DownloadChapterResult {
  source: SourceSummary
  comicTitle: string
  chapterTitle: string
  saveDir: string
  fileCount: number
}

function createEmptyImageResult(): ImageResult {
  return {
    source: {
      id: '',
      name: '',
      type: '',
      language: '',
      website: '',
      description: '',
    },
    comicTitle: '',
    chapterTitle: '',
    chapterUrl: '',
    images: [],
    hasNext: false,
    nextUrl: '',
  }
}

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const downloading = ref(false)
const errorMessage = ref('')
const imageResult = ref<ImageResult>(createEmptyImageResult())
const resolvedImages = ref<string[]>([])

const sourceId = computed(() => String(route.query.source ?? '').trim())
const chapterId = computed(() => String(route.query.chapter ?? '').trim())

onMounted(() => {
  void loadImages()
})

watch(
  () => [sourceId.value, chapterId.value],
  () => {
    void loadImages()
  },
)

async function loadImages() {
  if (!sourceId.value || !chapterId.value) {
    errorMessage.value = '缺少在线阅读参数，无法加载章节图片。'
    return
  }

  loading.value = true
  errorMessage.value = ''
  resolvedImages.value = []

  try {
    imageResult.value = await GetSourceImages(sourceId.value, chapterId.value)
    resolvedImages.value = await resolveImageUrls(imageResult.value.images)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '加载图片失败，请稍后再试。'
  } finally {
    loading.value = false
  }
}

async function resolveImageUrls(images: string[]) {
  const imagePromises = images.map(async (image) => {
    const normalized = String(image || '').trim()
    if (!normalized) {
      return null
    }

    if (!isLocalFilePath(normalized)) {
      return normalized
    }

    try {
      return await GetImageDataUrl(normalized)
    } catch (error) {
      console.error(`加载本地缓存图片失败: ${normalized}`, error)
      return null
    }
  })

  const loadedImages = await Promise.all(imagePromises)
  return loadedImages.filter((image): image is string => Boolean(image))
}

function isLocalFilePath(value: string) {
  if (!value) {
    return false
  }

  if (value.startsWith('data:') || value.startsWith('http://') || value.startsWith('https://')) {
    return false
  }

  return /^[a-zA-Z]:[\\/]/.test(value) || value.startsWith('\\\\')
}

async function downloadCurrentChapter() {
  if (!sourceId.value || !chapterId.value) {
    return
  }

  downloading.value = true
  try {
    const result = (await DownloadSourceChapter(
      sourceId.value,
      chapterId.value,
    )) as DownloadChapterResult
    await LoadActiveLibrary()
    toast.success('章节已下载到本地漫画库', {
      description: `${result.chapterTitle} · ${result.fileCount} 个文件`,
    })
  } catch (error) {
    toast.error(error instanceof Error ? error.message : '下载章节失败，请稍后再试。')
  } finally {
    downloading.value = false
  }
}

function openNextChapter() {
  if (!imageResult.value.nextUrl) {
    return
  }

  router.push({
    path: '/online/read',
    query: {
      source: imageResult.value.source.id,
      chapter: imageResult.value.nextUrl,
    },
  })
}
</script>

<style scoped></style>
