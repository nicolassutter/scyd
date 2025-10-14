<script setup lang="ts">
definePageMeta({
  middleware: ["authenticated"],
});
const { logoutMutation } = useAuth();
</script>

<template>
  <div class="min-h-screen flex flex-col items-center p-6 pt-36">
    <UButton
      title="Sign out"
      icon="i-lucide:log-out"
      class="absolute top-6 right-6"
      variant="subtle"
      color="neutral"
      @click="() => logoutMutation.mutate()"
    />
    <Hero />
    <DownloadSection />
  </div>
</template>
