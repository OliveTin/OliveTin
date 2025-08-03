<script setup>
  setup () {
    const tpl = document.getElementById('tplLoginForm')
    this.content = tpl.content.cloneNode(true)

    this.appendChild(this.content)

    this.querySelector('#local-user-login').addEventListener('submit', (e) => {
      e.preventDefault()
      this.localLoginRequest()
    })
  }

  async localLoginRequest () {
    const username = this.querySelector('input.username').value
    const password = this.querySelector('input.password').value

    document.querySelector('.error').innerHTML = ''

    const args = {
      username: username,
      password: password
    }

    const loginResult = await window.client.localUserLogin(args)

    if (loginResult.success) {
      window.location.href = '/'
    } else {
      document.querySelector('.error').innerHTML = 'Login failed.'
    }
  }

  processOAuth2Providers (providers) {
    if (providers === null) {
      return
    }

    if (providers.length > 0) {
      this.querySelector('.login-oauth2').hidden = false
      this.querySelector('.login-disabled').hidden = true

      for (const provider of providers) {
        const providerForm = document.createElement('form')
        providerForm.method = 'GET'
        providerForm.action = '/oauth/login'

        const hiddenField = document.createElement('input')
        hiddenField.type = 'hidden'
        hiddenField.name = 'provider'
        hiddenField.value = provider.Name

        providerForm.appendChild(hiddenField)

        const providerButton = document.createElement('button')
        providerButton.type = 'submit'
        providerButton.innerHTML = '<span class = "oauth2-icon">' + provider.Icon + '</span> Login with ' + provider.Title

        providerForm.appendChild(providerButton)

        this.querySelector('.login-oauth2').appendChild(providerForm)
      }
    }
  }

  processLocalLogin (enabled) {
    if (enabled) {
      this.querySelector('.login-local').hidden = false
      this.querySelector('.login-disabled').hidden = true
    }
  }
</script>