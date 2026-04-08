import { useRouter } from 'vue-router';
import { toast } from 'vue-sonner';
import {
  DeleteManga,
  GetAllMangas,
  GetImageDataUrl,
  GetMangaImages
} from '../../../../wailsjs/go/library/API';
import { useMangaStore } from '../stores';
import { ProgressService } from './progressService';
import type { ScrollService } from './scrollService';

export class MangaService {
  private router: ReturnType<typeof useRouter>;
  private mangaStore: ReturnType<typeof useMangaStore>;
  private scrollService: ScrollService;

  constructor(scrollService: ScrollService) {
    this.scrollService = scrollService;
    this.mangaStore = useMangaStore();
    this.router = useRouter();
  }

  async loadManga(path: string, callback?: () => void) {
    console.log('loadManga', path)
    try {
      this.mangaStore.loading = true;

      // 解码路径参数
      const mangaPath = decodeURIComponent(path);

      // 获取所有漫画以支持导航功能
      const mangas = await GetAllMangas();
      const currentMangaIndex = mangas.findIndex(m => m.path === mangaPath);

      let mangaName: string;
      if (currentMangaIndex >= 0) {
        mangaName = mangas[currentMangaIndex].name;
      } else {
        mangaName = mangaPath.split('/').pop() || '';
      }

      this.mangaStore.updateMangaStore({
        mangaPath,
        mangaName,
        mangas,
        currentMangaIndex,
        selectedImages: []
      });

      // 获取所有图片
      await this.loadImages(mangaPath);

    } catch (error) {
      console.error('加载漫画失败:', error);
    } finally {
      this.mangaStore.loading = false;
      this.scrollService.restoreScrollPosition();
    }
  }

  async loadImages(mangaPath: string) {
    try {
      // 获取所有图片路径
      const imagePaths = await GetMangaImages(mangaPath);

      // 并行加载所有图片，保持顺序
      const imagePromises = imagePaths.map(async (imagePath) => {
        try {
          return await GetImageDataUrl(imagePath);
        } catch (error) {
          console.error(`加载图片失败: ${imagePath}`, error);
          return null;
        }
      });

      // 等待所有图片加载完成
      const loadedImages = await Promise.all(imagePromises);

      // 过滤掉加载失败的图片（null值）
      const selectedImages = loadedImages.filter(img => img !== null);
      this.mangaStore.updateMangaStore({ selectedImages });
    } catch (error) {
      console.error("获取图片路径失败:", error);
    }
  }

  backToHome() {
    this.router.push('/');
  }


  navigateToNextManga() {
    if (this.mangaStore.currentMangaIndex < this.mangaStore.mangas.length - 1) {
      const nextManga = this.mangaStore.mangas[this.mangaStore.currentMangaIndex + 1];
      const encodedPath = encodeURIComponent(nextManga.path);

      // 使用替代路由方案处理相同路径不同参数的导航
      const currentLocation = window.location.href;
      if (currentLocation.includes('/manga/')) {
        // 如果当前已经在漫画页面，采用直接加载新数据的方式
        this.loadManga(nextManga.path);

        // 更新 URL 但不触发导航事件
        window.history.pushState(null, '', `/#/manga/${encodedPath}`);
      } else {
        // 否则正常导航
        this.router.push(`/manga/${encodedPath}`);
      }
    }
  }

  navigateToPrevManga() {
    if (this.mangaStore.currentMangaIndex > 0) {
      const prevManga = this.mangaStore.mangas[this.mangaStore.currentMangaIndex - 1];
      const encodedPath = encodeURIComponent(prevManga.path);

      // 使用替代路由方案处理相同路径不同参数的导航
      const currentLocation = window.location.href;
      if (currentLocation.includes('/manga/')) {
        // 如果当前已经在漫画页面，采用直接加载新数据的方式
        this.loadManga(prevManga.path);

        // 更新 URL 但不触发导航事件
        window.history.pushState(null, '', `/#/manga/${encodedPath}`);
      } else {
        // 否则正常导航
        this.router.push(`/manga/${encodedPath}`);
      }
    }
  }

  async deleteAndViewNextManga() {
    if (this.mangaStore.currentMangaIndex >= 0 && confirm(`确定要删除 "${this.mangaStore.mangaName}" 并查看下一部漫画吗？`)) {
      this.mangaStore.loading = true;

      try {
        // 记录下一个漫画的位置，因为删除后数组会变化
        const hasNextManga = this.mangaStore.currentMangaIndex < this.mangaStore.mangas.length - 1;
        const nextMangaPath = hasNextManga ? this.mangaStore.mangas[this.mangaStore.currentMangaIndex + 1].path : null;

        // 执行删除操作
        const success = await DeleteManga(this.mangaStore.mangaPath);

        if (success) {
          // 删除成功后，清除该漫画的阅读进度
          ProgressService.removeProgress(this.mangaStore.mangaPath);
          toast.success('删除成功');
          if (nextMangaPath) {
            // 重要：在导航前设置 loading 为 false，防止新页面保持加载状态
            this.mangaStore.loading = false;

            // 使用替代路由方案处理相同路径不同参数的导航
            const encodedPath = encodeURIComponent(nextMangaPath);
            const currentLocation = window.location.href;
            if (currentLocation.includes('/manga/')) {
              // 如果当前已经在漫画页面，采用直接加载新数据的方式
              this.loadManga(nextMangaPath);

              // 更新 URL 但不触发导航事件
              window.history.pushState(null, '', `/#/manga/${encodedPath}`);
            } else {
              // 否则正常导航
              this.router.push(`/manga/${encodedPath}`);
            }
          } else {
            // 如果没有下一部漫画，返回首页
            this.router.push('/');
          }
        } else {
          alert('删除失败!');
          this.mangaStore.loading = false;
        }
      } catch (error) {
        console.error('删除漫画失败:', error);
        this.mangaStore.loading = false;
      }
    }
  }
}