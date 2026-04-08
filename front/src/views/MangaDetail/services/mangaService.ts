import { useLibraryMetaStore } from '@/stores/libraryMetaStore'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import {
  DeleteManga,
  GetAllMangas,
  GetImageDataUrl,
  GetMangaImages,
} from '../../../../wailsjs/go/library/API'
import { useMangaStore } from '../stores'
import type { ScrollService } from './scrollService'

export class MangaService {
  private router: ReturnType<typeof useRouter>
  private mangaStore: ReturnType<typeof useMangaStore>
  private scrollService: ScrollService

  constructor(scrollService: ScrollService) {
    this.scrollService = scrollService
    this.mangaStore = useMangaStore()
    this.router = useRouter()
  }

  async loadManga(path: string) {
    try {
      this.mangaStore.loading = true

      const mangaPath = decodeURIComponent(path)
      const mangas = await GetAllMangas()
      const currentMangaIndex = mangas.findIndex((manga) => manga.path === mangaPath)

      const mangaName =
        currentMangaIndex >= 0 ? mangas[currentMangaIndex].name : mangaPath.split('/').pop() || ''

      this.mangaStore.updateMangaStore({
        mangaPath,
        mangaName,
        mangas,
        currentMangaIndex,
        selectedImages: [],
      })

      await this.loadImages(mangaPath)
    } catch (error) {
      console.error('加载漫画失败:', error)
    } finally {
      this.mangaStore.loading = false
      this.scrollService.restoreScrollPosition()
    }
  }

  async loadImages(mangaPath: string) {
    try {
      const imagePaths = await GetMangaImages(mangaPath)
      const imagePromises = imagePaths.map(async (imagePath) => {
        try {
          return await GetImageDataUrl(imagePath)
        } catch (error) {
          console.error(`加载图片失败: ${imagePath}`, error)
          return null
        }
      })

      const loadedImages = await Promise.all(imagePromises)
      const selectedImages = loadedImages.filter((image): image is string => image !== null)
      this.mangaStore.updateMangaStore({ selectedImages })
    } catch (error) {
      console.error('获取图片路径失败:', error)
    }
  }

  backToHome() {
    this.router.push('/')
  }

  navigateToNextManga() {
    if (this.mangaStore.currentMangaIndex >= this.mangaStore.mangas.length - 1) {
      return
    }

    const nextManga = this.mangaStore.mangas[this.mangaStore.currentMangaIndex + 1]
    this.navigateWithinReader(nextManga.path)
  }

  navigateToPrevManga() {
    if (this.mangaStore.currentMangaIndex <= 0) {
      return
    }

    const prevManga = this.mangaStore.mangas[this.mangaStore.currentMangaIndex - 1]
    this.navigateWithinReader(prevManga.path)
  }

  async deleteAndViewNextManga() {
    if (
      this.mangaStore.currentMangaIndex < 0 ||
      !confirm(`确定要删除 "${this.mangaStore.mangaName}" 并查看下一部漫画吗？`)
    ) {
      return
    }

    this.mangaStore.loading = true

    try {
      const hasNextManga = this.mangaStore.currentMangaIndex < this.mangaStore.mangas.length - 1
      const nextMangaPath = hasNextManga
        ? this.mangaStore.mangas[this.mangaStore.currentMangaIndex + 1].path
        : null

      const success = await DeleteManga(this.mangaStore.mangaPath)

      if (!success) {
        alert('删除失败!')
        this.mangaStore.loading = false
        return
      }

      useLibraryMetaStore().removeMangaState(this.mangaStore.mangaPath)
      toast.success('删除成功')

      if (nextMangaPath) {
        this.mangaStore.loading = false
        this.navigateWithinReader(nextMangaPath)
      } else {
        this.router.push('/')
      }
    } catch (error) {
      console.error('删除漫画失败:', error)
      this.mangaStore.loading = false
    }
  }

  private navigateWithinReader(path: string) {
    const encodedPath = encodeURIComponent(path)
    const currentLocation = window.location.href

    if (currentLocation.includes('/manga/')) {
      this.loadManga(path)
      window.history.pushState(null, '', `/#/manga/${encodedPath}`)
      return
    }

    this.router.push(`/manga/${encodedPath}`)
  }
}
