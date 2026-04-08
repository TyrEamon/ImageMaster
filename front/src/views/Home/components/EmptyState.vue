<template>
  <div v-if="type === 'no-libraries'" class="text-center p-10 text-white rounded-lg mt-5">
    <h2>欢迎使用漫画查看器</h2>
    <p>请前往"应用设置"页面添加一个漫画库</p>
  </div>
  <div v-else-if="type === 'no-mangas'" class="text-center p-10 text-white rounded-lg mt-5">
    <p>
      未找到漫画。请确保您添加的漫画库中包含子文件夹，且这些文件夹中含有图片文件。
    </p>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  type: 'no-libraries' | 'no-mangas';
}>();
</script>

<style scoped>

</style>