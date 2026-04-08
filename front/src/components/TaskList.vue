<template>

    <div v-if="tasks.length === 0">
        <div class="text-center h-80 flex items-center justify-center text-neutral-100">
            <p> no task</p>
        </div>
    </div>
    <table v-else class="w-full text-neutral-100 scroll-auto">
        <thead>
            <tr class="border-b border-neutral-300">
                <th>名字</th>
                <th v-if="mode === 'active'">url</th>
                <th>状态</th>
                <th v-if="mode === 'active'">进度</th>
                <th v-if="mode === 'history'">完成时间</th>
                <th v-if="mode === 'history'">失败原因</th>
                <th v-if="mode === 'history'">耗时</th>
                <th v-if="mode === 'active'">操作</th>
            </tr>
        </thead>

        <tbody>
            <tr v-for="task in tasks" :key="task.id" class="border-b border-neutral-500/50">
                <td :title="task.name" class="max-w-48">{{ task.name }}</td>
                <td v-if="mode === 'active'" :title="task.url" class="max-w-64">{{ task.url }}</td>
                <td >
                    <div class="flex items-center justify-center gap-2">
                        <component :is="getStatusIcon(task.status)?.icon" :size="16"
                            :class="getStatusIcon(task.status)?.class" />
                        <span>{{ formatStatus(task.status) }}</span>
                    </div>
                </td>
                <td v-if="mode === 'active'">
                    <div class="border border-neutral-300 rounded-xl h-2 w-full">
                        <div class="bg-neutral-300 rounded-xl h-full transition-all duration-300"
                            :style="{ width: `${calculateProgressPercentage(task.progress?.current ?? 0, task.progress?.total ?? 0)}%` }">
                        </div>
                    </div>
                </td>
                <td v-if="mode === 'history'">{{ task.status === 'completed' ? formatTime(task.completeTime) : '-' }}</td>
                <td v-if="mode === 'history'" class="max-w-64">
                    <span v-if="task.status === 'failed' && task.error"
                          :title="task.error"
                          class="text-red-400 text-xs truncate">
                        {{ task.error }}
                    </span>
                    <span v-else class="text-neutral-500">-</span>
                </td>
                <td v-if="mode === 'history'">
                    <span v-if="task.startTime && task.completeTime">{{ calculateTimeDifference(task.startTime, task.completeTime) }}</span>
                    <span v-else class="text-neutral-500">-</span>
                </td>
                <td v-if="mode === 'active'" class="text-center">
                    <Button
                        size="sm"
                        :disabled="!canCancel(task.status)"
                        @click="$emit('cancel', task.id)"
                    >停止下载</Button>
                </td>
            </tr>
        </tbody>

    </table>

</template>

<script setup lang="ts">
import { Loader, ArrowBigDownDash, CircleCheck, CircleX, CircleOff } from 'lucide-vue-next';
import Button from './Button.vue';
import type { dto, task } from '../../wailsjs/go/models';

type DownloadTaskLike = (dto.DownloadTaskDTO | task.DownloadTask) & {
    progress?: { current?: number; total?: number } | null
}

const props = defineProps<{
    tasks: DownloadTaskLike[],
    mode?: 'active' | 'history'
}>();

defineEmits<{
    (e: 'cancel', taskId: string): void
}>();

function calculateProgressPercentage(current: number, total: number): number {
    if (total <= 0) return 0;
    return Math.round((current / total) * 100);
}

function calculateTimeDifference(startTime: string, endTime: string): string {
    const startTimeDate = new Date(startTime);
    const endTimeDate = new Date(endTime);
    const timeDifference = endTimeDate.getTime() - startTimeDate.getTime();
    const seconds = Math.max(0, Math.floor(timeDifference / 1000));
    return `${seconds}秒`;
}

function formatTime(timeStr: string): string {
    if (!timeStr) return '';
    const date = new Date(timeStr);
    return `${date.toLocaleDateString()} ${date.toLocaleTimeString()}`;
}

function getStatusIcon(status: string) {
    if (status === 'pending') {
        return { icon: Loader, class: 'animate-spin' };
    } else if (status === 'parsing') {
        return { icon: Loader, class: 'animate-spin' };
    } else if (status === 'downloading') {
        return { icon: ArrowBigDownDash, class: 'animate-bounce' };
    } else if (status === 'completed') {
        return { icon: CircleCheck, class: '' };
    } else if (status === 'failed') {
        return { icon: CircleX, class: '' };
    } else if (status === 'cancelled') {
        return { icon: CircleOff, class: '' };
    }
}

function canCancel(status: string): boolean {
    return status === 'pending' || status === 'parsing' || status === 'downloading';
}

function formatStatus(status: string): string {
    const statusMap: Record<string, string> = {
        'pending': '等待中',
        'parsing': '解析中',
        'downloading': '下载中',
        'completed': '已完成',
        'failed': '失败',
        'cancelled': '已取消'
    };
    return statusMap[status] || status;
}

const mode = props.mode ?? 'active';

</script>

<style scoped>
@reference "tailwindcss";

td,
th,
tr {
    @apply text-center text-xs px-2 py-4;
}

tr {
    @apply hover:bg-neutral-800;
}

td {
    @apply truncate
}
</style>