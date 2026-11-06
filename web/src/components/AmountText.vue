<template>
  <span>{{ display }}</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { yuan } from '@/utils/money'

const props = withDefaults(defineProps<{ cents: number; suffix?: boolean }>(), { suffix: true })

const display = computed(() => `${yuan(props.cents)}${props.suffix ? ' 元' : ''}`)
</script>
