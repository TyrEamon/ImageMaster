import { useLibraryMetaStore } from '@/stores/libraryMetaStore'

export interface MangaProgress {
  scrollPosition: number
  timestamp: number
  totalImages: number
  progressPercent: number
  lastReadImage: number
  completed: boolean
}

export class ProgressService {
  constructor() {
    ProgressService.cleanupOldProgress()
  }

  static saveProgress(
    mangaPath: string,
    scrollPosition: number,
    totalImages: number,
    progressPercent = 0,
    lastReadImage = 0,
    completed = false,
  ): void {
    const libraryMetaStore = useLibraryMetaStore()

    libraryMetaStore.setReadingProgress(mangaPath, {
      scrollPosition,
      totalImages,
      progressPercent,
      lastReadImage,
      completed,
    })
  }

  static getProgress(mangaPath: string): MangaProgress | null {
    const libraryMetaStore = useLibraryMetaStore()
    return libraryMetaStore.getReadingProgress(mangaPath)
  }

  static removeProgress(mangaPath: string): void {
    const libraryMetaStore = useLibraryMetaStore()
    libraryMetaStore.removeReadingProgress(mangaPath)
  }

  static cleanupOldProgress(): void {
    const libraryMetaStore = useLibraryMetaStore()
    libraryMetaStore.cleanupOldReadingProgress()
  }

  static hasProgress(mangaPath: string): boolean {
    const progress = ProgressService.getProgress(mangaPath)
    return progress !== null && progress.progressPercent > 0
  }
}
