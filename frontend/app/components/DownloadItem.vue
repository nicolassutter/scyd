<script setup lang="ts">
// import { useEventSource } from "@vueuse/core";
import z from "zod";
import { client } from "~/utils/client/client.gen";

const props = defineProps<{
  id: string;
  url: string;
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

const logsScroller = useTemplateRef<HTMLElement>("logsScroller");

const logs = ref<string[]>([]);

const evtSource = new EventSource(
  client.buildUrl({
    baseUrl: client.getConfig().baseUrl,
    url: `/api/v1/download/stream/${props.id}`,
  })
);

evtSource.addEventListener("new_line", (event) => {
  try {
    const data = z
      .object({
        line: z.string(),
      })
      .parse(JSON.parse(event.data));

    logs.value.push(data.line);
  } catch (error) {
    console.error("Failed to parse new_line event data:", event.data, error);
  }
});

evtSource.addEventListener("download_success", () => {
  evtSource.close();
  logs.value.push("✅ Download completed successfully.");
});

evtSource.onerror = (_error) => {
  evtSource.close();
  console.log("EventSource connection closed due to error.");
};

evtSource.onopen = () => {
  console.log("EventSource connection opened");
};

onBeforeUnmount(() => {
  evtSource.close();
});

watch(
  logs,
  () => {
    // Auto-scroll to the bottom when new logs are added
    if (logsScroller.value) {
      logsScroller.value.scrollTop = logsScroller.value.scrollHeight;
    }
  },
  { deep: true }
);
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

    <ol ref="logsScroller" class="max-h-60 overflow-y-auto">
      <li v-for="log in logs" :key="log" class="text-sm font-mono">
        {{ log }}
      </li>
    </ol>
  </UCard>
</template>
