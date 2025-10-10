<script setup lang="ts">
defineProps<{
  url: string;
  loading: boolean;
}>();
</script>

<template>
  <UCard variant="subtle">
    <template #header>
      <h3 class="">{url}</h3>
    </template>

    <Placeholder class="h-32" />

    <template #footer>
      <Placeholder class="h-8" />
    </template>
  </UCard>
</template>
