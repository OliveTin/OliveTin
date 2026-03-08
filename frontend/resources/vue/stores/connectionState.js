import { reactive } from 'vue'

export const connectionState = reactive({
  connected: false,
  reconnecting: false
})
