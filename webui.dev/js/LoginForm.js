export class LoginForm extends window.HTMLElement {
  setup () {
    const tpl = document.getElementById('tplLoginForm')
    this.content = tpl.content.cloneNode(true)

    this.appendChild(this.content)

    this.querySelector('#local-user-login').addEventListener('submit', (e) => {
      e.preventDefault()
      this.localLoginRequest()
    })
  }

  localLoginRequest () {
    const username = this.querySelector('input.username').value
    const password = this.querySelector('input.password').value

    document.querySelector('.error').innerHTML = ''

    window.fetch('/api/LocalUserLogin', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        username: username,
        password: password
      })
    }).then((response) => {
      if (response.ok) {
        const res = response.json()

        if (res !== undefined) {
          return res
        } else {
          throw new Error('Failed to login - no res')
        }
      } else {
        throw new Error('Failed to login')
      }
    }).then((res) => {
      if (res.success) {
        window.location.reload()
      } else {
        document.querySelector('.error').innerHTML = 'Login failed.'
      }
    })
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
        providerForm.action = '/oauth2?provider=' + provider.Name

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
}

window.customElements.define('login-form', LoginForm)
