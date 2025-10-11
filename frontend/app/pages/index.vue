<script setup lang="ts">
import * as z from "zod";
import type { FormSubmitEvent } from "@nuxt/ui";
import { useMutation } from "@tanstack/vue-query";
import { postApiV1Download } from "~/utils/client";

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
  }[]
>([]);

const startDownloadMutation = useMutation({
  mutationFn: async (url: string) => {
    const response = await postApiV1Download({
      body: { url },
    });

    return response.data;
  },
  onSuccess: (data, inputUrl) => {
    downloadItems.value.push({
      id: data?.task_id,
      url: inputUrl,
    });
  },
});

async function handleSubmit(event: FormSubmitEvent<Schema>) {
  startDownloadMutation.mutate(event.data.url);
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
        <DownloadItem :id="item.id" :url="item.url" />
      </template>
    </div>
  </UContainer>
</template>
