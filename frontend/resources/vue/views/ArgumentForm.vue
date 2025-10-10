<template>
  <section id = "argument-popup">
    <div class="section-header">
      <h2>Start action: {{ title }}</h2>
    </div>
    <div class="section-content padding">
      <form @submit.prevent="handleSubmit">
        <template v-if="actionArguments.length > 0">

          <template v-for="arg in actionArguments" :key="arg.name" class="argument-group">
            <label :for="arg.name">
              {{ formatLabel(arg.title) }}
            </label>

            <datalist v-if="arg.suggestions && Object.keys(arg.suggestions).length > 0" :id="`${arg.name}-choices`">
              <option v-for="(suggestion, key) in arg.suggestions" :key="key" :value="key">
                {{ suggestion }}
              </option>
            </datalist>

            <select v-if="getInputComponent(arg) === 'select'" :id="arg.name" :name="arg.name" :value="getArgumentValue(arg)"
              :required="arg.required" @input="handleInput(arg, $event)" @change="handleChange(arg, $event)">
              <option v-for="choice in arg.choices" :key="choice.value" :value="choice.value">
                {{ choice.title || choice.value }}
              </option>
            </select>
            
            <component v-else :is="getInputComponent(arg)" :id="arg.name" :name="arg.name" :value="getArgumentValue(arg)"
              :list="arg.suggestions ? `${arg.name}-choices` : undefined" 
              :type="getInputComponent(arg) !== 'select' ? getInputType(arg) : undefined"
              :rows="arg.type === 'raw_string_multiline' ? 5 : undefined"
              :step="arg.type === 'datetime' ? 1 : undefined" :pattern="getPattern(arg)" :required="arg.required"
              @input="handleInput(arg, $event)" @change="handleChange(arg, $event)" />

            <span class="argument-description" v-html="arg.description"></span>
          </template>
        </template>
        <div v-else>
          <p>No arguments required</p>
        </div>

        <div class="buttons">
          <button name="start" type="submit" :disabled="!isFormValid || (hasConfirmation && !confirmationChecked)">
            Start
          </button>
          <button name="cancel" type="button" @click="handleCancel">
            Cancel
          </button>
        </div>
      </form>
    </div>
  </section>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const emit = defineEmits(['submit', 'cancel', 'close'])

// Reactive data
const dialog = ref(null)
const title = ref('')
const icon = ref('')
//const arguments = ref([])
const argValues = ref({})
const confirmationChecked = ref(false)
const hasConfirmation = ref(false)
const formErrors = ref({})
const actionArguments = ref([])

// Computed properties
const isFormValid = computed(() => Object.keys(formErrors.value).length === 0)

const props = defineProps({
  bindingId: {
    type: String,
    required: true
  }
})

// Methods
async function setup() {
  const ret = await window.client.getActionBinding({
    bindingId: props.bindingId
  })

  const action = ret.action
  console.log('action', action)

  title.value = action.title
  icon.value = action.icon
  actionArguments.value = action.arguments || []
  argValues.value = {}
  formErrors.value = {}
  confirmationChecked.value = false
  hasConfirmation.value = false

  // Initialize values from query params or defaults
  actionArguments.value.forEach(arg => {
    const paramValue = getQueryParamValue(arg.name)
    argValues.value[arg.name] = paramValue !== null ? paramValue : arg.defaultValue || ''

    if (arg.type === 'confirmation') {
      hasConfirmation.value = true
    }
  })
}

function getQueryParamValue(paramName) {
  const params = new URLSearchParams(window.location.search.substring(1))
  return params.get(paramName)
}

function formatLabel(title) {
  const lastChar = title.charAt(title.length - 1)
  if (lastChar === '?' || lastChar === '.' || lastChar === ':') {
    return title
  }
  return title + ':'
}

function getInputComponent(arg) {
  if (arg.type === 'html') {
    return 'div'
  } else if (arg.type === 'raw_string_multiline') {
    return 'textarea'
  } else if (arg.choices && arg.choices.length > 0 && (arg.type === 'select' || arg.type === '')) {
    return 'select'
  } else {
    return 'input'
  }
}

function getInputType(arg) {
  if (arg.type === 'html' || arg.type === 'raw_string_multiline' || arg.type === 'select') {
    return undefined
  }

  if (arg.type === 'ascii_identifier') {
    return 'text'
  }

  return arg.type
}

function getPattern(arg) {
  if (arg.type && arg.type.startsWith('regex:')) {
    return arg.type.replace('regex:', '')
  }
  return undefined
}

function getArgumentValue(arg) {
  if (arg.type === 'checkbox') {
    return argValues.value[arg.name] === '1' || argValues.value[arg.name] === true
  }
  return argValues.value[arg.name] || ''
}

function handleInput(arg, event) {
  const value = event.target.type === 'checkbox' ? event.target.checked : event.target.value
  argValues.value[arg.name] = value
  updateUrlWithArg(arg.name, value)
}

function handleChange(arg, event) {
  if (arg.type === 'confirmation') {
    confirmationChecked.value = event.target.checked
    return
  }

  // Validate the input
  validateArgument(arg, event.target.value)
}

async function validateArgument(arg, value) {
  if (!arg.type || arg.type.startsWith('regex:')) {
    return
  }

  try {
    const validateArgumentTypeArgs = {
      value: value,
      type: arg.type
    }

    const validation = await window.validateArgumentType(validateArgumentTypeArgs)

    if (validation.valid) {
      delete formErrors.value[arg.name]
    } else {
      formErrors.value[arg.name] = validation.description
    }
  } catch (err) {
    console.warn('Validation failed:', err)
  }
}

function updateUrlWithArg(name, value) {
  if (name && value !== undefined) {
    const url = new URL(window.location.href)

    // Don't add passwords to URL
    const arg = actionArguments.value.find(a => a.name === name)
    if (arg && arg.type === 'password') {
      return
    }

    url.searchParams.set(name, value)
    window.history.replaceState({}, '', url.toString())
  }
}

function getArgumentValues() {
  const ret = []

  for (const arg of actionArguments.value) {
    let value = argValues.value[arg.name] || ''

    if (arg.type === 'checkbox') {
      value = value ? '1' : '0'
    }

    ret.push({
      name: arg.name,
      value: value
    })
  }

  return ret
}

function handleSubmit() {
  // Validate all inputs
  let isValid = true

  for (const arg of actionArguments.value) {
    const value = argValues.value[arg.name]
    if (arg.required && (!value || value === '')) {
      formErrors.value[arg.name] = 'This field is required'
      isValid = false
    }
  }

  if (!isValid) {
    return
  }

  const argvs = getArgumentValues()
  emit('submit', argvs)
  close()
}

function handleCancel() {
  router.back()
  clearBookmark()
  emit('cancel')
  close()
}

function handleClose() {
  emit('close')
}

function clearBookmark() {
  // Remove the action from the URL
  window.history.replaceState({
    path: window.location.pathname
  }, '', window.location.pathname)
}

function show() {
  if (dialog.value) {
    dialog.value.showModal()
  }
}

function close() {
  if (dialog.value) {
    dialog.value.close()
  }
}

// Expose methods for parent components
defineExpose({
  show,
  close
})

// Lifecycle
onMounted(() => {
  setup()
})
</script>

<style scoped>

form {
  grid-template-columns: max-content auto auto;
}

.argument-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.argument-group label {
  font-weight: 500;
  color: #333;
}

.argument-group input:invalid,
.argument-group select:invalid,
.argument-group textarea:invalid {
  border-color: #dc3545;
}

.argument-description {
  font-size: 0.875rem;
  color: #666;
  margin-top: 0.25rem;
}

.buttons {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  padding-top: 1rem;
  border-top: 1px solid #eee;
}

/* Checkbox specific styling */
.argument-group input[type="checkbox"] {
  width: auto;
  margin-right: 0.5rem;
}

.argument-group input[type="checkbox"]+label {
  display: inline;
  font-weight: normal;
}
</style>