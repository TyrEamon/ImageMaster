<template>
    <div class="p-8 h-screen flex flex-col">
        <div class="flex gap-4 mt-2">
            <Input autofocus v-model="url" class="flex-1" help="please input the target manga url" @keydown="handleKeydown" />
            <Button @click="handleDownload">
                <div class="flex items-center gap-2">
                    <Download :size="16" class="text-white" />
                    <span>下载</span>
                </div>
            </Button>
        </div>

        <div class="h-1 border-b border-neutral-300/50 w-full my-8"></div>

        <div class="flex items-center justify-end gap-2">
            <Button @click="router.push('/history')">
                <div class="flex items-center gap-2">
                    <History :size="16" class="text-white" />
                    <span>历史记录</span>
                </div>
            </Button>
        </div>
        <div class="flex-1 overflow-auto">
            <TaskList class="mt-2" :tasks="activeTasks" mode="active" @cancel="onCancelTask" />
        </div>
    </div>

</template>

<script setup lang="ts">
import { Button, Input, TaskList } from '@/components';
import { Download, History } from 'lucide-vue-next';
import { storeToRefs } from 'pinia';
import { onMounted, onUnmounted, ref } from 'vue';
import { toast } from 'vue-sonner';
import { createDownloadHandler } from './services';
import { useDownloadStore } from './stores';
import { useRouter } from 'vue-router';

const router = useRouter();

const downloadStore = useDownloadStore();

let url = ref('');

const { activeTasks } = storeToRefs(downloadStore);

function handleKeydown(event: any) {
    if (event.key === 'Enter') {
        handleDownload();
    }
}

onMounted(async () => {
    await downloadStore.initializeStore();
    console.log("组件已挂载，轮询已开始");
});

onUnmounted(() => {
    downloadStore.stopPolling();
    console.log("组件已销毁，轮询已停止");
});


// 处理下载
async function handleDownload() {
    if (!url.value.trim()) {
        toast.error('请输入网址');
        return;
    }

    await downloadHandler(url.value.trim());
    url.value = '';
}

// 创建下载处理器
const downloadHandler = createDownloadHandler({
    onError: (errorMsg) => {
        toast.error(errorMsg);
    },
});

async function onCancelTask(taskId: string) {
    await downloadStore.cancelTask(taskId);
}


</script>

<style scoped></style>