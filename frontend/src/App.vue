<script setup>
import { RouterView } from "vue-router";
import { NMessageProvider } from "naive-ui";
</script>

<template>
  <n-message-provider>
    <n-config-provider> <RouterView /></n-config-provider>
  </n-message-provider>
</template>
