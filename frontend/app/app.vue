<script setup lang="ts">
import { client } from "./utils/client/client.gen";

// Set the base URL for the API client before making any requests
client.setConfig({
  baseUrl: import.meta.env.DEV ? "http://localhost:3000" : undefined,
  throwOnError: true,
});
</script>

<template>
  <UApp>
    <div>
      <NuxtRouteAnnouncer />
      <NuxtPage />
    </div>
  </UApp>
</template>
