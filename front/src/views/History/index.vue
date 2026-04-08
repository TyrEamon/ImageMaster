<template>
    <div class="p-8">
        <div class="flex items-center justify-between gap-2">
            <Button @click="router.back()">
                <div class="flex items-center text-white gap-2">
                    <ArrowLeft :size="16" class="text-white" />
                    <span>返回</span>
                </div>
            </Button>
            <Button @click="clearHistory">
                <div class="flex items-center text-white gap-2">
                    <Trash :size="16" class="text-white" />
                    <span>清空记录</span>
                </div>
            </Button>
        </div>
        <TaskList :tasks="historyTasks" mode="history" />
    </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { ClearHistory, GetHistoryTasks } from '../../../wailsjs/go/api/CrawlerAPI';
import type { dto } from '../../../wailsjs/go/models';
import { Button, TaskList } from '@/components';
import { useRouter } from 'vue-router';
import { ArrowLeft, Trash } from 'lucide-vue-next';

const router = useRouter();

let historyTasks = ref<dto.DownloadTaskDTO[]>([])

async function clearHistory() {
    try {
        if (!confirm(`确定要清空任务吗？`)) {
            return false;
        }
        await ClearHistory();
        await loadData()
    } catch (err) {
        console.error('清除历史出错:', err);
        throw err;
    }
}

async function loadData() {
    const history = await GetHistoryTasks();
    historyTasks.value = history
    console.log(historyTasks)
}

onMounted(() => {
    loadData()
})
</script>

<style scoped></style>