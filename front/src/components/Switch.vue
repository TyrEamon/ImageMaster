<template>
    <button @click="toggle" class="flex items-center gap-2 cursor-pointer">
        <div :class="modelValue ? 'bg-neutral-800' : 'border-neutral-400'" class="rounded-xl w-10 h-5 border-1 transition-all duration-300">
            <div class="rounded-full w-4 h-4 transition-all duration-300 mt-[1px]" :class="modelValue ? 'ml-6 bg-neutral-50' : 'ml-[1px] bg-neutral-300'">
            </div>
        </div>
        <span class="text-xs text-neutral-200">{{ showLabel }}</span>
    </button>
</template>

<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
    modelValue: boolean
    activeLabel: string
    inactiveLabel: string
}>();

const emit = defineEmits<{
    'update:modelValue': [value: boolean]
}>();

const showLabel = computed(() => {
    return props.modelValue ? props.activeLabel : props.inactiveLabel;
});

function toggle() {
    emit('update:modelValue', !props.modelValue);
}
</script>

<style scoped></style>