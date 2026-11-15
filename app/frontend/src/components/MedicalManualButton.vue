<script setup lang="ts">
import { BookOpen, Plus } from "lucide-vue-next"

defineProps<{
  active?: boolean
}>()

const emit = defineEmits<{
  click: []
}>()
</script>

<template>
  <button
    type="button"
    class="hud-card w-full h-full min-h-36 p-5 flex flex-col items-center justify-center gap-3 text-main transition-colors hover:bg-surface focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-line cursor-pointer"
    :class="active ? 'bg-surface' : ''"
    :aria-pressed="Boolean(active)"
    @click="emit('click')"
  >
    <div class="relative flex items-center justify-center" aria-hidden="true">
      <BookOpen class="w-14 h-14 text-main stroke-[1.6]" />
      <span
        class="absolute -top-1.5 -right-2 w-6 h-6 rounded-full bg-ground border border-solid border-line flex items-center justify-center"
      >
        <Plus class="w-3.5 h-3.5 text-ok stroke-[2.5]" />
      </span>
    </div>
    <span class="text-base font-medium leading-tight text-center text-main">
      Medical
      <br />
      Manual
    </span>
  </button>
</template>
