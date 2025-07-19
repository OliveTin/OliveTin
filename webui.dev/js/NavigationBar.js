export class NavigationBar {
  constructor () {
    this.navbar = document.getElementsByTagName('nav')[0]
    this.mainLinks = document.getElementById('navigation-links')
    this.supplementalLinks = document.getElementById('supplemental-links')
  }

  createLink (title, url, isSupplemental) {
    const parent = (isSupplemental) ? this.supplementalLinks : this.mainLinks

    const existsAlready = Array.from(parent.querySelectorAll('li')).some(el => el.title === title)

    if (existsAlready) {
      return
    }

    const linkA = document.createElement('a')
    linkA.href = url
    linkA.innerText = title

    const navigationLi = document.createElement('li')
    navigationLi.appendChild(linkA)
    navigationLi.title = title

    parent.appendChild(navigationLi)
  }

  refreshSectionPolicyLinks (policy) {
    if (policy.showDiagnostics) {
      this.createLink('Diagnostics', '/diagnostics', true)
    }

    if (policy.showLogList) {
      this.createLink('Logs', '/logs', true)
    }
  }
}
