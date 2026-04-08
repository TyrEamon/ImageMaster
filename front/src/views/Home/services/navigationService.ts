import { useRouter } from 'vue-router';
import type { Manga } from '../stores/homeStore';


export class NavigationService {
  static router = useRouter();
  /**
   * 跳转到漫画查看页面
   */
  static viewManga(manga: Manga): void {
    // 将路径编码后传递给路由
    const encodedPath = encodeURIComponent(manga.path);
    this.router.push(`/manga/${encodedPath}`);
  }
}