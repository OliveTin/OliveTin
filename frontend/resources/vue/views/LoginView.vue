<template>
  <section class = "small">
    <div class="login-container">
      <div class="login-form" style="display: grid; grid-template-columns: max-content 1fr; gap: 1em;">
        <h2>Login to OliveTin</h2>

        <div v-if="!hasOAuth && !hasLocalLogin" class="login-disabled">
          <p>This server is not configured with either OAuth, or local users, so you cannot login.</p>
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
    </div>
  </section>
</template>

<script>
export default {
  name: 'LoginView',
  data() {
    return {
      username: '',
      password: '',
      loading: false,
      loginError: '',
      hasOAuth: false,
      hasLocalLogin: false,
      oauthProviders: []
    }
  },
  mounted() {
    this.fetchLoginOptions()
  },
  methods: {
    async fetchLoginOptions() {
      try {
        const response = await fetch('webUiSettings.json')
        const settings = await response.json()

        this.hasOAuth = settings.AuthOAuth2Providers && settings.AuthOAuth2Providers.length > 0
        this.hasLocalLogin = settings.AuthLocalLogin

        if (this.hasOAuth) {
          this.oauthProviders = settings.AuthOAuth2Providers
        }
      } catch (err) {
        console.error('Failed to fetch login options:', err)
      }
    },

    async handleLocalLogin() {
      this.loading = true
      this.loginError = ''

      try {
        const response = await fetch('/api/login', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            username: this.username,
            password: this.password
          })
        })

        if (response.ok) {
          // Redirect to home page on successful login
          this.$router.push('/')
        } else {
          const error = await response.text()
          this.loginError = error || 'Login failed. Please check your credentials.'
        }
      } catch (err) {
        console.error('Login error:', err)
        this.loginError = 'Network error. Please try again.'
      } finally {
        this.loading = false
      }
    },

    loginWithOAuth(provider) {
      // Redirect to OAuth provider
      window.location.href = provider.authUrl
    }
  }
}
</script>

<style scoped>
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