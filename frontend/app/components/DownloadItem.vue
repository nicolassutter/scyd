<script setup lang="ts">
const props = defineProps<{
  url: string;
  loading: boolean;
  error?: string;
  downloadedFiles: string[];
}>();

const thumbnailUrl = computed(() => {
  const isYoutube =
    props.url.includes("youtube.com") || props.url.includes("youtu.be");

  if (isYoutube) {
    const videoIdMatch = props.url.match(
      /(?:youtube\.com\/(?:[^/\n\s]+\/\S+\/|(?:v|e(?:mbed)?)\/|\S*?[?&]v=)|youtu\.be\/)([a-zA-Z0-9_-]{11})/
    );
    const videoId = videoIdMatch ? videoIdMatch[1] : null;
    return videoId
      ? `https://img.youtube.com/vi/${videoId}/hqdefault.jpg`
      : undefined;
  }

  return undefined;
});
</script>

<template>
  <UCard variant="subtle">
    <template #header>
      <h3 class="">{{ url }}</h3>
    </template>

    <div v-if="thumbnailUrl" class="mb-4">
      <img
        :src="thumbnailUrl"
        alt=""
        class="max-w-lg aspect-video rounded-xl object-cover"
      />
    </div>

    <p v-if="loading" class="flex gap-2 items-center">
      Downloading...
      <UIcon name="i-lucide:loader-circle" class="animate-spin" />
    </p>

    <div v-else-if="error" class="text-red-600">{{ error }}</div>

    <div v-else-if="!loading && downloadedFiles.length === 0">
      No files were downloaded.
    </div>

    <template v-else-if="!loading && downloadedFiles.length > 0">
      <p class="font-bold">Downloaded files:</p>
      <ul>
        <li v-for="file in downloadedFiles" :key="file">{{ file }}</li>
      </ul>
    </template>
  </UCard>
</template>
