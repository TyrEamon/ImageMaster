<template>
    <header class="py-2 px-4 bg-transparent flex justify-between items-center gap-4">
        <div class="flex items-center gap-2 shrink-0">
            <Button @click="router.push('/')">
                <div class="flex items-center gap-2">
                    <ArrowLeft :size="16" class="text-white" />
                    <span>返回</span>
                </div>

            </Button>

        </div>
        <div class="text-white text-sm font-bold flex-1 truncate">
            {{ mangaName }}
        </div>

        <div class="flex items-center gap-2 shrink-0">
            <Button @click="showNavigation = !showNavigation">
                <div class="flex items-center gap-2">
                    <EyeClosed v-if="!showNavigation" :size="16" class="text-white" />
                    <Eye :size="16" class="text-white" v-else />
                    <span>显示导航</span>
                </div>
            </Button>

            <Button @click="showQuickDownloadModal = true">
                <div class="flex items-center gap-2">
                    <Download :size="16" class="text-white" />
                    <span>快速下载</span>
                </div>
            </Button>

            <Button @click="() => mangaService.deleteAndViewNextManga()">
                <div class="flex items-center gap-2">
                    <Trash :size="16" class="text-white" />
                    <span>删除并看下一部</span>
                </div>
            </Button>
        </div>
    </header>
    <QuickDownloadModal v-model="showQuickDownloadModal" />
    <Transition>
        <div v-if="showNavigation" class="fixed bottom-8 right-8">
            <div class="flex flex-col gap-2">
                <button class="bg-neutral-900/70 p-3 rounded-full cursor-pointer hover:bg-neutral-900/90"
                    @click="mangaService.navigateToNextManga()">
                    <ChevronRight :size="20" class="text-white" />
                </button>
                <button class="bg-neutral-900/70 p-3 rounded-full cursor-pointer hover:bg-neutral-900/90"
                    @click="mangaService.navigateToPrevManga()">
                    <ChevronLeft :size="20" class="text-white" />
                </button>
            </div>
        </div>
    </Transition>
</template>

<script setup lang="ts">
import { ArrowLeft, ChevronLeft, ChevronRight, Download, Eye, EyeClosed, Trash } from 'lucide-vue-next';
import { useRouter } from 'vue-router';
import { useMangaStore } from '../stores';
import { storeToRefs } from 'pinia';
import { Button, QuickDownloadModal } from '../../../components';
import { onMounted, onUnmounted, ref, type PropType } from 'vue';
import { MangaService } from '../services';

const router = useRouter();
const mangaStore = useMangaStore();
const { mangaName } = storeToRefs(mangaStore);

defineProps({
    mangaService: {
        type: Object as PropType<MangaService>,
        required: true
    }
})

let showNavigation = ref(false);
let showQuickDownloadModal = ref(false);

// 如果按下 ctrl + k ，则显示 QuickDownloadModal
function handleKeydown(event: KeyboardEvent) {
    if (event.ctrlKey && event.key === 'd') {
        showQuickDownloadModal.value = true;
    }
}

onMounted(() => {
    window.addEventListener('keydown', handleKeydown);
})

onUnmounted(() => {
    window.removeEventListener('keydown', handleKeydown);
})

</script>

<style scoped>
.v-enter-active,
.v-leave-active {
    transition: all 0.5s ease;
}

.v-enter-from,
.v-leave-to {
    opacity: 0;
    transform: translateY(20px);
}
</style>