const myPassword = 'sekrit'

const domMain = document.getElementsByTagName('main')[0]
domMain.style.display = 'none'

const domPassword = document.createElement('input')
const domLogin = document.createElement('button')

function checkPassword () {
  if (domPassword.value === myPassword) {
    domMain.style.display = 'block'
    domPassword.remove()
    domLogin.remove()
  } else {
    window.alert('Incorrect password. Please try again.')
  }
}

function setupPasswordForm () {
  domPassword.setAttribute('type', 'password')
  domPassword.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') {
      checkPassword()
    }
  })

  domLogin.innerText = 'Login'
  domLogin.onclick = checkPassword

  const domHeader = document.querySelector('header')
  domHeader.appendChild(domPassword)
  domHeader.appendChild(domLogin)
}

document.addEventListener('DOMContentLoaded', setupPasswordForm)
