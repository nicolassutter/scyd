<script setup lang="ts">
import * as z from "zod";
import type { FormSubmitEvent } from "@nuxt/ui";
import { useMutation } from "@tanstack/vue-query";
import { postApiV1Download } from "~/utils/client";

const schema = z.object({
  url: z.url(),
});

const route = useRoute();

type Schema = z.output<typeof schema>;

const state = reactive<Partial<Schema>>({
  url: undefined,
});

const downloadItems = defineModel<
  {
    id: string;
    url: string;
  }[]
>("items");

const startDownloadMutation = useMutation({
  mutationFn: async (url: string) => {
    const response = await postApiV1Download({
      body: { url },
    });

    return response.data;
  },
  onSuccess: (data, inputUrl) => {
    downloadItems.value?.unshift({
      id: data?.task_id ?? crypto.randomUUID(),
      url: inputUrl,
    });
    state.url = "";
  },
});

onMounted(() => {
  // when the app is installed as a PWA, users might share a URL to the app, which we will try to download
  const sharedUrl = z.url().safeParse(route.query.url).data;
  const sharedText = z.string().safeParse(route.query.text).data;

  // some apps put the URL inside the text field, so we try to extract it from there if needed
  const urlInText = sharedText
    ? (sharedText.match(
        /(https?:\/\/(?:www\.|(?!www))[a-zA-Z0-9][a-zA-Z0-9-]+[a-zA-Z0-9]\.[^\s]{2,}|www\.[a-zA-Z0-9][a-zA-Z0-9-]+[a-zA-Z0-9]\.[^\s]{2,}|https?:\/\/(?:www\.|(?!www))[a-zA-Z0-9]+\.[^\s]{2,}|www\.[a-zA-Z0-9]+\.[^\s]{2,})/g
      ) || [])[0]
    : null;

  const urlToUse = sharedUrl || z.url().safeParse(urlInText).data;

  if (urlToUse) {
    startDownloadMutation.mutate(urlToUse);
  }
});

async function handleSubmit(event: FormSubmitEvent<Schema>) {
  startDownloadMutation.mutate(event.data.url);
}
</script>

<template>
  <div class="max-w-4xl mx-auto text-center space-y-8">
    <h1 class="text-5xl md:text-7xl font-bold text-white leading-tight">
      Self-hosted
      <span
        class="block bg-gradient-to-r from-blue-400 to-cyan-400 bg-clip-text text-transparent"
        >Music Downloader
      </span>
    </h1>

    <p class="text-xl text-slate-400 max-w-2xl mx-auto leading-relaxed">
      Keep your favorite songs from a variety of platforms with just a single
      click. No account required, completely free and open-source.
    </p>

    <UForm :schema="schema" :state="state" @submit="handleSubmit">
      <div
        class="flex items-center gap-3 p-2 bg-white/5 backdrop-blur-sm border border-white/10 rounded-full shadow-2xl"
      >
        <UFormField name="url" class="w-full max-w-xl">
          <input
            v-model="state.url"
            aria-label="Url"
            type="text"
            placeholder="https://youtube.com/never-gonna-give-u-up"
            class="flex-1 w-full px-6 py-4 bg-transparent text-white placeholder:text-slate-500 outline-none text-lg"
            aria-required="true"
          />
        </UFormField>
        <button
          type="submit"
          class="shrink-0 cursor-pointer flex items-center justify-center w-14 h-14 bg-gradient-to-r from-blue-500 to-cyan-500 hover:from-blue-600 hover:to-cyan-600 rounded-full transition-all duration-200 hover:scale-105 active:scale-95 shadow-lg shadow-blue-500/25"
        >
          <UIcon name="i-lucide:download" size="20" />
        </button>
      </div>
    </UForm>
  </div>
</template>
