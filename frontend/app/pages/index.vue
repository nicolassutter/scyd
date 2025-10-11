<script setup lang="ts">
const downloadItems = ref<
  {
    id: string;
    url: string;
  }[]
>([]);
</script>

<template>
  <div
    class="min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-slate-900 flex flex-col items-center p-6 pt-36"
  >
    <Hero v-model:items="downloadItems" />
    <DownloadSection :downloads="downloadItems" />
  </div>
</template>
