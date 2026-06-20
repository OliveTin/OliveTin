import { reactive } from 'vue'

export const connectionState = reactive({
  connected: false,
  reconnecting: false,
  disconnectedAt: null,
  nextReconnectAt: null,
  scheduledReconnectDelayMs: 0,
  showDisconnectedBanner: false
})
