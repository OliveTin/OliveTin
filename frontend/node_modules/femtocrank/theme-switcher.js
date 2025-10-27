const darkStylesheet = document.getElementById("dark-theme");
const themeSwitcherButton = document.createElement("input");
themeSwitcherButton.type = "checkbox";
themeSwitcherButton.onclick = () => {
  if (darkStylesheet.getAttribute('media') === 'none') {
    darkStylesheet.setAttribute('media', 'all')
  } else {
    darkStylesheet.setAttribute('media', 'none');
  }
}

const label = document.createElement("label");
label.innerText = "Dark Theme";
label.appendChild(themeSwitcherButton);

document.body.getElementsByTagName("header")[0].appendChild(label);
