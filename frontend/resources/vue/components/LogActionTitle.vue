<template>
  <span class="log-action-title">
    <slot>{{ actionTitle }}</slot>
    <span v-if="justification" class="log-justification">— {{ justification }}</span>
  </span>
</template>

<script setup>
defineProps({
  actionTitle: {
    type: String,
    default: ''
  },
  justification: {
    type: String,
    default: ''
  }
})
</script>

<style scoped>
.log-justification {
  font-size: 0.875rem;
  color: #888;
  font-weight: normal;
}
</style>
