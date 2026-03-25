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

              <template v-if="arg.type === 'file_upload'">
                <div class="file-upload-field">
                  <div
                    class="file-upload-dropzone"
                    :class="{ 'file-upload-dropzone--active': (fileUploadDragDepth[arg.name] || 0) > 0 }"
                    @dragenter.prevent="onFileDragEnter(arg)"
                    @dragover.prevent="onFileDragOver"
                    @dragleave="onFileDragLeave(arg)"
                    @drop.prevent="onFileDrop(arg, $event)"
                  >
                    <input
                      :id="arg.name"
                      :name="arg.name"
                      type="file"
                      class="file-upload-input-overlay"
                      :accept="getFileAccept(arg)"
                      @change="handleChange(arg, $event)"
                    />
                    <div class="file-upload-dropzone-inner">
                      <span class="file-upload-prompt">{{ fileUploadPrompt(arg) }}</span>
                      <span v-if="formErrors[arg.name]" class="file-upload-error">{{ formErrors[arg.name] }}</span>
                    </div>
                    </div>
                </div>
              <span class="argument-description">
                <p v-html="arg.description"></p>
                <p v-if="maxUploadSizeSummary(arg)" class="file-upload-mime-types">{{ maxUploadSizeSummary(arg) }}</p>
                <p v-if="mimeTypesSummary(arg)" class="file-upload-mime-types">{{ mimeTypesSummary(arg) }}</p>
              </span>
          </template>

              <template v-else>
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
import { ref, reactive, onMounted, nextTick } from 'vue'
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
const popupOnStart = ref('')
const fileUploadDragDepth = reactive({})
const fileUploadDisplayName = reactive({})

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
  popupOnStart.value = action.popupOnStart || ''
  actionArguments.value = action.arguments || []
  argValues.value = {}
  formErrors.value = {}
  for (const key of Object.keys(fileUploadDragDepth)) {
    delete fileUploadDragDepth[key]
  }
  for (const key of Object.keys(fileUploadDisplayName)) {
    delete fileUploadDisplayName[key]
  }
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
    if (arg.type && !arg.type.startsWith('regex:') && arg.type !== 'select' && arg.type !== '' && arg.type !== 'confirmation' && arg.type !== 'checkbox' && arg.type !== 'file_upload') {
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

function getFileAccept(arg) {
  if (arg.type !== 'file_upload' || !arg.allowedMimeTypes || arg.allowedMimeTypes.length === 0) {
    return undefined
  }
  return arg.allowedMimeTypes.join(',')
}

function mimeTypesSummary(arg) {
  if (arg.type !== 'file_upload' || !arg.allowedMimeTypes || arg.allowedMimeTypes.length === 0) {
    return ''
  }
  return 'Supported MIME types: ' + arg.allowedMimeTypes.join(', ')
}

/** SI byte formatting (matches server-side humanize-style defaults such as "10 MB"). */
function formatBytesDecimal(numBytes) {
  if (!Number.isFinite(numBytes) || numBytes < 0) {
    return ''
  }
  const n = Math.floor(numBytes)
  if (n < 1000) {
    return `${n} B`
  }
  const units = ['kB', 'MB', 'GB', 'TB']
  let v = n
  let i = 0
  while (v >= 1000 && i < units.length) {
    v /= 1000
    i++
  }
  const unit = units[i - 1]
  const rounded = v < 10 ? Math.round(v * 10) / 10 : Math.round(v)
  return `${rounded} ${unit}`
}

function maxUploadSizeSummary(arg) {
  if (arg.type !== 'file_upload') {
    return ''
  }
  const max = maxUploadBytesNumber(arg)
  if (max <= 0) {
    return ''
  }
  return `Max file size: ${formatBytesDecimal(max)}`
}

function fileUploadPrompt(arg) {
  if (fileUploadDisplayName[arg.name]) {
    return fileUploadDisplayName[arg.name]
  }
  return 'Drop a file here or click to browse'
}

function onFileDragEnter(arg) {
  fileUploadDragDepth[arg.name] = (fileUploadDragDepth[arg.name] || 0) + 1
}

function onFileDragLeave(arg) {
  const next = Math.max(0, (fileUploadDragDepth[arg.name] || 0) - 1)
  if (next === 0) {
    delete fileUploadDragDepth[arg.name]
  } else {
    fileUploadDragDepth[arg.name] = next
  }
}

function onFileDragOver(event) {
  event.dataTransfer.dropEffect = 'copy'
}

function onFileDrop(arg, event) {
  delete fileUploadDragDepth[arg.name]
  const file = event.dataTransfer && event.dataTransfer.files && event.dataTransfer.files[0]
  if (file) {
    processStagedFileUpload(arg, file)
  }
}

function maxUploadBytesNumber(arg) {
  if (arg.maxUploadBytes === undefined || arg.maxUploadBytes === null) {
    return 0
  }
  return typeof arg.maxUploadBytes === 'bigint' ? Number(arg.maxUploadBytes) : Number(arg.maxUploadBytes)
}

function getInputType(arg) {
  if (arg.type === 'html' || arg.type === 'raw_string_multiline' || arg.type === 'select') {
    return undefined
  }

  if (arg.type === 'confirmation') {
    return 'checkbox'
  }

  if (arg.type === 'file_upload') {
    return 'file'
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

async function uploadStagedFile(arg, file) {
  const formData = new FormData()
  formData.append('binding_id', props.bindingId)
  formData.append('argument_name', arg.name)
  formData.append('file', file)

  const res = await fetch('/api/upload/action-argument', {
    method: 'POST',
    body: formData,
    credentials: 'same-origin'
  })
  const text = await res.text()
  if (!res.ok) {
    throw new Error(text || `Upload failed (${res.status})`)
  }
  let data
  try {
    data = JSON.parse(text)
  } catch (e) {
    throw new Error('Invalid upload response')
  }
  if (!data.uploadToken) {
    throw new Error('Upload response missing token')
  }
  return data.uploadToken
}

function handleChange(arg, event) {
  if (arg.type === 'confirmation') {
    confirmationChecked.value = event.target.checked
    return
  }

  if (arg.type === 'file_upload') {
    handleFileUploadChange(arg, event)
    return
  }

  // Validate the input
  validateArgument(arg, event.target.value)
}

async function processStagedFileUpload(arg, file) {
  const inputEl = document.getElementById(arg.name)
  if (!file) {
    argValues.value[arg.name] = ''
    delete fileUploadDisplayName[arg.name]
    if (inputEl) {
      inputEl.setCustomValidity('')
    }
    return
  }
  const maxBytes = maxUploadBytesNumber(arg)
  if (maxBytes > 0 && file.size > maxBytes) {
    const msg = `File is too large (max ${formatBytesDecimal(maxBytes)})`
    if (inputEl) {
      inputEl.setCustomValidity(msg)
    }
    formErrors.value[arg.name] = msg
    delete fileUploadDisplayName[arg.name]
    return
  }
  try {
    const token = await uploadStagedFile(arg, file)
    argValues.value[arg.name] = token
    fileUploadDisplayName[arg.name] = file.name
    if (inputEl) {
      inputEl.setCustomValidity('')
    }
    delete formErrors.value[arg.name]
    await validateArgument(arg, token)
  } catch (err) {
    console.warn('Upload failed:', err)
    const msg = err.message || 'Upload failed'
    formErrors.value[arg.name] = msg
    if (inputEl) {
      inputEl.setCustomValidity(msg)
    }
    argValues.value[arg.name] = ''
    delete fileUploadDisplayName[arg.name]
  }
}

async function handleFileUploadChange(arg, event) {
  const file = event.target.files && event.target.files[0]
  await processStagedFileUpload(arg, file)
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
    const inputElement = document.getElementById(arg.name)
    if (arg.type === 'file_upload') {
      const msg = 'Could not validate upload; try again or check your connection'
      formErrors.value[arg.name] = msg
      if (inputElement) {
        inputElement.setCustomValidity(msg)
      }
    } else if (inputElement) {
      inputElement.setCustomValidity('')
    }
  }
}

function updateUrlWithArg(name, value) {
  if (name && value !== undefined) {
    const url = new URL(window.location.href)

    // Don't add passwords to URL
    const arg = actionArguments.value.find(a => a.name === name)
    if (arg && (arg.type === 'password' || arg.type === 'file_upload')) {
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

      // Only save non-empty values for non-checkbox/confirmation/password types
      if (value && value !== '' && arg.type !== 'checkbox' && arg.type !== 'confirmation' && arg.type !== 'password' && arg.type !== 'file_upload') {
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
    if (popupOnStart.value && popupOnStart.value.includes('execution-dialog')) {
      router.push(`/logs/${response.executionTrackingId}`)
    } else {
      router.back()
    }
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

.file-upload-field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-width: 0;
}

.file-upload-dropzone {
  position: relative;
  min-height: 5.5rem;
  border: 2px dashed #bbb;
  border-radius: 0.5rem;
  background: #fafafa;
  transition: border-color 0.15s ease, background 0.15s ease, box-shadow 0.15s ease;
}

.file-upload-dropzone:hover:not(.file-upload-dropzone--active) {
  border-color: #7a9bbb;
  background: #f3f7fb;
  box-shadow: 0 2px 8px rgba(68, 136, 204, 0.12);
}

.file-upload-dropzone--active {
  border-color: #4488cc;
  background: #f0f6fc;
}

@media (prefers-color-scheme: dark) {
  .file-upload-dropzone {
    border-color: #555;
    background: #222;
  }
  .file-upload-dropzone:hover:not(.file-upload-dropzone--active) {
    border-color: #7a9bbb;
    background: #222;
    box-shadow: 0 2px 8px rgba(68, 136, 204, 0.12);
  }
  .file-upload-dropzone--active {
    border-color: #4488cc;
    background: #2a3b4c;
  }
}



.file-upload-input-overlay {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  margin: 0;
  padding: 0;
  opacity: 0;
  cursor: pointer;
  z-index: 2;
  font-size: 0;
}

.file-upload-dropzone-inner {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  min-height: 5.5rem;
  padding: 0.75rem 1rem;
  text-align: center;
  pointer-events: none;
}

.file-upload-prompt {
  font-size: 0.9375rem;
  word-break: break-word;
}

.file-upload-error {
  font-size: 0.8125rem;
  color: #b00020;
}

.file-upload-mime-types {
  font-size: 0.8125rem;
  color: #555;
  margin: 0;
  margin-top: 0.15rem;
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
