window.logs = []

export function addToLog (evt) {
	window.logs.append(evt)

	showLog(evt)
}

function showLog (evt) {
	let msg = document.createElement('pre')
	msg.innerText = evt;

	document.body.appendChild(msg)
}
