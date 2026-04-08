<template>
    <div class="relative" :class="attrs.class">
        <div v-if="help" class="absolute -top-5 text-xs text-neutral-500 mb-1">{{ help }}</div>
        <input ref="inputRef" v-bind="$attrs" :value="modelValue" @input="handleInput"
            class="w-full px-2 py-2 text-white text-sm outline-0 border-b-1 border-neutral-300 focus:border-b-blue-500 border-solid rounded-sm bg-neutral-700 focus:bg-neutral-800" />
    </div>
</template>

<script setup lang="ts">
import { defineProps, defineEmits, useAttrs, ref, onMounted } from "vue";

let inputRef = ref<HTMLInputElement>();

const attrs = useAttrs();
// 禁用属性继承，因为我们手动处理
defineOptions({
    inheritAttrs: false
});

const props = defineProps({
    modelValue: { type: [String, FileList, File], default: '' },
    help: { type: String, default: '' },
    autofocus: { type: Boolean, default: false }
});

onMounted(() => {
    if (props.autofocus) {
        inputRef.value?.focus();
    }
})

const emits = defineEmits(['update:modelValue']);

const handleInput = (e: Event) => {
    const target = e.target as HTMLInputElement;

    emits('update:modelValue', target.value);
}

defineExpose({
    // inputRef,
    focus: () => inputRef.value?.focus()
})
</script>

<style scoped></style>