<template>
  <section id = "argument-popup">
    <div class="section-header">
      <h2>Start action: {{ title }}</h2>
    </div>
    <div class="section-content padding">
      <form @submit="handleSubmit">
        <template v-if="actionArguments.length > 0">

          <template v-for="arg in actionArguments" :key="arg.name">
              <label :for="arg.name">
                {{ formatLabel(arg.title) }}
              </label>

              <datalist v-if="(arg.suggestions && Object.keys(arg.suggestions).length > 0) || getBrowserSuggestions(arg).length > 0" :id="`${arg.name}-choices`">
                <option v-for="(suggestion, key) in arg.suggestions" :key="key" :value="key">
                  {{ suggestion }}
                </option>
                <option v-for="(suggestion, index) in getBrowserSuggestions(arg)" :key="`browser-${index}`" :value="suggestion">
                  {{ suggestion }}
                </option>
              </datalist>

              <select v-if="getInputComponent(arg) === 'select'" :id="arg.name" :name="arg.name" :value="getArgumentValue(arg)"
                :required="arg.required" @input="handleInput(arg, $event)" @change="handleChange(arg, $event)">
                <option v-for="choice in arg.choices" :key="choice.value" :value="choice.value">
                  {{ choice.title || choice.value }}
                </option>
              </select>
              
              <component v-else :is="getInputComponent(arg)" :id="arg.name" :name="arg.name" 
                :value="(arg.type === 'checkbox' || arg.type === 'confirmation') ? undefined : getArgumentValue(arg)"
                :checked="(arg.type === 'checkbox' || arg.type === 'confirmation') ? getArgumentValue(arg) : undefined"
                :list="(arg.suggestions || getBrowserSuggestions(arg).length > 0) ? `${arg.name}-choices` : undefined" 
                :type="getInputComponent(arg) !== 'select' ? getInputType(arg) : undefined"
                :rows="arg.type === 'raw_string_multiline' ? 5 : undefined"
                :step="arg.type === 'datetime' ? 1 : undefined" :pattern="getPattern(arg)"
                @input="handleInput(arg, $event)" @change="handleChange(arg, $event)" />

            <span class="argument-description" v-html="arg.description"></span>
          </template>
        </template>
        <div v-else>
          <p>No arguments required</p>
        </div>

        <div class="buttons">
          <button name="start" type="submit" :disabled="hasConfirmation && !confirmationChecked">
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
import { ref, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

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

  title.value = action.title
  icon.value = action.icon
  actionArguments.value = action.arguments || []
  argValues.value = {}
  formErrors.value = {}
  confirmationChecked.value = false
  hasConfirmation.value = false

  // Initialize values from query params or defaults
  actionArguments.value.forEach(arg => {
    if (arg.type === 'confirmation') {
      hasConfirmation.value = true
      const paramValue = getQueryParamValue(arg.name)
      let checkedValue = false
      if (paramValue !== null) {
        checkedValue = paramValue === '1' || paramValue === 'true' || paramValue === true
      } else if (arg.defaultValue !== undefined && arg.defaultValue !== '') {
        checkedValue = arg.defaultValue === '1' || arg.defaultValue === 'true' || arg.defaultValue === true
      }
      argValues.value[arg.name] = checkedValue
      confirmationChecked.value = checkedValue
    } else {
      const paramValue = getQueryParamValue(arg.name)
      if (arg.type === 'checkbox') {
        // For checkboxes, handle boolean default values properly
        if (paramValue !== null) {
          argValues.value[arg.name] = paramValue === '1' || paramValue === 'true' || paramValue === true
        } else if (arg.defaultValue !== undefined && arg.defaultValue !== '') {
          argValues.value[arg.name] = arg.defaultValue === '1' || arg.defaultValue === 'true' || arg.defaultValue === true
        } else {
          argValues.value[arg.name] = false
        }
      } else {
        argValues.value[arg.name] = paramValue !== null ? paramValue : arg.defaultValue || ''
      }
    }
  })

  // Run initial validation on all fields after DOM is updated
  await nextTick()
  for (const arg of actionArguments.value) {
    if (arg.type && !arg.type.startsWith('regex:') && arg.type !== 'select' && arg.type !== '' && arg.type !== 'confirmation' && arg.type !== 'checkbox') {
      await validateArgument(arg, argValues.value[arg.name] || '')
    }
  }
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

  if (arg.type === 'confirmation') {
    return 'checkbox'
  }

  if (arg.type === 'ascii_identifier' || arg.type === 'ascii') {
    return 'text'
  }

  if (arg.type === 'datetime') {
    return 'datetime-local'
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
  if (arg.type === 'checkbox' || arg.type === 'confirmation') {
    return argValues.value[arg.name] === '1' || argValues.value[arg.name] === true || argValues.value[arg.name] === 'true'
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

  // Skip validation for datetime - backend will handle mangling values without seconds
  if (arg.type === 'datetime') {
    const inputElement = document.getElementById(arg.name)
    if (inputElement) {
      inputElement.setCustomValidity('')
    }
    delete formErrors.value[arg.name]
    return
  }

  // Skip validation for checkbox and confirmation - they're always valid
  if (arg.type === 'checkbox' || arg.type === 'confirmation') {
    const inputElement = document.getElementById(arg.name)
    if (inputElement) {
      inputElement.setCustomValidity('')
    }
    delete formErrors.value[arg.name]
    return
  }

  try {
    const validateArgumentTypeArgs = {
      value: value,
      type: arg.type,
      bindingId: props.bindingId,
      argumentName: arg.name
    }

    const validation = await window.client.validateArgumentType(validateArgumentTypeArgs)

    // Get the input element to set custom validity
    const inputElement = document.getElementById(arg.name)
    
    if (validation.valid) {
      delete formErrors.value[arg.name]
      // Clear custom validity message
      if (inputElement) {
        inputElement.setCustomValidity('')
      }
    } else {
      formErrors.value[arg.name] = validation.description
      // Set custom validity message
      if (inputElement) {
        inputElement.setCustomValidity(validation.description)
      }
    }
  } catch (err) {
    console.warn('Validation failed:', err)
    // On error, clear any custom validity
    const inputElement = document.getElementById(arg.name)
    if (inputElement) {
      inputElement.setCustomValidity('')
    }
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

    if (arg.type === 'checkbox' || arg.type === 'confirmation') {
      value = value ? '1' : '0'
    }

    ret.push({
      name: arg.name,
      value: value
    })
  }

  return ret
}

function getUniqueId() {
  if (window.isSecureContext) {
    return window.crypto.randomUUID()
  } else {
    return Date.now().toString()
  }
}

function getBrowserSuggestions(arg) {
  if (!arg.suggestionsBrowserKey) {
    return []
  }
  
  try {
    const stored = localStorage.getItem(`olivetin-suggestions-${arg.suggestionsBrowserKey}`)
    if (stored) {
      const suggestions = JSON.parse(stored)
      return Array.isArray(suggestions) ? suggestions : []
    }
  } catch (err) {
    console.warn('Failed to load browser suggestions:', err)
  }
  
  return []
}

function saveBrowserSuggestions() {
  for (const arg of actionArguments.value) {
    if (arg.suggestionsBrowserKey) {
      const value = argValues.value[arg.name]
      
      // Only save non-empty values for non-checkbox/confirmation types
      if (value && value !== '' && arg.type !== 'checkbox' && arg.type !== 'confirmation') {
        try {
          const key = `olivetin-suggestions-${arg.suggestionsBrowserKey}`
          const stored = localStorage.getItem(key)
          let suggestions = []
          
          if (stored) {
            suggestions = JSON.parse(stored)
            if (!Array.isArray(suggestions)) {
              suggestions = []
            }
          }
          
          // Add value if not already present
          if (!suggestions.includes(value)) {
            suggestions.unshift(value) // Add to beginning
            // Keep only the most recent 50 suggestions
            if (suggestions.length > 50) {
              suggestions = suggestions.slice(0, 50)
            }
            localStorage.setItem(key, JSON.stringify(suggestions))
          }
        } catch (err) {
          console.warn('Failed to save browser suggestions:', err)
        }
      }
    }
  }
}

async function startAction(actionArgs) {
  const startActionArgs = {
    bindingId: props.bindingId,
    arguments: actionArgs,
    uniqueTrackingId: getUniqueId()
  }

  try {
    const response = await window.client.startAction(startActionArgs)
    console.log('Action started successfully with tracking ID:', response.executionTrackingId)
    return response
  } catch (err) {
    console.error('Failed to start action:', err)
    throw err
  }
}

async function handleSubmit(event) {
  // Set custom validity for required fields
  for (const arg of actionArguments.value) {
    const value = argValues.value[arg.name]
    const inputElement = document.getElementById(arg.name)
    
    if (arg.required && (!value || value === '')) {
      formErrors.value[arg.name] = 'This field is required'
      // Set custom validity for required field validation
      if (inputElement) {
        inputElement.setCustomValidity('This field is required')
      }
    }
  }

  const form = event.target
  if (!form.checkValidity()) {
    console.log('argument form has elements that failed validation')
    return
  }

  event.preventDefault()
  
  const argvs = getArgumentValues()
  console.log('argument form has elements that passed validation')
  
  // Save values to localStorage for arguments with suggestionsBrowserKey
  saveBrowserSuggestions()
  
  try {
    const response = await startAction(argvs)
    router.push(`/logs/${response.executionTrackingId}`)
  } catch (err) {
    console.error('Failed to start action:', err)
  }
}

function handleCancel() {
  router.back()
  clearBookmark()
}

function clearBookmark() {
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