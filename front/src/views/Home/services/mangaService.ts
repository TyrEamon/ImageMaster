import { useLibraryMetaStore } from '@/stores/libraryMetaStore'
import {
  DeleteManga,
  GetAllMangas,
  GetImageDataUrl,
  LoadActiveLibrary,
} from '../../../../wailsjs/go/library/API'
import type { Manga } from '../stores/homeStore'
import { useHomeStore } from '../stores/homeStore'

export class MangaService {
  private homeStore: ReturnType<typeof useHomeStore>

  constructor() {
    this.homeStore = useHomeStore()
  }

  async initialize(): Promise<void> {
    this.homeStore.loading = true
    await this.loadMangas()
    this.homeStore.loading = false
  }

  async loadMangas(): Promise<void> {
    this.homeStore.loading = true
    await LoadActiveLibrary()

    const mangasData = await GetAllMangas()
    this.homeStore.mangas = mangasData

    const imageCache = this.homeStore.mangaImages
    for (const manga of mangasData) {
      if (!imageCache.has(manga.previewImg)) {
        const imageUrl = await GetImageDataUrl(manga.previewImg)
        imageCache.set(manga.previewImg, imageUrl)
      }
    }

    this.homeStore.mangaImages = imageCache
    this.homeStore.loading = false
  }

  async deleteManga(manga: Manga): Promise<boolean> {
    if (!confirm(`确定要删除 "${manga.name}" 吗？这会永久删除该文件夹及其内容。`)) {
      return false
    }

    this.homeStore.loading = true
    const success = await DeleteManga(manga.path)

    if (success) {
      useLibraryMetaStore().removeMangaState(manga.path)
      this.homeStore.mangas = this.homeStore.mangas.filter((item) => item.path !== manga.path)
    } else {
      alert('删除失败。')
    }

    this.homeStore.loading = false
    return success
  }

  getMangaImage(previewImg: string): string {
    return this.homeStore.mangaImages.get(previewImg) || ''
  }
}
