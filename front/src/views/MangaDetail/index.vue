<template>
    <div class="flex flex-col h-full over">
        <Header :mangaService="mangaService" />

        <Loading v-if="loading" />
        <div v-else-if="selectedImages.length === 0" class="flex-grow flex flex-col items-center justify-center h-full">
            <p class="text-gray-100 mb-5">未找到图片</p>
        </div>
        <main v-else ref="scrollContainer" @scroll="scrollService.debounceSaveProgress"
            class="flex-grow overflow-y-auto p-5 flex flex-col items-center gap-5 flex-1">
            <div v-for="(image, i) in selectedImages" :key="i">
                <div class="max-w-[1200px] w-full">
                    <img :src="image" :alt="`Manga page ${i + 1}`" class="w-full h-auto block rounded" />
                </div>
            </div>
        </main>
    </div>
</template>

<script setup lang="ts">
import { storeToRefs } from "pinia";
import { onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { Loading } from "../../components";
import { Header } from "./components";
import { MangaService, ScrollService } from "./services";
import { useMangaStore } from "./stores";
import { UrlDecode } from "../../utils";

let scrollContainer = ref<HTMLElement | null>(null);

const mangaStore = useMangaStore();
const { loading, selectedImages } =
    storeToRefs(mangaStore);
const route = useRoute();
const scrollService = new ScrollService(scrollContainer, mangaStore);
const mangaService = new MangaService(scrollService);


onMounted(() => {
    init();
    scrollService.registerEvent();
});

watch(() => route.params.path, (newPath) => {
    if (newPath) {
        init();
    }
});

function init() {
    mangaService.loadManga(UrlDecode(route.params.path as string));
}

</script>

<style scoped></style>