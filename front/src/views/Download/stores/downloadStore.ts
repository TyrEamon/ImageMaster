import { defineStore } from 'pinia';
import { ref } from 'vue';
import {
  CancelCrawl,
  GetActiveTasks
} from '../../../../wailsjs/go/api/CrawlerAPI';
import type { task } from '../../../../wailsjs/go/models';

// 轮询相关
const POLL_INTERVAL = 1000;
let pollTimer: ReturnType<typeof setInterval> | null = null;

export type DownloadStore = ReturnType<typeof useDownloadStore>;

export const useDownloadStore = defineStore('downloadStore', {
  state: () => ({
    activeTasks: ref<task.DownloadTask[]>([]),
    historyTasks: ref<task.DownloadTask[]>([]),
    loading: false as boolean
  }),
  getters: {
    activeTasksCount: (state) => state.activeTasks.length
  },
  actions: {
    async initializeStore() {
      try {
        await this.pollTasks();
        this.startPolling();
      } catch (err) {
        console.error('初始化store失败:', err);
      }
    },
    async pollTasks() {
      try {
        const active = await GetActiveTasks();
        this.activeTasks = active;
      } catch (err) {
        console.error('轮询任务状态出错:', err);
      }
    },
    startPolling() {
      this.stopPolling();
      pollTimer = setInterval(this.pollTasks, POLL_INTERVAL);
    },
    stopPolling() {
      if (pollTimer) {
        clearInterval(pollTimer);
        pollTimer = null;
      }
    },
    async cancelTask(taskId: string) {
      try {
        await CancelCrawl(taskId);
        await this.pollTasks();
      } catch (err) {
        console.error('取消任务出错:', err);
        throw err;
      }
    },
  }
});
