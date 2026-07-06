export function parseChecklistValue(value) {
  if (!value || value === '') {
    return []
  }

  return value.split(',').map((segment) => segment.trim()).filter((segment) => segment !== '')
}

export function formatChecklistValue(selected) {
  if (!Array.isArray(selected) || selected.length === 0) {
    return ''
  }

  return selected.join(',')
}

export function toggleChoice(selected, value) {
  const current = Array.isArray(selected) ? [...selected] : []
  const index = current.indexOf(value)

  if (index === -1) {
    current.push(value)
    return current
  }

  current.splice(index, 1)
  return current
}

export function choiceLabel(choice) {
  return choice.title || choice.value || ''
}

export function allChoiceValues(choices) {
  return choices.map((choice) => choice.value)
}
