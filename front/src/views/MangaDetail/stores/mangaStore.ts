import { defineStore } from 'pinia';

export interface MangaState {
  mangaPath: string;
  mangaName: string;
  selectedImages: string[];
  loading: boolean;
  mangas: any[];
  currentMangaIndex: number;
}

const initialState: MangaState = {
  mangaPath: '',
  mangaName: '',
  selectedImages: [],
  loading: true,
  mangas: [],
  currentMangaIndex: -1,
};

export const useMangaStore = defineStore('mangaStore', {
  state: () => initialState,
  actions: {
    updateMangaStore(updates: Partial<MangaState>) {
      this.$patch(updates);
    },
    resetMangaStore() {
      this.$reset();
    }
  }
});

// 便捷的更新函数
// export const updateMangaStore = (updates: Partial<MangaState>) => {
//   useMangaStore().updateMangaStore(updates);
// };

// // 重置状态
// export const resetMangaStore = () => {
//   useMangaStore().resetMangaStore();
// };