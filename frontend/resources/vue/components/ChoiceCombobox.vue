<template>
  <div class="choice-combobox" ref="rootRef">
    <input
      ref="searchInputRef"
      :id="id"
      type="text"
      class="choice-combobox-input"
      role="combobox"
      autocomplete="off"
      :aria-expanded="isOpen"
      :aria-controls="listboxId"
      :aria-activedescendant="activeDescendantId"
      :placeholder="placeholderText"
      :value="query"
      :required="required"
      @focus="handleFocus"
      @input="handleSearchInput"
      @keydown="handleKeydown"
      @blur="handleBlur"
    />
    <input
      :name="name"
      type="hidden"
      :value="modelValue"
    />
    <ul
      v-if="isOpen && filteredChoices.length > 0"
      :id="listboxId"
      role="listbox"
      class="choice-combobox-list"
    >
      <li
        v-for="(choice, index) in filteredChoices"
        :id="`${listboxId}-option-${index}`"
        :key="choice.value"
        role="option"
        :aria-selected="choice.value === modelValue"
        :class="{
          highlighted: index === highlightedIndex,
          selected: choice.value === modelValue
        }"
        @mousedown.prevent="selectChoice(choice)"
      >
        {{ choiceLabel(choice) }}
      </li>
    </ul>
    <div v-else-if="isOpen && query" class="choice-combobox-list choice-combobox-empty">
      No matching options
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  choiceDisplayLabel,
  syncStateFromModelValue
} from './choiceComboboxHelpers.js'

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

const closeOthersEvent = 'olivetin-choice-combobox-close-others'

const rootRef = ref(null)
const searchInputRef = ref(null)
const isOpen = ref(false)
const query = ref('')
const isUserFiltering = ref(false)
const highlightedIndex = ref(0)

const listboxId = computed(() => `${props.id}-listbox`)

const activeDescendantId = computed(() => {
  if (!isOpen.value || filteredChoices.value.length === 0) {
    return undefined
  }

  return `${listboxId.value}-option-${highlightedIndex.value}`
})

const placeholderText = computed(() => {
  if (props.required) {
    return 'Search and select...'
  }

  return 'Search options...'
})

const filteredChoices = computed(() => {
  if (!isUserFiltering.value) {
    return props.choices
  }

  const search = query.value.trim().toLowerCase()
  if (!search) {
    return props.choices
  }

  return props.choices.filter(choice => {
    const label = choiceLabel(choice).toLowerCase()
    const value = String(choice.value).toLowerCase()
    return label.includes(search) || value.includes(search)
  })
})

watch([() => props.modelValue, () => props.choices], () => {
  if (!isOpen.value) {
    syncFromModelValue()
  }
}, { immediate: true })

function choiceLabel(choice) {
  return choiceDisplayLabel(choice)
}

function syncFromModelValue() {
  const next = syncStateFromModelValue(props.choices, props.modelValue)
  query.value = next.query

  if (next.modelValue !== props.modelValue) {
    emitValue(next.modelValue)
  }
}

function selectedChoiceIndex(choices) {
  const index = choices.findIndex(choice => choice.value === props.modelValue)
  return index >= 0 ? index : 0
}

function openList() {
  document.dispatchEvent(new CustomEvent(closeOthersEvent, { detail: { id: props.id } }))
  isOpen.value = true
  highlightedIndex.value = selectedChoiceIndex(filteredChoices.value)
}

function closeList() {
  isOpen.value = false
  isUserFiltering.value = false
  syncFromModelValue()
}

function emitValue(value) {
  emit('update:modelValue', value)
}

function selectChoice(choice) {
  emitValue(choice.value)
  query.value = choiceLabel(choice)
  isOpen.value = false
}

function handleFocus() {
  if (!isOpen.value) {
    syncFromModelValue()
    isUserFiltering.value = false
  }

  openList()
}

function handleSearchInput(event) {
  isUserFiltering.value = true
  query.value = event.target.value
  openList()
  highlightedIndex.value = 0
}

function moveHighlight(delta) {
  if (filteredChoices.value.length === 0) {
    return
  }

  const nextIndex = highlightedIndex.value + delta
  if (nextIndex < 0) {
    highlightedIndex.value = filteredChoices.value.length - 1
    return
  }

  if (nextIndex >= filteredChoices.value.length) {
    highlightedIndex.value = 0
    return
  }

  highlightedIndex.value = nextIndex
}

function handleKeydown(event) {
  if (event.key === 'ArrowDown') {
    event.preventDefault()
    const wasOpen = isOpen.value
    openList()
    if (wasOpen) {
      moveHighlight(1)
    }
    return
  }

  if (event.key === 'ArrowUp') {
    event.preventDefault()
    const wasOpen = isOpen.value
    openList()
    if (wasOpen) {
      moveHighlight(-1)
    } else if (filteredChoices.value.length > 0) {
      highlightedIndex.value = filteredChoices.value.length - 1
    }
    return
  }

  if (event.key === 'Enter') {
    if (!isOpen.value || filteredChoices.value.length === 0) {
      return
    }

    event.preventDefault()
    selectChoice(filteredChoices.value[highlightedIndex.value])
    return
  }

  if (event.key === 'Escape') {
    event.preventDefault()
    closeList()
    searchInputRef.value?.blur()
  }
}

function handleBlur() {
  closeList()
}

function handleCloseOthers(event) {
  if (event.detail.id !== props.id) {
    closeList()
  }
}

function handleOutsideMouseDown(event) {
  if (!isOpen.value || rootRef.value?.contains(event.target)) {
    return
  }

  closeList()
}

watch(isOpen, open => {
  if (open) {
    document.addEventListener('mousedown', handleOutsideMouseDown, true)
    return
  }

  document.removeEventListener('mousedown', handleOutsideMouseDown, true)
})

onMounted(() => {
  document.addEventListener(closeOthersEvent, handleCloseOthers)
})

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', handleOutsideMouseDown, true)
  document.removeEventListener(closeOthersEvent, handleCloseOthers)
})
</script>

<style scoped>
.choice-combobox {
  position: relative;
  width: 100%;
}

.choice-combobox:focus-within {
  z-index: 11;
}

.choice-combobox-input {
  width: 100%;
}

.choice-combobox-list {
  position: absolute;
  z-index: 10;
  left: 0;
  right: 0;
  max-height: 12rem;
  overflow-y: auto;
  margin: 0.125rem 0 0;
  padding: 0;
  list-style: none;
  border: 1px solid var(--border-color, #ccc);
  border-radius: 0.25rem;
  background: var(--standout-bg-color, #fff);
  color: var(--text-color, inherit);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.12);
}

.choice-combobox-list li {
  padding: 0.375rem 0.5rem;
  cursor: pointer;
}

.choice-combobox-list li.highlighted,
.choice-combobox-list li:hover {
  background: var(--hover-background-color, #eef3ff);
  color: var(--hover-text-color, inherit);
}

.choice-combobox-list li.selected {
  font-weight: 600;
}

.choice-combobox-empty {
  padding: 0.375rem 0.5rem;
  color: var(--disabled-text-color, #666);
  font-size: 0.875rem;
}

@media (prefers-color-scheme: dark) {
  .choice-combobox-list,
  .choice-combobox-empty {
    background-color: #4e4e4e;
    color: #ddd;
    border-color: var(--border-color, #595959);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.45);
  }

  .choice-combobox-list li.highlighted,
  .choice-combobox-list li:hover {
    background-color: var(--hover-background-color, #1d345c);
    color: #fff;
  }

  .choice-combobox-empty {
    color: var(--disabled-text-color, #999);
  }
}
</style>
