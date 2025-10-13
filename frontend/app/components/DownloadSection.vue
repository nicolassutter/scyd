<script setup lang="ts">
defineProps<{
  downloads: {
    id: string;
    url: string;
  }[];
}>();
</script>

<template>
  <UContainer>
    <section class="w-full py-16">
      <h2 class="text-3xl font-bold text-white mb-8 text-balance">Downloads</h2>

      <div v-if="downloads.length" class="grid md:grid-cols-2 gap-6">
        <template v-for="download in downloads" :key="download.id">
          <DownloadItem :id="download.id" :url="download.url" />
        </template>
      </div>

      <div v-else class="text-slate-400 italic text-lg">
        <p>Your downloads will display here.</p>
      </div>
    </section>
  </UContainer>
</template>
