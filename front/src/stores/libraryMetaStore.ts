import { defineStore } from 'pinia'

const SHELF_STORAGE_KEY = 'imagemaster:shelf-state:v1'
const PROGRESS_STORAGE_KEY = 'imagemaster:reading-progress:v2'
const READING_PROGRESS_RETENTION_DAYS = 90

export interface MangaShelfState {
  favorite: boolean
  pinned: boolean
  readLater: boolean
  updatedAt: number
}

export interface MangaReadingProgress {
  scrollPosition: number
  timestamp: number
  totalImages: number
  progressPercent: number
  lastReadImage: number
  completed: boolean
}

function createDefaultShelfState(): MangaShelfState {
  return {
    favorite: false,
    pinned: false,
    readLater: false,
    updatedAt: 0,
  }
}

function loadStorageObject<T>(key: string, fallback: T): T {
  if (typeof window === 'undefined') {
    return fallback
  }

  try {
    const rawValue = window.localStorage.getItem(key)
    return rawValue ? (JSON.parse(rawValue) as T) : fallback
  } catch {
    return fallback
  }
}

function persistStorageObject<T>(key: string, value: T) {
  if (typeof window === 'undefined') {
    return
  }

  window.localStorage.setItem(key, JSON.stringify(value))
}

export const useLibraryMetaStore = defineStore('libraryMeta', {
  state: () => ({
    shelfStates: loadStorageObject<Record<string, MangaShelfState>>(SHELF_STORAGE_KEY, {}),
    readingProgress: loadStorageObject<Record<string, MangaReadingProgress>>(PROGRESS_STORAGE_KEY, {}),
  }),

  getters: {
    getShelfState: (state) => (path: string): MangaShelfState => {
      return {
        ...createDefaultShelfState(),
        ...(state.shelfStates[path] ?? {}),
      }
    },

    getReadingProgress: (state) => (path: string): MangaReadingProgress | null => {
      return state.readingProgress[path] ?? null
    },
  },

  actions: {
    persistShelfStates() {
      persistStorageObject(SHELF_STORAGE_KEY, this.shelfStates)
    },

    persistReadingProgress() {
      persistStorageObject(PROGRESS_STORAGE_KEY, this.readingProgress)
    },

    cleanupOldReadingProgress() {
      const threshold = Date.now() - READING_PROGRESS_RETENTION_DAYS * 24 * 60 * 60 * 1000
      const nextProgress: Record<string, MangaReadingProgress> = {}

      for (const [path, progress] of Object.entries(this.readingProgress)) {
        if (progress.timestamp >= threshold) {
          nextProgress[path] = progress
        }
      }

      this.readingProgress = nextProgress
      this.persistReadingProgress()
    },

    setShelfState(path: string, updates: Partial<MangaShelfState>) {
      const nextState: MangaShelfState = {
        ...this.getShelfState(path),
        ...updates,
        updatedAt: Date.now(),
      }

      this.shelfStates = {
        ...this.shelfStates,
        [path]: nextState,
      }
      this.persistShelfStates()
    },

    removeShelfState(path: string) {
      if (!(path in this.shelfStates)) {
        return
      }

      const nextStates = { ...this.shelfStates }
      delete nextStates[path]
      this.shelfStates = nextStates
      this.persistShelfStates()
    },

    toggleFavorite(path: string) {
      const shelfState = this.getShelfState(path)
      this.setShelfState(path, { favorite: !shelfState.favorite })
    },

    togglePinned(path: string) {
      const shelfState = this.getShelfState(path)
      this.setShelfState(path, { pinned: !shelfState.pinned })
    },

    toggleReadLater(path: string) {
      const shelfState = this.getShelfState(path)
      this.setShelfState(path, { readLater: !shelfState.readLater })
    },

    setReadingProgress(path: string, progress: Omit<MangaReadingProgress, 'timestamp'>) {
      this.readingProgress = {
        ...this.readingProgress,
        [path]: {
          ...progress,
          timestamp: Date.now(),
        },
      }
      this.persistReadingProgress()
    },

    removeReadingProgress(path: string) {
      if (!(path in this.readingProgress)) {
        return
      }

      const nextProgress = { ...this.readingProgress }
      delete nextProgress[path]
      this.readingProgress = nextProgress
      this.persistReadingProgress()
    },

    removeMangaState(path: string) {
      this.removeShelfState(path)
      this.removeReadingProgress(path)
    },
  },
})
