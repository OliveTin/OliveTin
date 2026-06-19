<template>
  <span v-if="showLimits" class="action-group-limit">
    {{ t('logs.action-group-limits', { concurrent: maxConcurrent, queueSize }) }}
  </span>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  maxConcurrent: {
    type: Number,
    default: 0
  },
  queueSize: {
    type: Number,
    default: 0
  }
})

const { t } = useI18n()

const showLimits = computed(() => props.maxConcurrent > 0 && props.queueSize > 0)
</script>

<style scoped>
.action-group-limit {
  white-space: nowrap;
  font-weight: 500;
  font-size: smaller;
  color: var(--text-secondary, #666);
}
</style>
