<template>
  <Section title="Login to OliveTin" class="small">
    <div class="login-form" style="display: grid; grid-template-columns: max-content 1fr; gap: 1em;">
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
          <div v-if="loginError" class="error-message">
            {{ loginError }}
          </div>

          <label for="username">Username:</label>
          <input id="username" v-model="username" type="text" name="username" autocomplete="username" required />

          <label for="password">Password:</label>
          <input id="password" v-model="password" type="password" name="password" autocomplete="current-password"
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
import { ref, onMounted } from 'vue'
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

async function fetchLoginOptions() {
  try {
    const response = await fetch('webUiSettings.json')
    const settings = await response.json()

    hasOAuth.value = settings.AuthOAuth2Providers && settings.AuthOAuth2Providers.length > 0
    hasLocalLogin.value = settings.AuthLocalLogin

    if (hasOAuth.value) {
      oauthProviders.value = settings.AuthOAuth2Providers
    }
  } catch (err) {
    console.error('Failed to fetch login options:', err)
  }
}

async function handleLocalLogin() {
  loading.value = true
  loginError.value = ''

  try {
    const response = await fetch('/api/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        username: username.value,
        password: password.value
      })
    })

    if (response.ok) {
      // Redirect to home page on successful login
      router.push('/')
    } else {
      const error = await response.text()
      loginError.value = error || 'Login failed. Please check your credentials.'
    }
  } catch (err) {
    console.error('Login error:', err)
    loginError.value = 'Network error. Please try again.'
  } finally {
    loading.value = false
  }
}

function loginWithOAuth(provider) {
  // Redirect to OAuth provider
  window.location.href = provider.authUrl
}

onMounted(() => {
  fetchLoginOptions()
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
  grid-template-columns: max-content 1fr;
  gap: 1em;
}
</style>
