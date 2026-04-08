<template>
  <div class="flex flex-col gap-8 p-8 text-white">
    <div class="flex flex-col gap-4">
      <div class="text-xl">下载目录</div>
      <div class="flex gap-4">
        <Input
          v-model="downloadDir"
          class="flex-1 cursor-pointer"
          placeholder="请选择下载目录"
          @click="changeOutputDir"
        />
      </div>
    </div>

    <div class="flex flex-col gap-4">
      <div class="text-xl">漫画库</div>
      <div class="flex flex-wrap gap-2">
        <div v-for="library in libraries" :key="library" class="flex items-center gap-2">
          <button
            :class="{ 'bg-neutral-500/50': library === activeLibrary }"
            class="cursor-pointer rounded-2xl border border-neutral-300/50 px-4 py-2 hover:bg-neutral-500/50"
            @click="changeActiveLibrary(library)"
          >
            {{ library }}
          </button>
        </div>
      </div>
      <div class="flex justify-end">
        <button class="rounded-2xl border border-neutral-300/50 px-4 py-2" @click="addLibrary">
          添加漫画库
        </button>
      </div>
    </div>

    <div class="flex flex-col gap-4">
      <div class="text-xl">代理设置</div>
      <div class="flex gap-4">
        <Input
          v-model="proxyUrl"
          class="flex-1"
          placeholder="请输入代理地址，例如 http://127.0.0.1:10808"
          @blur="saveProxy"
        />
      </div>
    </div>

    <div class="flex flex-col gap-4">
      <div class="text-xl">日志</div>
      <div class="flex flex-col gap-2 text-neutral-300/90">
        <div>目录：<span class="select-all">{{ logInfo?.dir || '-' }}</span></div>
        <div>当前文件：<span class="select-all">{{ logInfo?.currentFile || '-' }}</span></div>
        <div>大小：{{ formatSize(logInfo?.sizeBytes) }}</div>
      </div>
      <div class="flex gap-2">
        <button
          class="rounded-2xl border border-neutral-300/50 px-4 py-2"
          @click="copyText(logInfo?.currentFile)"
        >
          复制日志文件路径
        </button>
        <button
          class="rounded-2xl border border-neutral-300/50 px-4 py-2"
          @click="copyText(logInfo?.dir)"
        >
          复制日志目录
        </button>
      </div>
    </div>

    <div class="flex flex-col gap-4">
      <div class="text-xl">Links Tips</div>
      <div class="rounded-2xl border border-neutral-300/20 bg-neutral-900/40 p-4">
        <div class="mb-3 text-sm text-neutral-400">
          建议只使用“具体作品页 / 画廊页 / 文章页”链接，不要使用首页、分类页、搜索结果页、标签页或作者页。
        </div>
        <div class="mb-4 rounded-xl border border-amber-300/20 bg-amber-400/5 p-3 text-xs text-neutral-300">
          <div>1. 先复制浏览器地址栏里的详情页，再粘贴到下载页。</div>
          <div>2. 403、挑战页、登录限制、站点改版都会导致失败。</div>
          <div>3. 18Comic 当前适配较旧，只建议尝试 `photo/...` 这种作品页。</div>
          <div>4. 如果报 unsupported site、403 或找不到图片，优先检查链接类型是否正确。</div>
        </div>
        <div class="flex flex-col gap-3">
          <div
            v-for="tip in linkTips"
            :key="tip.name"
            class="rounded-xl border border-neutral-300/10 bg-neutral-950/40 p-3"
          >
            <div class="mb-1 flex items-center justify-between gap-3">
              <div class="text-sm font-medium text-white">{{ tip.name }}</div>
              <button
                class="rounded-xl border border-neutral-300/30 px-3 py-1 text-xs text-neutral-200"
                @click="copyText(tip.template)"
              >
                Copy
              </button>
            </div>
            <div class="select-all break-all font-mono text-xs text-neutral-300">
              {{ tip.template }}
            </div>
            <div class="mt-3 grid gap-2 text-xs text-neutral-400">
              <div><span class="text-neutral-200">页面类型：</span>{{ tip.pageType }}</div>
              <div><span class="text-neutral-200">不要用：</span>{{ tip.avoid }}</div>
              <div><span class="text-neutral-200">备注：</span>{{ tip.note }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Input } from '@/components'
import { debounce } from '@/utils'
import { onMounted, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  AddLibrary,
  GetActiveLibrary,
  GetLibraries,
  GetOutputDir,
  GetProxy,
  SetActiveLibrary,
  SetOutputDir,
  SetProxy,
} from '../../../wailsjs/go/config/API'
import { LoadLibrary } from '../../../wailsjs/go/library/API'
import { GetLogInfo } from '../../../wailsjs/go/logger/API'

type LinkTip = {
  name: string
  template: string
  pageType: string
  avoid: string
  note: string
}

const linkTips: LinkTip[] = [
  {
    name: 'E-Hentai',
    template: 'https://e-hentai.org/g/{gallery-id}/{token}/',
    pageType: '具体 gallery 页',
    avoid: '首页、搜索结果页、tag 页、收藏列表页',
    note: '程序会继续翻分页并解析每张图的真实地址。',
  },
  {
    name: 'ExHentai',
    template: 'https://exhentai.org/g/{gallery-id}/{token}/',
    pageType: '具体 gallery 页',
    avoid: '首页、搜索页、排行榜页',
    note: '通常要求站点本身可访问；没有访问权限时会直接失败。',
  },
  {
    name: 'Telegraph',
    template: 'https://telegra.ph/{slug}',
    pageType: '具体文章页',
    avoid: '首页、频道页、跳转页',
    note: '当前逻辑是直接抓文章里的全部 img。',
  },
  {
    name: 'Telegraph Mirror',
    template: 'https://telegraph.com/{slug}',
    pageType: '具体文章页',
    avoid: '首页、非文章落地页',
    note: '代码里注册了这个域名，但实战上 telegra.ph 更常见。',
  },
  {
    name: 'WNACG',
    template: 'https://www.wnacg.com/photos-index-aid-{id}.html',
    pageType: '具体本子详情页',
    avoid: '首页、目录页、标签页、搜索页',
    note: '会翻分页再进入每页图片链接，比较依赖当前页面结构。',
  },
  {
    name: 'nhentai',
    template: 'https://nhentai.xxx/g/{id}/',
    pageType: '具体 gallery 页，路径需包含 /g/{id}/',
    avoid: '首页、随机页、列表页',
    note: '代码对这个路径格式要求很明确，建议直接用作品详情页链接。',
  },
  {
    name: 'Hitomi',
    template: 'https://hitomi.la/{category}/{slug}-{id}.html',
    pageType: '具体作品 html 页，结尾需像 -123456.html',
    avoid: '首页、tag 页、系列列表页',
    note: '程序会从作品 ID 生成真实图片地址，并带 Referer 下载。',
  },
  {
    name: '18Comic',
    template: 'https://18comic.vip/photo/{id}',
    pageType: '具体 photo 作品页',
    avoid: '首页、分类页、演员页、搜索结果页',
    note: '当前只认 .scramble-page > img，适配较旧，403 概率较高。',
  },
  {
    name: '18Comic Mirror',
    template: 'https://18comic.org/photo/{id}',
    pageType: '具体 photo 作品页',
    avoid: '首页、列表页、频道页',
    note: '逻辑和 18comic.vip 一样，只是换了 host。',
  },
]

const proxyUrl = ref('')
const downloadDir = ref('')
const libraries = ref<string[]>([])
const activeLibrary = ref('')
const logInfo = ref<any>(null)

const saveProxy = debounce((e: Event) => {
  SetProxy((e.target as HTMLInputElement).value).then(() => {
    void refreshConfig()
  })
}, 1000)

async function refreshConfig() {
  proxyUrl.value = await GetProxy()
  downloadDir.value = await GetOutputDir()
  libraries.value = await GetLibraries()
  activeLibrary.value = await GetActiveLibrary()

  try {
    logInfo.value = await GetLogInfo()
  } catch {
    logInfo.value = null
  }
}

async function changeOutputDir() {
  const ok = await SetOutputDir()
  if (!ok) return

  toast.success('设置成功')
  await refreshConfig()
}

async function changeActiveLibrary(library: string) {
  const ok = await SetActiveLibrary(library)
  if (!ok) return

  toast.success('设置成功')
  await refreshConfig()
}

async function addLibrary() {
  const ok = await AddLibrary()
  if (!ok) return

  toast.success('添加成功')
  await refreshConfig()

  if (activeLibrary.value) {
    await LoadLibrary(activeLibrary.value)
    toast.success('加载成功')
  }
}

onMounted(async () => {
  await refreshConfig()
})

function copyText(text?: string) {
  if (!text) return

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
