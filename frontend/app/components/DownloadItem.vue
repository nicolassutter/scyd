<script setup lang="ts">
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
  async () => {
    // Auto-scroll to the bottom when new logs are added
    await nextTick(); // Wait for DOM to update

    logsScroller.value?.scrollTo({
      top: logsScroller.value.scrollHeight,
      behavior: "smooth",
    });
  },
  { deep: true }
);
</script>

<template>
  <UCard
    variant="subtle"
    class="bg-slate-800/50 backdrop-blur-sm border border-slate-700/50 rounded-2xl hover:border-slate-600/50 transition-colors"
  >
    <template #header>
      <div class="flex items-start gap-4">
        <div
          v-if="thumbnailUrl"
          class="w-16 h-16 rounded-lg overflow-hidden flex-shrink-0 bg-slate-700"
        >
          <img :src="thumbnailUrl" alt="" class="w-full h-full object-cover" />
        </div>
        <div
          v-else
          class="w-16 h-16 rounded-lg flex-shrink-0 bg-gradient-to-br from-blue-500/20 to-purple-500/20 flex items-center justify-center"
        >
          <UIcon name="i-lucide:download" class="w-8 h-8 text-blue-400" />
        </div>

        <h3 class="text-lg font-semibold text-white mb-1 truncate">
          {{ props.url }}
        </h3>
      </div>
    </template>

    <div
      ref="logsScroller"
      class="bg-slate-900/50 rounded-lg p-4 font-mono text-sm max-h-48 overflow-y-auto"
    >
      <p v-if="!logs.length" class="text-slate-500 italic">
        No logs to display.
      </p>

      <div
        v-for="(log, index) in logs"
        :key="index"
        class="text-slate-300 mb-1 leading-relaxed"
      >
        {{ log }}
      </div>
    </div>
  </UCard>
</template>
