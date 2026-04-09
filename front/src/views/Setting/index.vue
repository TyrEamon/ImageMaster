<template>
  <div class="flex flex-col gap-8 p-8 text-white">
    <section class="flex flex-col gap-4">
      <div class="text-xl">下载目录</div>
      <div class="flex gap-4">
        <Input
          v-model="downloadDir"
          class="flex-1 cursor-pointer"
          placeholder="点击选择下载目录"
          @click="changeOutputDir"
        />
      </div>
    </section>

    <section class="flex flex-col gap-4">
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
    </section>

    <section class="flex flex-col gap-4">
      <div class="text-xl">代理设置</div>
      <div class="flex gap-4">
        <Input
          v-model="proxyUrl"
          class="flex-1"
          placeholder="例如 http://127.0.0.1:10808"
          @blur="saveProxy"
        />
      </div>
    </section>

    <section class="flex flex-col gap-4">
      <div class="text-xl">解压设置</div>
      <div class="flex gap-4">
        <Input
          v-model="bandizipPath"
          class="flex-1"
          placeholder="输入 bz.exe 路径，留空则自动检测常见安装位置"
          @blur="saveBandizipPath"
        />
      </div>
      <div class="text-xs leading-6 text-neutral-400">
        当前解压功能依赖外部 Bandizip 控制台工具。推荐填写类似
        <span class="select-all text-neutral-200">D:\bandizip\bz.exe</span>
        的可执行文件路径；如果填写的是
        <span class="select-all text-neutral-200">Bandizip.exe</span>
        ，软件会自动尝试同目录下的
        <span class="select-all text-neutral-200">bz.exe</span>
        。留空时会自动检测常见安装位置。
      </div>
    </section>

    <section class="flex flex-col gap-4">
      <div class="text-xl">源仓库</div>
      <div class="rounded-2xl border border-neutral-300/20 bg-neutral-900/40 p-4">
        <div class="flex flex-col gap-4">
          <Input
            v-model="sourceRepoUrl"
            class="w-full"
            placeholder="填写 GitHub 仓库地址，或直接填写远程 index.json 地址"
            @blur="saveSourceRepoUrl"
          />

          <div class="grid gap-4 md:grid-cols-2">
            <div class="rounded-xl border border-neutral-300/10 bg-neutral-950/40 p-3 text-xs leading-6 text-neutral-400">
              <div><span class="text-neutral-200">内置源目录：</span>{{ sourceStorageInfo?.bundledDir || '-' }}</div>
              <div><span class="text-neutral-200">本地源目录：</span>{{ sourceStorageInfo?.userDir || '-' }}</div>
            </div>

            <div class="rounded-xl border border-neutral-300/10 bg-neutral-950/40 p-3 text-xs leading-6 text-neutral-400">
              <div>同步时会把仓库里的 index、manifest 和脚本下载到本地源目录。</div>
              <div>以后你只要维护 GitHub 源仓库，软件同步后就能加载更新。</div>
            </div>
          </div>

          <div class="flex flex-wrap gap-2">
            <button class="rounded-2xl border border-neutral-300/50 px-4 py-2" @click="syncSourceRepo">
              同步源仓库
            </button>
            <button class="rounded-2xl border border-neutral-300/50 px-4 py-2" @click="reloadLocalSources">
              重载本地源
            </button>
            <button class="rounded-2xl border border-neutral-300/50 px-4 py-2" @click="copyText(sourceStorageInfo?.userDir)">
              复制本地源目录
            </button>
          </div>
        </div>
      </div>
    </section>

    <section class="flex flex-col gap-4">
      <div class="text-xl">JM 在线缓存</div>
      <div class="rounded-2xl border border-neutral-300/20 bg-neutral-900/40 p-4">
        <div class="grid gap-4 md:grid-cols-2">
          <div class="md:col-span-2">
            <Input
              v-model="jmCacheDir"
              class="w-full"
              placeholder="留空则使用系统临时目录；也可以改到 D:\\ImageMasterCache\\jm-reader"
              @blur="saveJmCacheDir"
            />
          </div>

          <div class="flex flex-col gap-2">
            <label class="text-xs text-neutral-400">保留时长（小时）</label>
            <Input
              v-model="jmCacheRetentionHours"
              type="number"
              min="1"
              placeholder="24"
              @blur="saveJmCacheRetentionHours"
            />
          </div>

          <div class="flex flex-col gap-2">
            <label class="text-xs text-neutral-400">总缓存上限（MB）</label>
            <Input
              v-model="jmCacheSizeLimitMB"
              type="number"
              min="128"
              step="128"
              placeholder="2048"
              @blur="saveJmCacheSizeLimitMB"
            />
          </div>
        </div>

        <div class="mt-4 text-xs leading-6 text-neutral-400">
          JM 在线阅读会先把当前章节解到临时缓存，再交给软件显示。默认使用系统临时目录，通常在
          <span class="text-neutral-200">C:\Users\你的用户名\AppData\Local\Temp</span>
          。如果你不想占用 C 盘，可以把缓存目录改到 D 盘或任意自定义路径。
        </div>
        <div class="mt-2 text-xs leading-6 text-neutral-400">
          清理规则是：先删除超过保留时长的旧章节缓存；如果总大小仍然超过上限，再按最旧缓存继续删除。
        </div>
      </div>
    </section>

    <section class="flex flex-col gap-4">
      <div class="text-xl">日志</div>
      <div class="flex flex-col gap-2 text-neutral-300/90">
        <div>目录：<span class="select-all">{{ logInfo?.dir || '-' }}</span></div>
        <div>当前文件：<span class="select-all">{{ logInfo?.currentFile || '-' }}</span></div>
        <div>大小：{{ formatSize(logInfo?.sizeBytes) }}</div>
      </div>
      <div class="flex gap-2">
        <button class="rounded-2xl border border-neutral-300/50 px-4 py-2" @click="copyText(logInfo?.currentFile)">
          复制日志文件路径
        </button>
        <button class="rounded-2xl border border-neutral-300/50 px-4 py-2" @click="copyText(logInfo?.dir)">
          复制日志目录
        </button>
      </div>
    </section>

    <section class="flex flex-col gap-4">
      <div class="text-xl">Links Tips</div>
      <div class="rounded-2xl border border-neutral-300/20 bg-neutral-900/40 p-4">
        <div class="mb-3 text-sm text-neutral-400">
          建议只使用“具体作品页 / 画廊页 / 文章页”链接，不要使用首页、分类页、搜索结果页、标签页或作者页。
        </div>
        <div class="mb-4 rounded-xl border border-amber-300/20 bg-amber-400/5 p-3 text-xs text-neutral-300">
          <div>1. 先复制浏览器地址栏里的详情页，再粘贴到下载页。</div>
          <div>2. 403、挑战页、登录限制、站点改版都会导致失败。</div>
          <div>3. 18Comic 当前适配较老，只建议尝试 `photo/...` 这种作品页。</div>
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
    </section>

    <section class="flex flex-col gap-4">
      <div class="text-xl">Version</div>
      <div class="rounded-2xl border border-neutral-300/20 bg-neutral-900/40 p-4">
        <div class="flex flex-col gap-2 text-neutral-300/90">
          <div>Current Version: <span class="select-all">{{ versionInfo?.display || '-' }}</span></div>
          <div>Commit: <span class="select-all">{{ versionInfo?.commit || '-' }}</span></div>
          <div>Build Time: <span class="select-all">{{ versionInfo?.buildTime || '-' }}</span></div>
        </div>
        <div class="mt-3 flex gap-2">
          <button class="rounded-2xl border border-neutral-300/50 px-4 py-2" @click="copyText(versionInfo?.display)">
            Copy Version
          </button>
          <button
            class="rounded-2xl border border-neutral-300/50 px-4 py-2"
            @click="
              copyText(
                versionInfo
                  ? `${versionInfo.display} | ${versionInfo.commit} | ${versionInfo.buildTime || 'no-build-time'}`
                  : '',
              )
            "
          >
            Copy Build Info
          </button>
        </div>
      </div>
    </section>

    <section class="flex flex-col gap-4">
      <div class="text-xl">JM Runtime</div>
      <div class="rounded-2xl border border-neutral-300/20 bg-neutral-900/40 p-4">
        <div class="flex flex-col gap-2 text-neutral-300/90">
          <div>
            Status:
            <span :class="jmRuntimeInfo?.available ? 'text-emerald-300' : 'text-amber-300'">
              {{ jmRuntimeInfo?.available ? 'Ready' : 'Not Ready' }}
            </span>
          </div>
          <div>Name: <span class="select-all">{{ jmRuntimeInfo?.name || '-' }}</span></div>
          <div>Version: <span class="select-all">{{ jmRuntimeInfo?.version || '-' }}</span></div>
          <div>Engine: <span class="select-all">{{ jmRuntimeInfo?.engine || '-' }}</span></div>
          <div>Source: <span class="select-all">{{ jmRuntimeInfo?.source || '-' }}</span></div>
          <div>Helper Path: <span class="select-all">{{ jmRuntimeInfo?.helperPath || '-' }}</span></div>
          <div>Build Time: <span class="select-all">{{ jmRuntimeInfo?.buildTime || '-' }}</span></div>
        </div>
        <div class="mt-3 text-xs leading-6 text-neutral-400">
          JM runtime is bundled as a helper under <span class="text-neutral-200">runtime/</span>.
          If this area shows <span class="text-neutral-200">Not Ready</span>, JM links will fall back to the legacy crawler.
        </div>
        <div class="mt-3 flex gap-2">
          <button class="rounded-2xl border border-neutral-300/50 px-4 py-2" @click="copyText(jmRuntimeInfo?.helperPath)">
            Copy Helper Path
          </button>
          <button
            class="rounded-2xl border border-neutral-300/50 px-4 py-2"
            @click="
              copyText(
                jmRuntimeInfo
                  ? `${jmRuntimeInfo.name} | ${jmRuntimeInfo.version} | ${jmRuntimeInfo.source || 'unknown'}`
                  : '',
              )
            "
          >
            Copy Runtime Info
          </button>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { Input } from '@/components'
import { onMounted, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  AddLibrary,
  GetActiveLibrary,
  GetBandizipPath,
  GetJmCacheDir,
  GetJmCacheRetentionHours,
  GetJmCacheSizeLimitMB,
  GetLibraries,
  GetOutputDir,
  GetProxy,
  GetSourceRepoURL,
  SetActiveLibrary,
  SetBandizipPath,
  SetJmCacheDir,
  SetJmCacheRetentionHours,
  SetJmCacheSizeLimitMB,
  SetOutputDir,
  SetProxy,
  SetSourceRepoURL,
} from '../../../wailsjs/go/config/API'
import { LoadLibrary } from '../../../wailsjs/go/library/API'
import { GetLogInfo } from '../../../wailsjs/go/logger/API'
import { GetJmRuntimeInfo, GetVersionInfo } from '../../../wailsjs/go/meta/API'
import { GetSourceStorageInfo, ReloadSources, SyncSourceRepository } from '../../../wailsjs/go/source/API'

type LinkTip = {
  name: string
  template: string
  pageType: string
  avoid: string
  note: string
}

type VersionInfo = {
  version: string
  display: string
  commit: string
  buildTime: string
  isDevBuild: boolean
}

type JmRuntimeInfo = {
  name: string
  version: string
  engine: string
  upstream: string
  buildTime: string
  manifestPath: string
  helperPath: string
  available: boolean
  source: string
}

type SourceStorageInfo = {
  bundledDir: string
  userDir: string
}

type SourceRepoSyncResult = {
  repoUrl: string
  indexUrl: string
  localDir: string
  manifestCount: number
  scriptCount: number
  enabledManifestCount: number
  updatedAt: string
}

const linkTips: LinkTip[] = [
  {
    name: 'E-Hentai',
    template: 'https://e-hentai.org/g/{gallery-id}/{token}/',
    pageType: '具体 gallery 页',
    avoid: '首页、搜索结果页、Tag 页、收藏列表页',
    note: '程序会继续翻分页并解析每张图的真实地址。',
  },
  {
    name: 'ExHentai',
    template: 'https://exhentai.org/g/{gallery-id}/{token}/',
    pageType: '具体 gallery 页',
    avoid: '首页、搜索页、排行页',
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
    note: '代码里注册了这个域名，但实战中 telegra.ph 更常见。',
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
    avoid: '首页、Tag 页、系列列表页',
    note: '程序会从作品 ID 生成真实图片地址，并带 Referer 下载。',
  },
  {
    name: '18Comic',
    template: 'https://18comic.vip/photo/{id}',
    pageType: '具体 photo 作品页',
    avoid: '首页、分类页、演员页、搜索结果页',
    note: '当前只认 .scramble-page > img，适配较老，403 概率较高。',
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
const bandizipPath = ref('')
const sourceRepoUrl = ref('')
const jmCacheDir = ref('')
const jmCacheRetentionHours = ref('24')
const jmCacheSizeLimitMB = ref('2048')
const downloadDir = ref('')
const libraries = ref<string[]>([])
const activeLibrary = ref('')
const logInfo = ref<any>(null)
const versionInfo = ref<VersionInfo | null>(null)
const jmRuntimeInfo = ref<JmRuntimeInfo | null>(null)
const sourceStorageInfo = ref<SourceStorageInfo | null>(null)

async function refreshConfig() {
  proxyUrl.value = await GetProxy()
  bandizipPath.value = await GetBandizipPath()
  sourceRepoUrl.value = await GetSourceRepoURL()
  jmCacheDir.value = await GetJmCacheDir()
  jmCacheRetentionHours.value = String(await GetJmCacheRetentionHours())
  jmCacheSizeLimitMB.value = String(await GetJmCacheSizeLimitMB())
  downloadDir.value = await GetOutputDir()
  libraries.value = await GetLibraries()
  activeLibrary.value = await GetActiveLibrary()

  try {
    logInfo.value = await GetLogInfo()
  } catch {
    logInfo.value = null
  }

  try {
    versionInfo.value = (await GetVersionInfo()) as VersionInfo
  } catch {
    versionInfo.value = null
  }

  try {
    jmRuntimeInfo.value = (await GetJmRuntimeInfo()) as JmRuntimeInfo
  } catch {
    jmRuntimeInfo.value = null
  }

  try {
    sourceStorageInfo.value = (await GetSourceStorageInfo()) as SourceStorageInfo
  } catch {
    sourceStorageInfo.value = null
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

async function saveProxy(event: Event) {
  const ok = await SetProxy((event.target as HTMLInputElement).value.trim())
  if (!ok) return

  toast.success('代理已保存')
  await refreshConfig()
}

async function saveBandizipPath(event: Event) {
  const ok = await SetBandizipPath((event.target as HTMLInputElement).value.trim())
  if (!ok) return

  toast.success('解压工具路径已保存')
  await refreshConfig()
}

async function saveSourceRepoUrl(event: Event) {
  const ok = await SetSourceRepoURL((event.target as HTMLInputElement).value.trim())
  if (!ok) return

  toast.success('源仓库地址已保存')
  await refreshConfig()
}

async function syncSourceRepo() {
  const repoURL = sourceRepoUrl.value.trim()
  if (!repoURL) {
    toast.error('请先填写源仓库地址')
    return
  }

  try {
    const result = (await SyncSourceRepository(repoURL)) as SourceRepoSyncResult
    toast.success('源仓库同步成功', {
      description: `已同步 ${result.manifestCount} 个清单，${result.scriptCount} 个脚本`,
    })
    await refreshConfig()
  } catch (error) {
    toast.error('源仓库同步失败', {
      description: error instanceof Error ? error.message : '请检查仓库地址或网络连接。',
    })
  }
}

async function reloadLocalSources() {
  try {
    const sources = await ReloadSources()
    toast.success('本地源已重载', {
      description: `当前共加载 ${sources.length} 个在线源`,
    })
    await refreshConfig()
  } catch (error) {
    toast.error('重载本地源失败', {
      description: error instanceof Error ? error.message : '请稍后再试。',
    })
  }
}

async function saveJmCacheDir(event: Event) {
  const ok = await SetJmCacheDir((event.target as HTMLInputElement).value.trim())
  if (!ok) return

  toast.success('JM 缓存目录已保存')
  await refreshConfig()
}

async function saveJmCacheRetentionHours(event: Event) {
  const hours = normalizePositiveInt((event.target as HTMLInputElement).value, 24)
  const ok = await SetJmCacheRetentionHours(hours)
  if (!ok) return

  toast.success('JM 缓存保留时长已保存')
  await refreshConfig()
}

async function saveJmCacheSizeLimitMB(event: Event) {
  const limit = normalizePositiveInt((event.target as HTMLInputElement).value, 2048)
  const ok = await SetJmCacheSizeLimitMB(limit)
  if (!ok) return

  toast.success('JM 缓存大小上限已保存')
  await refreshConfig()
}

onMounted(async () => {
  await refreshConfig()
})

function normalizePositiveInt(raw: string, fallback: number) {
  const parsed = Number.parseInt(raw, 10)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback
}

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
