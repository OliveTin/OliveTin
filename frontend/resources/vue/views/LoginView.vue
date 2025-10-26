<template>
  <Section title="Login to OliveTin" class="small">
    <div class="login-form">
      <div v-if="!hasOAuth && !hasLocalLogin" class="login-disabled">
        <span>This server is not configured with either OAuth, or local users, so you cannot login.</span>
      </div>

      <div v-if="hasOAuth" class="login-oauth2">
        <h3>OAuth Login</h3>
        <div class="oauth-providers">
          <button v-for="provider in oauthProviders" :key="provider.name" class="oauth-button"
            @click="loginWithOAuth(provider)">
            <span v-if="provider.icon" class="provider-icon" v-html="provider.icon"></span>
            <span class="provider-name">Login with {{ provider.name }}</span>
          </button>
        </div>
      </div>

      <div v-if="hasLocalLogin" class="login-local">
        <h3>Local Login</h3>
        <form @submit.prevent="handleLocalLogin" class="local-login-form">
          <div v-if="loginError" class="bad">
            {{ loginError }}
          </div>

          <input id="username" v-model="username" type="text" name="username" autocomplete="username" required placeholder="Username" />
          <input id="password" v-model="password" type="password" name="password" autocomplete="current-password" placeholder="Password"
            required />

          <button type="submit" :disabled="loading" class="login-button">
            {{ loading ? 'Logging in...' : 'Login' }}
          </button>
        </form>
      </div>
    </div>
  </Section>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'

const router = useRouter()

const username = ref('')
const password = ref('')
const loading = ref(false)
const loginError = ref('')
const hasOAuth = ref(false)
const hasLocalLogin = ref(false)
const oauthProviders = ref([])

function loadLoginOptions() {
  // Use the init response data that was loaded in App.vue
  if (window.initResponse) {
    hasOAuth.value = window.initResponse.oAuth2Providers && window.initResponse.oAuth2Providers.length > 0
    hasLocalLogin.value = window.initResponse.authLocalLogin

    if (hasOAuth.value) {
      oauthProviders.value = window.initResponse.oAuth2Providers
    }
  } else {
    console.warn('Init response not available yet, login options will be empty')
  }
}

async function handleLocalLogin() {
  loading.value = true
  loginError.value = ''

  try {
    const response = await window.client.localUserLogin({
      username: username.value,
      password: password.value
    })

    if (response.success) {
      // Re-initialize to get updated user context
      try {
        const initResponse = await window.client.init({})
        window.initResponse = initResponse
        window.initError = false
        window.initErrorMessage = ''
        window.initCompleted = true
      } catch (initErr) {
        console.error('Failed to reinitialize after login:', initErr)
      }
      
      // Redirect to home page on successful login
      router.push('/')
    } else {
      loginError.value = 'Login failed. Please check your credentials.'
    }
  } catch (err) {
    console.error('Login error:', err)
    loginError.value = err.message || 'Network error. Please try again.'
  } finally {
    loading.value = false
  }
}

function loginWithOAuth(provider) {
  // Redirect to OAuth provider
  window.location.href = provider.authUrl
}

onMounted(() => {
  loadLoginOptions()
  
  // Also watch for when init response becomes available
  const stopWatcher = watch(() => window.initResponse, () => {
    loadLoginOptions()
  }, { immediate: true })
})
</script>

<style scoped>
section {
  margin: auto;
}

.login-view {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

form {
  grid-template-columns: 1fr;
  gap: 1em;
}
</style>
