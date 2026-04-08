import { defineStore } from 'pinia';

export interface Manga {
  name: string;
  path: string;
  previewImg: string;
  imagesCount: number;
}

export interface Library {
  // 根据实际Library结构定义
  [key: string]: any;
}

// 状态管理
export const useHomeStore = defineStore('homeStore', {
  state: () => ({
    mangas: [] as Manga[],
    libraries: [] as string[],
    loading: true,
    mangaImages: new Map<string, string>(),
    showScrollTop: false,
    scrollY: 0
  }),
  actions: {
    setMangas(mangas: Manga[]) {
      this.mangas = mangas;
    }
  }
});