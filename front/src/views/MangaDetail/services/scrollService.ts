import type { Ref } from 'vue'
import { ProgressService } from '.'
import type { useMangaStore } from '../stores'

export class ScrollService {
  private scrollContainer: Ref<HTMLElement | null, HTMLElement | null>
  private mangaStore: ReturnType<typeof useMangaStore>
  private saveTimeout: number | null = null
  private smoothScroller: SmoothScroller

  constructor(
    scrollContainer: Ref<HTMLElement | null, HTMLElement | null>,
    mangaStore: ReturnType<typeof useMangaStore>,
  ) {
    this.scrollContainer = scrollContainer
    this.mangaStore = mangaStore
    this.smoothScroller = new SmoothScroller(this.scrollContainer)
  }

  registerEvent = () => {
    window.addEventListener('keydown', this.handleKeyDown)
    window.addEventListener('keyup', this.handleKeyUp)
    return () => {
      window.removeEventListener('keydown', this.handleKeyDown)
      window.removeEventListener('keyup', this.handleKeyUp)
    }
  }

  handleKeyDown = (event: KeyboardEvent) => {
    if (event.key === 'j') {
      this.smoothScroller.scrollDown()
    } else if (event.key === 'k') {
      this.smoothScroller.scrollUp()
    }
  }

  handleKeyUp = (event: KeyboardEvent) => {
    if (event.key === 'j' || event.key === 'k') {
      this.smoothScroller.stopScroll()
    }
  }

  restoreScrollPosition() {
    const progress = ProgressService.getProgress(this.mangaStore.mangaPath)
    if (!progress || progress.scrollPosition <= 0) {
      return
    }

    setTimeout(() => {
      this.scrollContainer.value?.scrollTo({
        top: progress.scrollPosition,
      })
    }, 100)
  }

  debounceSaveProgress() {
    if (this.saveTimeout) {
      clearTimeout(this.saveTimeout)
    }

    this.saveTimeout = window.setTimeout(() => {
      const container = this.scrollContainer.value
      const mangaPath = this.mangaStore.mangaPath
      const totalImages = this.mangaStore.selectedImages.length

      if (!container || !mangaPath) {
        return
      }

      const scrollPosition = container.scrollTop
      const maxScrollTop = Math.max(0, container.scrollHeight - container.clientHeight)
      const rawProgress = maxScrollTop <= 0 ? (totalImages > 0 ? 1 : 0) : scrollPosition / maxScrollTop
      const progressPercent = Math.min(1, Math.max(0, rawProgress))
      const lastReadImage =
        totalImages <= 0
          ? 0
          : Math.min(totalImages, Math.max(1, Math.round(progressPercent * Math.max(totalImages - 1, 0)) + 1))
      const completed = totalImages > 0 && progressPercent >= 0.98

      ProgressService.saveProgress(
        mangaPath,
        scrollPosition,
        totalImages,
        progressPercent,
        lastReadImage,
        completed,
      )
    }, 1000)
  }
}

export class SmoothScroller {
  private container: Ref<HTMLElement | null, HTMLElement | null>
  private targetScrollPos: number
  private isScrolling: boolean
  private scrollDirection: number
  private scrollAmount: number
  private scrollDuration: number
  private frameDuration: number

  constructor(
    container: Ref<HTMLElement | null, HTMLElement | null>,
    scrollAmount = 64,
    scrollDuration = 128,
  ) {
    this.container = container
    this.targetScrollPos = this.container.value?.scrollTop || 0
    this.isScrolling = false
    this.scrollDirection = 0
    this.scrollAmount = scrollAmount
    this.scrollDuration = scrollDuration
    this.frameDuration = 16
  }

  easeLinear(t: number) {
    return t
  }

  animateScroll = () => {
    if (this.scrollDirection === 0) {
      this.isScrolling = false
      return
    }

    const currentPos = this.container.value?.scrollTop || 0
    const distance = this.targetScrollPos - currentPos

    if (Math.abs(distance) < 0.1) {
      this.isScrolling = false
      if (this.container.value) {
        this.container.value.scrollTop = this.targetScrollPos
      }

      if (this.scrollDirection !== 0) {
        this.startScroll(this.scrollDirection)
      }
      return
    }

    const frameCount = this.scrollDuration / this.frameDuration
    const scrollThisFrame = (this.scrollAmount / frameCount) * this.scrollDirection

    if (this.container.value) {
      this.container.value.scrollTop = currentPos + scrollThisFrame
    }

    requestAnimationFrame(this.animateScroll)
  }

  startScroll = (direction: number) => {
    this.scrollDirection = direction
    this.targetScrollPos = (this.container.value?.scrollTop || 0) + this.scrollAmount * direction

    if (!this.isScrolling) {
      this.isScrolling = true
      requestAnimationFrame(this.animateScroll)
    }
  }

  stopScroll = () => {
    this.scrollDirection = 0
  }

  scrollDown = () => this.startScroll(1)
  scrollUp = () => this.startScroll(-1)
}
