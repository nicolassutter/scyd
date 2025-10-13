<script setup lang="ts">
import z from "zod";
import type { Download } from "~/utils/client";

const props = defineProps<Download>();

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

const { websocketEmitter } = useDownloads();

const websocketEventName = `download-${props.id}` as `download-${number}`;

// Listen for websocket events specific to this download item
websocketEmitter.on(websocketEventName, (incomingMessage) => {
  if (incomingMessage.event === "progress") {
    logs.value.push(incomingMessage.data);
  } else if (incomingMessage.event === "success") {
    logs.value.push("✅ Download completed successfully.");
    close();
  }
});

onBeforeUnmount(() => {
  websocketEmitter.off(websocketEventName);
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

const downloadState = computed(
  () =>
    z.enum(["success", "pending", "progress", "error"]).safeParse(props.state)
      .data
);
</script>

<template>
  <UCard variant="subtle">
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

        <div>
          <h3 class="text-lg font-semibold text-white mb-1 truncate">
            {{ props.url }}
          </h3>

          <p v-if="downloadState">State: {{ downloadState }}</p>
        </div>
      </div>
    </template>

    <div
      v-if="logs.length"
      ref="logsScroller"
      class="bg-slate-900/50 rounded-lg p-4 font-mono text-sm max-h-48 overflow-y-auto"
    >
      <div
        v-for="(log, index) in logs"
        :key="index"
        class="text-slate-300 mb-1 leading-relaxed"
      >
        {{ log }}
      </div>

      <template v-if="downloadState === 'error' && props.error_message">
        last error message: {{ props.error_message }}
      </template>
    </div>
  </UCard>
</template>
