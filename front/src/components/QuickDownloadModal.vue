<template>
    <Teleport to="body">
        <Transition name="fade">
            <div v-if="modelValue"
                class="fixed inset-0 h-full w-full bg-neutral-900/30 backdrop-blur-xs flex items-center justify-center"
                @click="closeModal">
                <div @keydown="handleKeydown" @click.stop
                    class="bg-neutral-800 py-8 px-4 rounded-sm w-full max-w-md flex flex-col gap-4">
                    <h2 class="text-lg font-bold text-white">{{ title }}</h2>

                    <div class="flex items-center gap-2 mt-4">
                        <Input ref="inputRef" v-model="url" class="flex-1" help="please input the target manga url" />
                    </div>
                    <div class="flex justify-end">
                        <div class="flex gap-2">
                            <Button @click="closeModal">
                                <div class="flex items-center gap-2">
                                    <X :size="16" class="text-white" />
                                    <span>Close</span>
                                </div>
                            </Button>
                            <Button @click="handleDownload">
                                <div class="flex items-center gap-2">
                                    <Download :size="16" class="text-white" />
                                    <span>Download</span>
                                </div>
                            </Button>
                        </div>

                    </div>

                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<script setup lang="ts">
import { createDownloadHandler } from "@/views/Download/services";
import { defineProps, defineEmits, ref, onMounted, watch, nextTick } from "vue";
import { Button, Input } from ".";
import { Download, X } from "lucide-vue-next";
import { toast } from "vue-sonner";


let url = ref('');
let inputRef = ref<InstanceType<typeof Input>>();

const props = defineProps({
    modelValue: { type: Boolean, default: false }, // v-model 控制显示/隐藏
    title: { type: String, default: '快速下载' },
});

watch(() => props.modelValue, (newVal) => {
    if (newVal) {
        nextTick(() => {
            inputRef.value?.focus();
        })
    }
})

const emits = defineEmits(["update:modelValue"]);

const closeModal = () => {
    url.value = '';
    emits("update:modelValue", false); // 关闭 Modal
};

// 创建下载处理器
const downloadHandler = createDownloadHandler({
    onSuccess: (taskId, downloadUrl) => {
        closeModal();
    },
    onError: (errorMsg) => {
        toast.error(errorMsg);
    },
});

// 处理下载
async function handleDownload() {
    if (!url.value.trim()) {
        toast.error('请输入网址');
        return;
    }

    await downloadHandler(url.value.trim());
}

// 处理键盘事件
function handleKeydown(event: any) {
    if (event.key === 'Enter') {
        handleDownload();
    } else if (event.key === 'Escape') {
        closeModal();
    }
}

</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}
</style>