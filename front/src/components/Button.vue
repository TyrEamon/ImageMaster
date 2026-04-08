<template>
    <button :class="style"
        class="text-xs text-white px-4 py-2 rounded-md border  cursor-pointer border-neutral-700/50 transition-all duration-300 active:outline-1">
        <slot></slot>
    </button>
</template>

<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps({
    type: {
        type: String,
        default: 'default'
    },
    size: {
        type: String,
        default: 'md'
    }
})

const style = computed(() => {
    return {
        'bg-sky-800 hover:bg-sky-700': props.type === 'primary',
        'bg-transparent hover:bg-neutral-700/50': props.type === 'default',
        'bg-red-800 hover:bg-red-600': props.type === 'danger'
    }
})
</script>

<style scoped></style>