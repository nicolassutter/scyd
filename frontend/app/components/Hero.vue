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
  <UPageHero
    title="Scyd - Self-hosted music downloader"
    description="Keep your favorite songs from a variety of platforms with just a single
      click. No account required, completely free and open-source."
    :links="[]"
  />

  <UForm
    class="w-full max-w-xl mx-auto"
    :schema="schema"
    :state="state"
    @submit="handleSubmit"
  >
    <div
      class="flex items-center gap-3 py-2 px-4 bg-white/5 backdrop-blur-sm border border-white/10 rounded-full shadow-2xl"
    >
      <UFormField name="url" class="w-full">
        <input
          v-model="state.url"
          aria-label="Url"
          type="text"
          placeholder="https://youtube.com/never-gonna-give-u-up"
          class="flex-1 w-full py-4 bg-transparent text-white placeholder:text-slate-500 outline-none text-lg"
          aria-required="true"
        />
      </UFormField>
      <UButton
        type="submit"
        icon="i-lucide:download"
        size="xl"
        class="rounded-full"
      />
    </div>
  </UForm>
</template>
