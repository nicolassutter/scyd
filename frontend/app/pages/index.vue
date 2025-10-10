<script setup lang="ts">
import * as z from "zod";
import type { FormSubmitEvent } from "@nuxt/ui";
import { useMutation } from "@tanstack/vue-query";

const schema = z.object({
  url: z.url(),
});

type Schema = z.output<typeof schema>;

const state = reactive<Partial<Schema>>({
  url: undefined,
});

const downloadItems = ref<{ url: string; loading: boolean }[]>([]);

const mutation = useMutation({
  mutationKey: ["download", downloadItems],
  mutationFn: async (data: Schema) => {
    downloadItems.value.push({ url: data.url, loading: true });

    await $fetch("/api/v1/download", {
      method: "POST",
      body: data,
    });

    return data;
  },
  onSettled: (data: Schema) => {
    const idx = downloadItems.value.findIndex((item) => item.url === data.url);
    if (idx > -1) item.loading = false;
  },
});
</script>

<template>
  <UContainer class="flex flex-col items-center gap-8">
    <h1 class="text-center font-bold text-5xl">Insert a url to get started!</h1>

    <UForm
      :schema="schema"
      :state="state"
      class="flex gap-4 items-start justify-center w-full"
      @submit="(event) => mutation.mutate(event.data)"
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
      <template v-for="item in downloadItems">
        <DownloadItem :url="item.url" :loading="item.loading"></DownloadItem>
      </template>
    </div>
  </UContainer>
</template>
