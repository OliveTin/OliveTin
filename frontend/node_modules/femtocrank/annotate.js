
//main()

function hex(x) {
  var hexDigits = new Array("0","1","2","3","4","5","6","7","8","9","a","b","c","d","e","f");

  return isNaN(x) ? "00" : hexDigits[(x - x % 16) / 16] + hexDigits[x % 16];
 }

function rgb2hex(rgb) {
 console.log(rgb)
 rgb = rgb.match(/^rgba?\((\d+),\s*(\d+),\s*(\d+)/);

 return "#" + hex(rgb[1]) + hex(rgb[2]) + hex(rgb[3]);
}


function getHex(style, prop) {
  const rgb = style.getPropertyValue(prop)

  return rgb2hex(rgb);
}

function main() {
  for (const el of document.getElementsByClassName("show")) {
    const lbl = document.createElement('strong')
    const p = document.createElement('p')
    p.classList.add('annotation')
    
    lbl.innerText = el.tagName + " " 
    p.append(lbl)

    const style = window.getComputedStyle(el, null)

    const bgColor = document.createElement('span')
    bgColor.innerText = 'bg: ' + getHex(style, 'background-color') + "  "
    p.appendChild(bgColor)

    const fgColor = document.createElement('span')
    fgColor.innerText = 'fg: ' + getHex(style, 'color')
    p.appendChild(fgColor)

    el.appendChild(p)
  }
}
