<template>
  <Section title="User Information" class="small">
    <div v-if="!isLoggedIn" class="user-not-logged-in">
      <p>You are not currently logged in.</p>
      <p>To access user settings and logout, please <router-link to="/login">log in</router-link>.</p>
    </div>

    <div v-else class="user-control-panel">
      <dl class="user-info">
        <dt>Username</dt>
        <dd>{{ username }}</dd>
        <dt v-if="userProvider !== 'system'">Provider</dt>
        <dd v-if="userProvider !== 'system'">{{ userProvider }}</dd>
        <dt v-if="usergroup">Group</dt>
        <dd v-if="usergroup">{{ usergroup }}</dd>
      </dl>

      <div class="user-actions">
        <div class="action-buttons">
          <button @click="handleLogout" class="button bad" :disabled="loggingOut">
            {{ loggingOut ? 'Logging out...' : 'Logout' }}
          </button>
        </div>
      </div>
    </div>
  </Section>
</template>

<script setup>
import { ref, onMounted, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'

const router = useRouter()

const isLoggedIn = ref(false)
const username = ref('guest')
const userProvider = ref('system')
const usergroup = ref('')
const loggingOut = ref(false)

function updateUserInfo() {
  if (window.initResponse) {
    isLoggedIn.value = window.initResponse.authenticatedUser !== '' && window.initResponse.authenticatedUser !== 'guest'
    username.value = window.initResponse.authenticatedUser
    userProvider.value = window.initResponse.authenticatedUserProvider || 'system'
    usergroup.value = window.initResponse.effectivePolicy?.usergroup || ''
  }
}

async function handleLogout() {
  loggingOut.value = true
  
  try {
    await window.client.logout({})
    
    // Re-initialize to get updated user context (should be guest)
    try {
      const initResponse = await window.client.init({})
      window.initResponse = initResponse
      window.initError = false
      window.initErrorMessage = ''
      window.initCompleted = true
      
      // Update the header with new user info
      if (window.updateHeaderFromInit) {
        window.updateHeaderFromInit()
      }
    } catch (initErr) {
      console.error('Failed to reinitialize after logout:', initErr)
    }
    
    // Redirect to home page
    router.push('/')
  } catch (err) {
    console.error('Logout error:', err)
  } finally {
    loggingOut.value = false
  }
}

let watchInterval = null

onMounted(() => {
  updateUserInfo()
  
  // Watch for changes to init response
  watchInterval = setInterval(() => {
    if (window.initResponse) {
      updateUserInfo()
    }
  }, 1000)
})

onUnmounted(() => {
  if (watchInterval) {
    clearInterval(watchInterval)
  }
})
</script>

<style scoped>
section {
  margin: auto;
}

.user-not-logged-in {
  padding: 2rem;
  text-align: center;
}

.user-not-logged-in p {
  margin: 1rem 0;
}

.user-control-panel {
  display: grid;
  grid-template-columns: 1fr;
  gap: 2rem;
}

.action-buttons {
  display: flex;
  gap: 1rem;
}

.button {
  padding: 0.75rem 1.5rem;
  border-radius: 4px;
  border: none;
  cursor: pointer;
  text-align: center;
  font-weight: 500;
  transition: background-color 0.2s;
}

.button.bad {
  background-color: #dc3545;
  color: white;
}

.button.bad:hover:not(:disabled) {
  background-color: #c82333;
}

.button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
