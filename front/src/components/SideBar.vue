<template>
    <aside class="border-r border-neutral-500/20 transition-[width] duration-200" :class="collapsed
        ? 'w-12'
        : 'w-48'">
        <!-- <div class="justify-between flex items-center">
        {#if !collapsed}
            <div class="p-4 text-neutral-300 text-sm">Image Master</div>
        {/if} -->
        <div class="border-b border-neutral-500/20">
            <button class="p-4 hover:bg-neutral-700/50 transition-colors duration-200 cursor-pointer"
                @click="collapsed = !collapsed">
                <AlignJustify class="w-4 h-4 text-neutral-300" />
            </button>
        </div>
        <!-- </div> -->

        <div class="border-b border-neutral-500/20">
            <template v-for="menu in menuList" :key="menu.path">
                <button @click="router.push(menu.path)" :class="`relative w-full hover:bg-neutral-700/50 transition-colors duration-200 cursor-pointer ${menu.active
                    ? 'bg-neutral-700/50'
                    : ''}`">
                    <template v-if="menu.active">
                        <div
                            class="absolute left-0 top-0 w-0.5 h-full rounded-full bg-blue-500 transition-all duration-500 animate-shrink">
                        </div>
                    </template>
                    <div class="p-4 flex items-center gap-2">
                        <component :is="menu.icon" class="w-4 h-4 text-neutral-300" />
                        <Transition >
                            <span v-if="!collapsed" class="text-neutral-300 text-xs ml-2">{{ menu.label }}</span>
                        </Transition>
                    </div>
                </button>
            </template>
        </div>
    </aside>
</template>

<script setup lang="ts">
import { AlignJustify, Archive, Download, Home, Settings } from "lucide-vue-next";
import { computed, ref } from "vue";
import { useRouter } from "vue-router";

const router = useRouter();

let collapsed = ref(false);

let menuList = computed(() => [
    {
        icon: Home,
        label: "Home",
        path: "/",
        active: router.currentRoute.value.path === "/",
    },
    {
        icon: Download,
        label: "Download",
        path: "/download",
        active: router.currentRoute.value.path === "/download",
    },
    {
        icon: Archive,
        label: "Extract",
        path: "/extract",
        active: router.currentRoute.value.path === "/extract",
    },
    {
        icon: Settings,
        label: "Setting",
        path: "/setting",
        active: router.currentRoute.value.path === "/setting",
    },

]);
</script>

<style scoped>
.v-enter-active,
.v-leave-active {
    transition: all 0.5s ease;
}

.v-enter-from,
.v-leave-to {
    position: absolute;
    opacity: 0;
    transform: translateX(-20px);
}
</style>
