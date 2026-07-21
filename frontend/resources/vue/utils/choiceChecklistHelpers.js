function parseLegacyChecklistValue (value) {
  return value.split(',').map((segment) => segment.trim()).filter((segment) => segment !== '')
}

export function parseChecklistValue (value) {
  if (!value || value === '') {
    return []
  }

  const trimmed = value.trim()
  if (trimmed.startsWith('[')) {
    try {
      const parsed = JSON.parse(trimmed)
      if (!Array.isArray(parsed)) {
        return []
      }

      return parsed.map((segment) => String(segment).trim()).filter((segment) => segment !== '')
    } catch {
      return []
    }
  }

  return parseLegacyChecklistValue(value)
}

export function formatChecklistValue (selected) {
  if (!Array.isArray(selected) || selected.length === 0) {
    return ''
  }

  return JSON.stringify(selected)
}

export function toggleChoice (selected, value) {
  const current = Array.isArray(selected) ? [...selected] : []
  const index = current.indexOf(value)

  if (index === -1) {
    current.push(value)
    return current
  }

  current.splice(index, 1)
  return current
}

export function choiceLabel (choice) {
  return choice.title || choice.value || ''
}

export function allChoiceValues (choices) {
  return choices.map((choice) => choice.value)
}
