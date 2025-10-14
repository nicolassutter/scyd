<script setup lang="ts">
import type { DropdownMenuItem } from "#ui/components/DropdownMenu.vue";
const { downloadsQuery, sortDownloadsMutation } = useDownloads();
const downloads = computed(() => downloadsQuery.data.value ?? []);
const toasts = useToast();

const actions = computed<DropdownMenuItem[]>(() => [
  {
    label: "Sort downloads directory",
    icon: "i-lucide:arrow-down-up",
    loading: sortDownloadsMutation.isPending.value,
    onSelect: () => {
      sortDownloadsMutation.mutate(undefined, {
        onSuccess(result) {
          if (result?.files_with_errors?.length) {
            toasts.add({
              title: "Warning",
              description: `Some downloads could not be sorted: ${result.files_with_errors.join(
                ", "
              )}`,
              color: "warning",
            });
            return;
          } else {
            toasts.add({
              title: "Success",
              description: "Downloads sorted successfully",
            });
          }
        },
      });
    },
  },
]);
</script>

<template>
  <UContainer>
    <section class="w-full py-16 flex flex-col items-start gap-6">
      <h2 class="text-3xl font-bold text-white text-balance">Downloads</h2>

      <UDropdownMenu
        :items="actions"
        :content="{
          align: 'start',
        }"
      >
        <UButton
          icon="i-lucide-menu"
          color="neutral"
          variant="outline"
          aria-label="Actions"
        />
      </UDropdownMenu>

      <div v-if="downloads.length" class="grid md:grid-cols-2 gap-6">
        <template v-for="download in downloads" :key="download.id">
          <DownloadItem v-bind="download" />
        </template>
      </div>

      <div v-else class="text-slate-400 italic text-lg">
        <p>Your downloads will display here.</p>
      </div>
    </section>
  </UContainer>
</template>
