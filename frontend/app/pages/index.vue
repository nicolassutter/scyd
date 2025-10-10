<script setup lang="ts">
import * as z from "zod";
import { useMutation } from "@tanstack/vue-query";
import { postApiV1Download } from "~/utils/client";
import type { FormSubmitEvent } from "@nuxt/ui";

const schema = z.object({
  url: z.url(),
});

type Schema = z.output<typeof schema>;

const state = reactive<Partial<Schema>>({
  url: undefined,
});

const downloadItems = ref<
  {
    id: string;
    url: string;
    loading: boolean;
    error?: string;
    thumbnailUrl?: string;
    downloadedFiles: string[];
  }[]
>([]);

const mutation = useMutation({
  mutationKey: ["download", downloadItems],
  mutationFn: async (data: Schema & { id: string }) => {
    downloadItems.value.push({
      id: data.id,
      url: data.url,
      loading: true,
      downloadedFiles: [],
    });

    const output = await postApiV1Download({
      body: { url: data.url },
    });

    return output.data;
  },
  onSettled: (response, err, { id }) => {
    downloadItems.value = downloadItems.value.map((item) => {
      // update the status of the item that matches the id
      if (item.id === id) {
        return {
          ...item,
          loading: false,
          error: err
            ? "Failed to download, check the server logs for details."
            : undefined,
          downloadedFiles: response?.downloaded_files ?? [],
        };
      }

      return item;
    });
  },
});

function handleSubmit(event: FormSubmitEvent<Schema>) {
  mutation.mutate({ ...event.data, id: crypto.randomUUID() });
}
</script>

<template>
  <UContainer class="flex flex-col items-center gap-8">
    <h1 class="text-center font-bold text-5xl">Insert a url to get started!</h1>

    <UForm
      :schema="schema"
      :state="state"
      class="flex gap-4 items-start justify-center w-full"
      @submit="handleSubmit"
    >
      <UFormField name="url" class="w-full max-w-xl">
        <UInput
          v-model="state.url"
          size="xl"
          class="w-full"
          aria-label="Url"
          placeholder="https://youtube.com/never-gonna-give-u-up"
        />
      </UFormField>

      <UButton size="xl" type="submit">Download</UButton>
    </UForm>

    <h2 class="text-lg font-semibold mb-4">Your downloads</h2>

    <div class="grid gap-2 w-full">
      <template v-for="item in downloadItems" :key="item.id">
        <DownloadItem
          :url="item.url"
          :loading="item.loading"
          :error="item.error"
          :downloaded-files="item.downloadedFiles"
        />
      </template>
    </div>
  </UContainer>
</template>
