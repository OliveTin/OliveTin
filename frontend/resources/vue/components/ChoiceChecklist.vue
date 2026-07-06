<template>
  <div class="choice-checklist" :id="`${id}-wrapper`">
    <div class="choice-checklist-controls">
      <button type="button" class="choice-checklist-control" @click="selectAll">
        Select all
      </button>
      <button type="button" class="choice-checklist-control" @click="selectNone">
        Select none
      </button>
    </div>
    <fieldset class="choice-checklist-fieldset">
      <legend class="visually-hidden">{{ name }}</legend>
      <label
        v-for="(choice, index) in choices"
        :key="choice.value"
        class="choice-checklist-item"
        :for="`${id}-${index}`"
      >
        <input
          :id="`${id}-${index}`"
          type="checkbox"
          :checked="isSelected(choice.value)"
          @change="handleToggle(choice.value)"
        />
        <span>{{ choiceLabel(choice) }}</span>
      </label>
    </fieldset>
    <input
      :id="`${id}-value`"
      :name="name"
      type="hidden"
      :value="modelValue"
      :required="required && modelValue === ''"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import {
  allChoiceValues,
  choiceLabel,
  formatChecklistValue,
  parseChecklistValue,
  toggleChoice
} from '../utils/choiceChecklistHelpers.js'

const props = defineProps({
  id: {
    type: String,
    required: true
  },
  name: {
    type: String,
    required: true
  },
  choices: {
    type: Array,
    required: true
  },
  modelValue: {
    type: String,
    default: ''
  },
  required: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue'])

const selectedValues = computed(() => parseChecklistValue(props.modelValue))

function isSelected(value) {
  return selectedValues.value.includes(value)
}

function emitSelection(selected) {
  emit('update:modelValue', formatChecklistValue(selected))
}

function handleToggle(value) {
  emitSelection(toggleChoice(selectedValues.value, value))
}

function selectAll() {
  emitSelection(allChoiceValues(props.choices))
}

function selectNone() {
  emitSelection([])
}
</script>

<style scoped>
.choice-checklist {
  display: flex;
  flex-direction: column;
  gap: 0.5em;
}

.choice-checklist-controls {
  display: flex;
  gap: 0.75em;
}

.choice-checklist-control {
  background: none;
  border: none;
  color: inherit;
  cursor: pointer;
  font: inherit;
  padding: 0;
  text-decoration: underline;
}

.choice-checklist-fieldset {
  border: none;
  display: grid;
  gap: 0.5em 1em;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  margin: 0;
  padding: 0;
}

.choice-checklist-item {
  align-items: center;
  display: flex;
  gap: 0.4em;
  margin: 0;
}

.choice-checklist-item input[type="checkbox"] {
  margin: 0;
}

.visually-hidden {
  border: 0;
  clip: rect(0 0 0 0);
  height: 1px;
  margin: -1px;
  overflow: hidden;
  padding: 0;
  position: absolute;
  white-space: nowrap;
  width: 1px;
}
</style>
