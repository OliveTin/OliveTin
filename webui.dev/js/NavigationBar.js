export class NavigationBar {
	constructor() {
		this.navbar = document.getElementsByTagName('nav')[0]
		this.mainLinks = document.getElementById('navigation-links')
		this.supplementalLinks = document.getElementById('supplemental-links')
	}

	createLink(title, url, isSupplemental) {
		const linkA = document.createElement('a')
		linkA.href = url
		linkA.innerText = title

		const navigationLi = document.createElement('li')
		navigationLi.appendChild(linkA)
		navigationLi.title = title

		if (isSupplemental) {
			this.supplementalLinks.appendChild(navigationLi)
		} else {
			this.mainLinks.appendChild(navigationLi)
		}
	}

	refreshSectionPolicyLinks(policy) {
		if (policy.showDiagnostics) {
			this.createLink('Diagnostics', '/diagnostics', true)
		}

		if (policy.showLogList) {
			this.createLink('Logs', '/logs', true)
		}
	}
}
