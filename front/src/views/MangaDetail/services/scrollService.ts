import type { Ref } from "vue";
import { ProgressService } from ".";
import type { useMangaStore } from "../stores";
import type { MangaService } from "./mangaService";

export class ScrollService {
    private scrollContainer: Ref<HTMLElement | null, HTMLElement | null>
    private mangaStore: ReturnType<typeof useMangaStore>;
    private saveTimeout: number | null = null;
    private smoothScroller: SmoothScroller;

    constructor(scrollContainer: Ref<HTMLElement | null, HTMLElement | null>, mangaStore: ReturnType<typeof useMangaStore>) {
        this.scrollContainer = scrollContainer;
        this.mangaStore = mangaStore;
        this.smoothScroller = new SmoothScroller(this.scrollContainer);
    }

    registerEvent = () => {
        window.addEventListener("keydown", this.handleKeyDown);
        window.addEventListener("keyup", this.handleKeyUp);
        return () => {
            window.removeEventListener("keydown", this.handleKeyDown);
            window.removeEventListener("keyup", this.handleKeyUp);
        }
    }

    handleKeyDown = (event: KeyboardEvent) => {
        if (event.key === "j") {
            this.smoothScroller.scrollDown();
        } else if (event.key === "k") {
            this.smoothScroller.scrollUp();
        }
    }

    handleKeyUp = (event: KeyboardEvent) => {
        if (event.key === "j" || event.key === "k") {
            this.smoothScroller.stopScroll();
        }
    }

    restoreScrollPosition() {
        console.log('restoreScrollPosition')

        const progress = ProgressService.getProgress(this.mangaStore.mangaPath);
        if (progress && progress.scrollPosition > 0) {
            // 延迟恢复滚动位置，确保图片已加载
            setTimeout(() => {
                if (this.scrollContainer) {
                    this.scrollContainer.value?.scrollBy({
                        top: progress.scrollPosition,
                        // behavior: "smooth"
                    })
                    console.log(
                        `已恢复到上次阅读位置：${progress.scrollPosition}px`,
                    );
                }
            }, 100);
        }
    }

    debounceSaveProgress() {
        if (this.saveTimeout) {
            clearTimeout(this.saveTimeout);
        }

        this.saveTimeout = setTimeout(() => {
            if (this.scrollContainer && this.mangaStore.mangaPath
                // && !this.isRestoringProgress
            ) {
                const scrollPosition = this.scrollContainer.value?.scrollTop;
                ProgressService.saveProgress(
                    this.mangaStore.mangaPath,
                    scrollPosition || 0,
                    this.mangaStore.selectedImages.length,
                );
            }
        }, 1000); // 1秒防抖
    }
}

export class SmoothScroller {
    private container: Ref<HTMLElement | null, HTMLElement | null>
    private targetScrollPos: number;
    private isScrolling: boolean;
    private scrollDirection: number; // 0: 无滚动, 1: 向下, -1: 向上
    private scrollAmount: number; // 每次滚动量
    private scrollDuration: number; // 每次滚动周期(ms)
    private frameDuration: number;

    constructor(container: Ref<HTMLElement | null, HTMLElement | null>, scrollAmount = 64, scrollDuration = 128) {
        this.container = container;
        this.targetScrollPos = this.container.value?.scrollTop || 0;
        this.isScrolling = false;
        this.scrollDirection = 0;
        this.scrollAmount = scrollAmount;
        this.scrollDuration = scrollDuration;
        this.frameDuration = 16;
    }

    // 缓动函数 (线性)
    easeLinear(t: number) {
        return t;
    }

    animateScroll = () => {
        if (this.scrollDirection === 0) {
            this.isScrolling = false
            return
        }
        const currentPos = this.container.value?.scrollTop || 0;
        const distance = this.targetScrollPos - currentPos;

        if (Math.abs(distance) < 0.1) {
            // 滚动完成
            this.isScrolling = false;
            if (this.container.value) {
                this.container.value.scrollTop = this.targetScrollPos;
            }

            // 如果有待处理的滚动，继续
            if (this.scrollDirection !== 0) {
                this.startScroll(this.scrollDirection);
            }
            return;
        }

        // 计算这一帧应该滚动的距离
        const frameCount = this.scrollDuration / this.frameDuration;
        const scrollThisFrame = this.scrollAmount / frameCount * this.scrollDirection;

        if (this.container.value) {
            this.container.value.scrollTop = currentPos + scrollThisFrame;
        }

        // 继续下一帧
        requestAnimationFrame(this.animateScroll);
    }

    /**
    * 开始滚动
    * @param {number} direction - 1: 向下, -1: 向上
    */
    startScroll = (direction: number) => {
        this.scrollDirection = direction;
        this.targetScrollPos = (this.container.value?.scrollTop || 0) + (this.scrollAmount * direction) || 0;

        if (!this.isScrolling) {
            this.isScrolling = true;
            requestAnimationFrame(this.animateScroll);
        }
    }

    stopScroll = () => {
        this.scrollDirection = 0;
    }

    scrollDown = () => this.startScroll(1)
    scrollUp = () => this.startScroll(-1)

}