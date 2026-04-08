/**
 * 漫画浏览进度管理服务
 * 使用localStorage保存和恢复漫画的阅读进度
 */

const PROGRESS_KEY = 'manga_reading_progress';

export interface MangaProgress {
  scrollPosition: number;
  timestamp: number;
  totalImages: number;
}

export interface ProgressData {
  [mangaPath: string]: MangaProgress;
}

export class ProgressService {
  constructor() {
    ProgressService.cleanupOldProgress();

  }
  /**
   * 保存漫画的浏览进度
   * @param mangaPath 漫画路径，作为唯一标识符
   * @param scrollPosition 滚动位置
   * @param totalImages 总图片数量
   */
  static saveProgress(mangaPath: string, scrollPosition: number, totalImages: number): void {
    try {
      const progressData = ProgressService.getAllProgress();

      progressData[mangaPath] = {
        scrollPosition,
        timestamp: Date.now(),
        totalImages
      };

      localStorage.setItem(PROGRESS_KEY, JSON.stringify(progressData));
    } catch (error) {
      console.error('保存阅读进度失败:', error);
    }
  }

  /**
   * 获取漫画的浏览进度
   * @param mangaPath 漫画路径
   * @returns 进度信息或null
   */
  static getProgress(mangaPath: string): MangaProgress | null {
    try {
      const progressData = ProgressService.getAllProgress();
      return progressData[mangaPath] || null;
    } catch (error) {
      console.error('读取阅读进度失败:', error);
      return null;
    }
  }

  /**
   * 删除指定漫画的进度记录
   * @param mangaPath 漫画路径
   */
  static removeProgress(mangaPath: string): void {
    try {
      const progressData = ProgressService.getAllProgress();
      delete progressData[mangaPath];
      localStorage.setItem(PROGRESS_KEY, JSON.stringify(progressData));
    } catch (error) {
      console.error('删除阅读进度失败:', error);
    }
  }

  /**
   * 清理过期的进度记录（超过30天）
   */
  static cleanupOldProgress(): void {
    try {
      const progressData = ProgressService.getAllProgress();
      const thirtyDaysAgo = Date.now() - (30 * 24 * 60 * 60 * 1000);

      Object.keys(progressData).forEach(mangaPath => {
        if (progressData[mangaPath].timestamp < thirtyDaysAgo) {
          delete progressData[mangaPath];
        }
      });

      localStorage.setItem(PROGRESS_KEY, JSON.stringify(progressData));
    } catch (error) {
      console.error('清理过期进度失败:', error);
    }
  }

  /**
   * 获取所有进度数据
   * @returns 所有进度数据
   */
  private static getAllProgress(): ProgressData {
    try {
      const data = localStorage.getItem(PROGRESS_KEY);
      return data ? JSON.parse(data) : {};
    } catch (error) {
      console.error('解析进度数据失败:', error);
      return {};
    }
  }

  /**
   * 检查是否有保存的进度
   * @param mangaPath 漫画路径
   * @returns 是否存在进度记录
   */
  static hasProgress(mangaPath: string): boolean {
    const progress = ProgressService.getProgress(mangaPath);
    return progress !== null && progress.scrollPosition > 0;
  }
} 