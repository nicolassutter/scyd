<script setup lang="ts">
definePageMeta({
  middleware: ["authenticated"],
});
</script>

<template>
  <div class="min-h-screen flex flex-col items-center p-6 pt-36">
    <Hero />
    <DownloadSection />
  </div>
</template>
