export function findSelectedChoice (choices, modelValue) {
  if (!modelValue) {
    return null
  }

  return choices.find(choice => choice.value === modelValue) ?? null
}

export function choiceDisplayLabel (choice) {
  return choice.title || choice.value
}

export function displayLabelForModelValue (choices, modelValue) {
  const match = findSelectedChoice(choices, modelValue)
  return match ? choiceDisplayLabel(match) : ''
}

export function normalizedModelValue (choices, modelValue) {
  if (!modelValue) {
    return ''
  }

  return findSelectedChoice(choices, modelValue) ? modelValue : ''
}

export function syncStateFromModelValue (choices, modelValue) {
  const normalizedValue = normalizedModelValue(choices, modelValue)

  return {
    query: displayLabelForModelValue(choices, normalizedValue),
    modelValue: normalizedValue
  }
}
