const typesWithoutInputType = ['html', 'raw_string_multiline', 'select']

const htmlInputTypesByArgumentType = {
  checkbox: 'checkbox',
  confirmation: 'checkbox',
  datetime: 'datetime-local',
  email: 'email',
  int: 'number',
  password: 'password',
  url: 'url'
}

// Argument types are OliveTin validation types, not HTML input types; anything
// without a matching HTML type renders as text so that it is styled and the
// browser does not have to guess.
export function argumentInputType (arg) {
  if (typesWithoutInputType.includes(arg.type)) {
    return undefined
  }

  return htmlInputTypesByArgumentType[arg.type] ?? 'text'
}
