<script setup lang="ts">
definePageMeta({
  middleware: ["authenticated"],
});

const downloadItems = ref<
  {
    id: string;
    url: string;
  }[]
>([]);
</script>

<template>
  <div class="min-h-screen flex flex-col items-center p-6 pt-36">
    <Hero v-model:items="downloadItems" />
    <DownloadSection :downloads="downloadItems" />
  </div>
</template>
