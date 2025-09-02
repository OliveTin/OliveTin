<template>
  <div 
    :id="`execution-${executionTrackingId}`"
    class="execution-button"
  >
    <button
      :title="`${ellapsed}s`"
      @click="show"
    >
      {{ buttonText }}
    </button>
  </div>
</template>

<script>
//import { ExecutionFeedbackButton } from '../js/ExecutionFeedbackButton.js'

export default {
  name: 'ExecutionButton',
//  mixins: [ExecutionFeedbackButton],
  props: {
    executionTrackingId: {
      type: String,
      required: true
    }
  },
  data() {
    return {
      ellapsed: 0,
      isWaiting: true
    }
  },
  computed: {
    buttonText() {
      if (this.isWaiting) {
        return 'Executing...'
      } else {
        return `${this.ellapsed}s`
      }
    }
  },
  mounted() {
    this.constructFromJson(this.executionTrackingId)
  },
  methods: {
    constructFromJson(json) {
      this.executionTrackingId = json
      this.ellapsed = 0
      this.isWaiting = true
    },
    
    show() {
      this.$emit('show')
      
      if (window.executionDialog) {
        window.executionDialog.reset()
        window.executionDialog.show()
        window.executionDialog.fetchExecutionResult(this.executionTrackingId)
      }
    },
    
    onExecStatusChanged() {
      this.isWaiting = false
      this.domTitle = this.ellapsed + 's'
    },
    
    // Override from ExecutionFeedbackButton
    onExecutionFinished(logEntry) {
      if (logEntry.timedOut) {
        this.renderExecutionResult('action-timeout', 'Timed out')
      } else if (logEntry.blocked) {
        this.renderExecutionResult('action-blocked', 'Blocked!')
      } else if (logEntry.exitCode !== 0) {
        this.renderExecutionResult('action-nonzero-exit', 'Exit code ' + logEntry.exitCode)
      } else {
        this.ellapsed = Math.ceil(new Date(logEntry.datetimeFinished) - new Date(logEntry.datetimeStarted)) / 1000
        this.renderExecutionResult('action-success', 'Success!')
      }
    },
    
    renderExecutionResult(resultCssClass, temporaryStatusMessage) {
      this.updateDom(resultCssClass, '[' + temporaryStatusMessage + ']')
      this.onExecStatusChanged()
    },
    
    updateDom(resultCssClass, title) {
      // For execution button, we don't need to update classes as much
      // since it's a simpler component
      if (resultCssClass) {
        this.$el.classList.add(resultCssClass)
      }
    }
  }
}
</script>

<style scoped>
.execution-button {
  display: inline-block;
}

.execution-button button {
  padding: 0.25em 0.5em;
  border: 1px solid #ccc;
  border-radius: 3px;
  background: #fff;
  cursor: pointer;
  font-size: 0.9em;
  transition: all 0.2s ease;
}

.execution-button button:hover {
  background: #f5f5f5;
  border-color: #999;
}

/* Animation classes */
.execution-button button.action-timeout {
  background: #fff3cd;
  border-color: #ffeaa7;
  color: #856404;
}

.execution-button button.action-blocked {
  background: #f8d7da;
  border-color: #f5c6cb;
  color: #721c24;
}

.execution-button button.action-nonzero-exit {
  background: #f8d7da;
  border-color: #f5c6cb;
  color: #721c24;
}

.execution-button button.action-success {
  background: #d4edda;
  border-color: #c3e6cb;
  color: #155724;
}
</style> 