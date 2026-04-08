<script setup lang="ts">
// import routes from './routes';
import { Toaster, toast } from 'vue-sonner';
import { onMounted } from 'vue';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { SideBar } from './components';
import 'vue-sonner/style.css'

onMounted(() => {
  // 监听下载完成事件
  EventsOn('download:completed', (data: any) => {
    toast.success(`下载完成！`, {
      description: `已成功下载 ${data.name}`,
      duration: 5000,
    });
  });

  // 监听下载失败事件
  EventsOn('download:failed', (data: any) => {
    toast.error(`下载失败`, {
      description: `下载 ${data.name} 失败：${data.message || '下载过程中发生错误'}`,
      duration: 5000,
    });
  });

  // 监听下载取消事件
  EventsOn('download:cancelled', (data: any) => {
    toast.warning(`下载已取消`, {
      description: `已取消下载任务：${data.name}`,
      duration: 3000,
    });
  });
});

</script>

<template>

  <div class="flex h-screen bg-neutral-900">
    <SideBar />
    <main class="flex-1 h-screen overflow-auto">
      <RouterView />
    </main>
  </div>
  <Toaster />
</template>

<style scoped></style>
