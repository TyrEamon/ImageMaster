<template>
    <main>
        <EmptyState v-if="activeLibrary === '' && !loading" type="no-libraries" />
        <Loading v-if="loading" />
        <MangaGrid v-else-if="mangas.length > 0" />
        <EmptyState v-else-if="activeLibrary !== '' && !loading" type="no-mangas" />
    </main>

</template>

<script setup lang="ts">
import { storeToRefs } from "pinia";
import { onMounted, ref } from "vue";
import { GetActiveLibrary } from "../../../wailsjs/go/config/API";
import { Loading } from "../../components";
import { EmptyState, MangaGrid } from "./components";
import { MangaService } from "./services/mangaService";
import { useHomeStore } from "./stores/homeStore";

const homeStore = useHomeStore();
const { mangas } = storeToRefs(homeStore);

let loading = ref(false);

let activeLibrary = ref("");

async function getActiveLibrary() {
    const library = await GetActiveLibrary();
    activeLibrary.value = library;
}


onMounted(async () => {
    loading.value = true;
    getActiveLibrary();
    const mangaService = new MangaService();
    await mangaService.initialize();

    loading.value = false;
});

</script>

<style scoped></style>