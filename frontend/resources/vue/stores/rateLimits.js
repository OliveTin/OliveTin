import { reactive } from 'vue'

// Store rate limit expiry times by bindingId
// This allows all ActionButton components to reactively update when rate limits change
export const rateLimits = reactive({})
