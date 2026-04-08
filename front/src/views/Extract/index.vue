<template>
  <div class="flex h-screen flex-col gap-6 overflow-hidden p-8 text-white">
    <div class="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
      <div class="max-w-3xl">
        <h1 class="text-2xl font-semibold">解压管理</h1>
        <p class="mt-2 text-sm leading-6 text-neutral-400">
          扫描已配置漫画库中的压缩包，按“已有子文件夹或已有图片文件”的规则判断是否已解压。
        </p>
      </div>

      <div class="flex flex-wrap gap-2">
        <Button :disabled="loading || extracting" @click="loadArchives">刷新扫描</Button>
        <Button
          type="primary"
          :disabled="loading || extracting || pendingCount === 0 || !scanResult?.bandizipPath"
          @click="extractPendingArchives"
        >
          一键批量解压
        </Button>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <div class="rounded-2xl border border-neutral-700/70 bg-neutral-900/70 p-4">
        <div class="text-xs uppercase tracking-wide text-neutral-500">扫描根目录</div>
        <div class="mt-3 text-2xl font-semibold">{{ rootsCount }}</div>
        <div class="mt-2 text-xs text-neutral-400">活动漫画库会优先参与扫描。</div>
      </div>

      <div class="rounded-2xl border border-neutral-700/70 bg-neutral-900/70 p-4">
        <div class="text-xs uppercase tracking-wide text-neutral-500">待解压</div>
        <div class="mt-3 text-2xl font-semibold text-amber-300">{{ pendingCount }}</div>
        <div class="mt-2 text-xs text-neutral-400">批量解压只会处理这些压缩包。</div>
      </div>

      <div class="rounded-2xl border border-neutral-700/70 bg-neutral-900/70 p-4">
        <div class="text-xs uppercase tracking-wide text-neutral-500">判定已解压</div>
        <div class="mt-3 text-2xl font-semibold text-emerald-300">{{ extractedCount }}</div>
        <div class="mt-2 text-xs text-neutral-400">已有子目录或图片时会自动跳过。</div>
      </div>

      <div class="rounded-2xl border border-neutral-700/70 bg-neutral-900/70 p-4">
        <div class="text-xs uppercase tracking-wide text-neutral-500">Bandizip</div>
        <div class="mt-3 truncate text-sm font-medium">
          {{ scanResult?.bandizipPath || '未检测到，请去 Setting 配置 bz.exe 路径' }}
        </div>
        <div class="mt-2 text-xs text-neutral-400">当前版本通过外部 Bandizip 控制台完成解压。</div>
      </div>
    </div>

    <div class="rounded-2xl border border-neutral-700/70 bg-neutral-900/60 p-4">
      <div class="flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
        <div class="flex flex-wrap gap-2">
          <button
            v-for="option in filterOptions"
            :key="option.value"
            class="cursor-pointer rounded-xl border px-3 py-1.5 text-xs transition-colors duration-200"
            :class="
              statusFilter === option.value
                ? 'border-blue-400/70 bg-blue-500/10 text-blue-100'
                : 'border-neutral-700 text-neutral-300 hover:bg-neutral-800'
            "
            @click="statusFilter = option.value"
          >
            {{ option.label }} {{ option.count }}
          </button>
        </div>

        <div class="text-xs text-neutral-400">
          当前活动漫画库：
          <span class="select-all text-neutral-200">{{ scanResult?.activeLibrary || '-' }}</span>
        </div>
      </div>

      <div
        v-if="scanResult?.roots?.length"
        class="mt-4 flex flex-col gap-2 text-xs text-neutral-400"
      >
        <div class="text-neutral-500">扫描路径</div>
        <div
          v-for="root in scanResult.roots"
          :key="root"
          class="select-all rounded-xl border border-neutral-800 bg-neutral-950/70 px-3 py-2"
        >
          {{ root }}
        </div>
      </div>
    </div>

    <div
      v-if="scanResult && !scanResult.bandizipPath"
      class="rounded-2xl border border-amber-300/20 bg-amber-400/5 px-4 py-3 text-sm text-amber-100"
    >
      还没检测到可用的 Bandizip 控制台工具。你可以先到 Setting 里填入 `bz.exe`
      的本地路径，再回来批量解压。
    </div>

    <div
      class="min-h-0 flex-1 overflow-auto rounded-2xl border border-neutral-700/70 bg-neutral-900/60"
    >
      <div v-if="loading" class="flex h-72 items-center justify-center text-sm text-neutral-400">
        正在扫描压缩包...
      </div>

      <div
        v-else-if="!scanResult?.roots?.length"
        class="flex h-72 items-center justify-center text-sm text-neutral-400"
      >
        还没有可扫描的漫画库。先去 Setting 设置下载目录或添加漫画库目录。
      </div>

      <div
        v-else-if="visibleItems.length === 0"
        class="flex h-72 items-center justify-center text-sm text-neutral-400"
      >
        当前筛选条件下没有记录。
      </div>

      <div v-else class="space-y-4 p-4">
        <div
          v-for="item in visibleItems"
          :key="item.archivePath"
          class="rounded-2xl border border-neutral-800 bg-neutral-950/70 p-4 transition-colors duration-200 hover:border-neutral-700"
        >
          <div class="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
            <div class="min-w-0 flex-1">
              <div class="flex flex-wrap items-center gap-2">
                <div
                  class="inline-flex rounded-full border px-2 py-1 text-[11px]"
                  :class="statusClassMap[item.status] || statusClassMap.failed"
                >
                  {{ statusLabelMap[item.status] || item.status }}
                </div>
                <div
                  class="rounded-full border border-neutral-700 px-2 py-1 text-[11px] text-neutral-300"
                >
                  {{ formatSize(item.sizeBytes) }}
                </div>
                <div
                  v-if="item.reason"
                  class="rounded-full border border-neutral-700 px-2 py-1 text-[11px] text-neutral-400"
                >
                  {{ item.reason }}
                </div>
              </div>

              <div class="mt-3 break-words text-base font-semibold text-neutral-100">
                {{ item.archiveName }}
              </div>
            </div>

            <div class="flex flex-wrap gap-2 xl:justify-end">
              <button
                class="cursor-pointer rounded-xl border border-neutral-700 px-3 py-2 text-[11px] text-neutral-200 transition-colors duration-200 hover:bg-neutral-800 disabled:cursor-not-allowed disabled:opacity-40"
                :disabled="extracting"
                @click="copyPath(item.archivePath)"
              >
                复制压缩包路径
              </button>

              <button
                class="cursor-pointer rounded-xl border border-neutral-700 px-3 py-2 text-[11px] text-neutral-200 transition-colors duration-200 hover:bg-neutral-800 disabled:cursor-not-allowed disabled:opacity-40"
                :disabled="extracting"
                @click="copyPath(item.targetDir)"
              >
                复制目标目录
              </button>

              <button
                class="cursor-pointer rounded-xl border border-blue-500/40 px-3 py-2 text-[11px] text-blue-100 transition-colors duration-200 hover:bg-blue-500/10 disabled:cursor-not-allowed disabled:opacity-40"
                :disabled="extracting || item.status !== 'pending' || !scanResult?.bandizipPath"
                @click="extractSingle(item)"
              >
                解压
              </button>
            </div>
          </div>

          <div class="mt-4 grid gap-3 xl:grid-cols-3">
            <div class="rounded-xl border border-neutral-800 bg-neutral-900/70 p-3">
              <div class="text-[11px] uppercase tracking-wide text-neutral-500">压缩包路径</div>
              <div class="mt-2 break-all text-xs leading-5 text-neutral-300">
                {{ item.archivePath }}
              </div>
            </div>

            <div class="rounded-xl border border-neutral-800 bg-neutral-900/70 p-3">
              <div class="text-[11px] uppercase tracking-wide text-neutral-500">漫画库</div>
              <div class="mt-2 break-all text-xs leading-5 text-neutral-300">
                {{ item.libraryPath }}
              </div>
            </div>

            <div class="rounded-xl border border-neutral-800 bg-neutral-900/70 p-3">
              <div class="text-[11px] uppercase tracking-wide text-neutral-500">目标目录</div>
              <div class="mt-2 break-all text-xs leading-5 text-neutral-300">
                {{ item.targetDir }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Button } from '@/components'
import { computed, onMounted, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  ExtractArchive,
  ExtractPendingArchives,
  ScanArchives,
} from '../../../wailsjs/go/archive/API'
import { LoadActiveLibrary } from '../../../wailsjs/go/library/API'

type StatusFilter = 'all' | 'pending' | 'extracted' | 'failed'

interface ArchiveItem {
  archivePath: string
  archiveName: string
  libraryPath: string
  targetDir: string
  status: string
  reason: string
  sizeBytes: number
}

interface ScanResult {
  roots: string[]
  activeLibrary: string
  bandizipPath: string
  totalCount: number
  pendingCount: number
  extractedCount: number
  failedCount: number
  items: ArchiveItem[]
}

interface ExtractSummary {
  totalCount: number
  extractedCount: number
  skippedCount: number
  failedCount: number
}

interface ExtractActionResult {
  status: string
  message: string
}

const scanResult = ref<ScanResult | null>(null)
const loading = ref(false)
const extracting = ref(false)
const statusFilter = ref<StatusFilter>('pending')

const pendingCount = computed(() => scanResult.value?.pendingCount ?? 0)
const extractedCount = computed(() => scanResult.value?.extractedCount ?? 0)
const failedCount = computed(() => scanResult.value?.failedCount ?? 0)
const rootsCount = computed(() => scanResult.value?.roots?.length ?? 0)

const filterOptions = computed(() => [
  { value: 'pending' as const, label: '待解压', count: pendingCount.value },
  { value: 'extracted' as const, label: '判定已解压', count: extractedCount.value },
  { value: 'failed' as const, label: '异常', count: failedCount.value },
  { value: 'all' as const, label: '全部', count: scanResult.value?.totalCount ?? 0 },
])

const visibleItems = computed(() => {
  if (!scanResult.value) {
    return [] as ArchiveItem[]
  }

  if (statusFilter.value === 'all') {
    return scanResult.value.items
  }

  return scanResult.value.items.filter((item: ArchiveItem) => item.status === statusFilter.value)
})

const statusLabelMap: Record<string, string> = {
  pending: '待解压',
  extracted: '判定已解压',
  failed: '异常',
}

const statusClassMap: Record<string, string> = {
  pending: 'border-amber-400/40 bg-amber-400/10 text-amber-100',
  extracted: 'border-emerald-400/40 bg-emerald-400/10 text-emerald-100',
  failed: 'border-red-400/40 bg-red-400/10 text-red-100',
}

onMounted(async () => {
  await loadArchives()
})

async function loadArchives() {
  loading.value = true
  try {
    scanResult.value = (await ScanArchives()) as ScanResult
  } catch (error: any) {
    toast.error(error?.message || '扫描压缩包失败')
  } finally {
    loading.value = false
  }
}

async function extractPendingArchives() {
  extracting.value = true
  try {
    const summary = (await ExtractPendingArchives()) as ExtractSummary

    if (summary.totalCount === 0) {
      toast.info('没有待解压的压缩包')
      return
    }

    if (summary.extractedCount > 0) {
      await LoadActiveLibrary()
    }

    toast.success('批量解压完成', {
      description: `成功 ${summary.extractedCount}，跳过 ${summary.skippedCount}，失败 ${summary.failedCount}`,
    })
  } catch (error: any) {
    toast.error(error?.message || '批量解压失败')
  } finally {
    extracting.value = false
    await loadArchives()
  }
}

async function extractSingle(item: ArchiveItem) {
  extracting.value = true
  try {
    const result = (await ExtractArchive(item.archivePath)) as ExtractActionResult

    if (result.status === 'extracted') {
      await LoadActiveLibrary()
      toast.success('解压完成', {
        description: item.archiveName,
      })
    } else {
      toast.info(result.message)
    }
  } catch (error: any) {
    toast.error(error?.message || '解压失败')
  } finally {
    extracting.value = false
    await loadArchives()
  }
}

function copyPath(text: string) {
  navigator.clipboard.writeText(text).then(() => {
    toast.success('已复制到剪贴板')
  })
}

function formatSize(size?: number) {
  if (!size && size !== 0) return '-'

  const units = ['B', 'KB', 'MB', 'GB']
  let index = 0
  let currentSize = size

  while (currentSize >= 1024 && index < units.length - 1) {
    currentSize /= 1024
    index++
  }

  return `${currentSize.toFixed(2)} ${units[index]}`
}
</script>

<style scoped></style>
