<script setup lang="ts">
import { ref } from "vue"
import { Ping } from "@bindings/hrig/app/app.js"

const status = ref("Ready")

async function checkBridge() {
  status.value = await Ping()
}
</script>

<template>
  <main class="flex-1 flex flex-col items-center justify-center p-8 gap-4">
    <h1 class="text-4xl font-black text-main">HRIG</h1>
    <p class="text-base text-side">Wails v3 + Vue 3 + UnoCSS + Bun scaffold.</p>
    <div class="flex items-center gap-3">
      <button type="button" class="button-main" @click="checkBridge">
        Test Go Bridge ({{ status }})
      </button>
    </div>
  </main>
</template>
