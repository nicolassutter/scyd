<script setup lang="ts">
useHead({
  title: "Scyd - Self hosted music downloader",
  meta: [
    {
      name: "description",
      content:
        "Scyd is a simple and automated music downloader built with Nuxt 3, Go and open-source projects.",
    },
    { name: "theme-color", content: "#1e293b" },
  ],
  link: [
    { rel: "icon", href: "/favicon.ico" },
    {
      rel: "apple-touch-icon",
      sizes: "180x180",
      href: "/apple-touch-icon-180x180.png",
    },
    { rel: "mask-icon", href: "/maskable-icon-512x512.svg", color: "#FFFFFF" },
  ],
});
</script>

<template>
  <UApp>
    <VitePwaManifest />
    <NuxtRouteAnnouncer />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </UApp>
</template>
