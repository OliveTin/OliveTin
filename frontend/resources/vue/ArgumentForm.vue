<template>
  <dialog 
    ref="dialog" 
    title="Arguments" 
    class="action-arguments"
    @close="handleClose"
  >
    <form class="padded-content" @submit.prevent="handleSubmit">
      <div class="wrapper">
        <div class="action-header">
          <span class="icon" v-html="icon"></span>
          <h2>{{ title }}</h2>
        </div>

        <div class="arguments">
          <div 
            v-for="arg in arguments" 
            :key="arg.name"
            class="argument-group"
          >
            <label :for="arg.name">
              {{ formatLabel(arg.title) }}
            </label>
            
            <datalist 
              v-if="arg.suggestions && Object.keys(arg.suggestions).length > 0"
              :id="`${arg.name}-choices`"
            >
              <option 
                v-for="(suggestion, key) in arg.suggestions" 
                :key="key"
                :value="key"
              >
                {{ suggestion }}
              </option>
            </datalist>
            
            <component 
              :is="getInputComponent(arg)"
              :id="arg.name"
              :name="arg.name"
              :value="getArgumentValue(arg)"
              :list="arg.suggestions ? `${arg.name}-choices` : undefined"
              :type="getInputType(arg)"
              :rows="arg.type === 'raw_string_multiline' ? 5 : undefined"
              :step="arg.type === 'datetime' ? 1 : undefined"
              :pattern="getPattern(arg)"
              :required="arg.required"
              @input="handleInput(arg, $event)"
              @change="handleChange(arg, $event)"
            />
            
            <span 
              v-if="arg.description"
              class="argument-description"
              v-html="arg.description"
            ></span>
          </div>
        </div>

        <div class="buttons">
          <button 
            name="start" 
            type="submit"
            :disabled="!isFormValid || (hasConfirmation && !confirmationChecked)"
          >
            Start
          </button>
          <button 
            name="cancel" 
            type="button"
            @click="handleCancel"
          >
            Cancel
          </button>
        </div>
      </div>
    </form>
  </dialog>
</template>

<script>
export default {
  name: 'ArgumentForm',
  props: {
    actionData: {
      type: Object,
      required: true
    }
  },
  data() {
    return {
      title: '',
      icon: '',
      arguments: [],
      argValues: {},
      confirmationChecked: false,
      hasConfirmation: false,
      formErrors: {}
    }
  },
  computed: {
    isFormValid() {
      return Object.keys(this.formErrors).length === 0
    }
  },
  mounted() {
    this.setup()
  },
  methods: {
    setup() {
      this.title = this.actionData.title
      this.icon = this.actionData.icon
      this.arguments = this.actionData.arguments || []
      this.argValues = {}
      this.formErrors = {}
      this.confirmationChecked = false
      this.hasConfirmation = false
      
      // Initialize values from query params or defaults
      this.arguments.forEach(arg => {
        const paramValue = this.getQueryParamValue(arg.name)
        this.argValues[arg.name] = paramValue !== null ? paramValue : arg.defaultValue || ''
        
        if (arg.type === 'confirmation') {
          this.hasConfirmation = true
        }
      })
    },
    
    getQueryParamValue(paramName) {
      const params = new URLSearchParams(window.location.search.substring(1))
      return params.get(paramName)
    },
    
    formatLabel(title) {
      const lastChar = title.charAt(title.length - 1)
      if (lastChar === '?' || lastChar === '.' || lastChar === ':') {
        return title
      }
      return title + ':'
    },
    
    getInputComponent(arg) {
      if (arg.type === 'html') {
        return 'div'
      } else if (arg.type === 'raw_string_multiline') {
        return 'textarea'
      } else if (arg.choices && arg.choices.length > 0 && (arg.type === 'select' || arg.type === '')) {
        return 'select'
      } else {
        return 'input'
      }
    },
    
    getInputType(arg) {
      if (arg.type === 'html' || arg.type === 'raw_string_multiline' || arg.type === 'select') {
        return undefined
      }
      return arg.type
    },
    
    getPattern(arg) {
      if (arg.type && arg.type.startsWith('regex:')) {
        return arg.type.replace('regex:', '')
      }
      return undefined
    },
    
    getArgumentValue(arg) {
      if (arg.type === 'checkbox') {
        return this.argValues[arg.name] === '1' || this.argValues[arg.name] === true
      }
      return this.argValues[arg.name] || ''
    },
    
    handleInput(arg, event) {
      const value = event.target.type === 'checkbox' ? event.target.checked : event.target.value
      this.argValues[arg.name] = value
      this.updateUrlWithArg(arg.name, value)
    },
    
    handleChange(arg, event) {
      if (arg.type === 'confirmation') {
        this.confirmationChecked = event.target.checked
        return
      }
      
      // Validate the input
      this.validateArgument(arg, event.target.value)
    },
    
    async validateArgument(arg, value) {
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
          this.$delete(this.formErrors, arg.name)
        } else {
          this.$set(this.formErrors, arg.name, validation.description)
        }
      } catch (err) {
        console.warn('Validation failed:', err)
      }
    },
    
    updateUrlWithArg(name, value) {
      if (name && value !== undefined) {
        const url = new URL(window.location.href)
        
        // Don't add passwords to URL
        const arg = this.arguments.find(a => a.name === name)
        if (arg && arg.type === 'password') {
          return
        }
        
        url.searchParams.set(name, value)
        window.history.replaceState({}, '', url.toString())
      }
    },
    
    getArgumentValues() {
      const ret = []
      
      for (const arg of this.arguments) {
        let value = this.argValues[arg.name] || ''
        
        if (arg.type === 'checkbox') {
          value = value ? '1' : '0'
        }
        
        ret.push({
          name: arg.name,
          value: value
        })
      }
      
      return ret
    },
    
    handleSubmit() {
      // Validate all inputs
      let isValid = true
      
      for (const arg of this.arguments) {
        const value = this.argValues[arg.name]
        if (arg.required && (!value || value === '')) {
          this.$set(this.formErrors, arg.name, 'This field is required')
          isValid = false
        }
      }
      
      if (!isValid) {
        return
      }
      
      const argvs = this.getArgumentValues()
      this.$emit('submit', argvs)
      this.close()
    },
    
    handleCancel() {
      this.clearBookmark()
      this.$emit('cancel')
      this.close()
    },
    
    handleClose() {
      this.$emit('close')
    },
    
    clearBookmark() {
      // Remove the action from the URL
      window.history.replaceState({
        path: window.location.pathname
      }, '', window.location.pathname)
    },
    
    show() {
      this.$refs.dialog.showModal()
    },
    
    close() {
      this.$refs.dialog.close()
    }
  }
}
</script>

<style scoped>
.action-arguments {
  border: none;
  border-radius: 8px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  max-width: 500px;
  width: 90vw;
}

.wrapper {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.action-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #eee;
}

.action-header .icon {
  font-size: 1.5em;
}

.action-header h2 {
  margin: 0;
  font-size: 1.2em;
}

.arguments {
  display: flex;
  flex-direction: column;
  gap: 1rem;
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

.argument-group input,
.argument-group select,
.argument-group textarea {
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
  transition: border-color 0.2s ease;
}

.argument-group input:focus,
.argument-group select:focus,
.argument-group textarea:focus {
  outline: none;
  border-color: #007bff;
  box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
}

.argument-group input:invalid,
.argument-group select:invalid,
.argument-group textarea:invalid {
  border-color: #dc3545;
}

.argument-group textarea {
  resize: vertical;
  min-height: 100px;
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

.buttons button {
  padding: 0.5rem 1rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: #fff;
  cursor: pointer;
  font-size: 1rem;
  transition: all 0.2s ease;
}

.buttons button:hover:not(:disabled) {
  background: #f8f9fa;
  border-color: #adb5bd;
}

.buttons button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.buttons button[name="start"] {
  background: #007bff;
  color: white;
  border-color: #007bff;
}

.buttons button[name="start"]:hover:not(:disabled) {
  background: #0056b3;
  border-color: #0056b3;
}

/* Checkbox specific styling */
.argument-group input[type="checkbox"] {
  width: auto;
  margin-right: 0.5rem;
}

.argument-group input[type="checkbox"] + label {
  display: inline;
  font-weight: normal;
}
</style> 