<template>
    <button
        class="rounded-lg overflow-hidden xl:w-64 w-32 bg-neutral-400 cursor-pointer text-left hover:translate-y-[-4px] transition-transform duration-300 relative"
        @click="toMangaDetail">
        <div class="h-48 overflow-hidden">
            <img :src="mangaImageSrc" :alt="manga.name" class="w-full h-full object-cover" />
        </div>
        <div class="p-2 bg-neutral-600 flex flex-col gap-2 xl:h-24 h-20">
            <h3 class="xl:text-sm text-xs font-bold text-white line-clamp-2">
                {{ manga.name }}
            </h3>
            <h3 class="xl:text-xs text-[10px] text-neutral-400">
                total {{ manga.imagesCount }} images
            </h3>
        </div>
    </button>
</template>

<script setup lang="ts">
import { MangaService } from "../services";
import type { Manga } from "../stores/homeStore";
// import { push } from "svelte-spa-router";
import { useRouter } from "vue-router";
import { UrlEncode } from "../../../utils";

// export let manga: Manga;
const props = defineProps<{
    manga: Manga;
}>();

const mangaService = new MangaService();
const mangaImageSrc = mangaService.getMangaImage(props.manga.previewImg);
const router = useRouter();

function toMangaDetail() {
    router.push(`/manga/${UrlEncode(props.manga.path)}`);
}
</script>

<style scoped></style>