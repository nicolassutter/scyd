<script setup lang="ts">
import { useQuery } from "@tanstack/vue-query";
import { getApiV1Downloads } from "~/utils/client";

definePageMeta({
  middleware: ["authenticated"],
});

const downloadsQuery = useQuery({
  queryKey: ["downloads"],
  queryFn: async () => {
    const response = await getApiV1Downloads();
    const downloads = response.data?.downloads ?? [];
    return downloads;
  },
});
</script>

<template>
  <div class="min-h-screen flex flex-col items-center p-6 pt-36">
    <Hero @download-started="() => downloadsQuery.refetch()" />
    <DownloadSection :downloads="downloadsQuery.data.value ?? []" />
  </div>
</template>
