// Role-prefixed IDs keep related elements from colliding when argument names
// contain another argument's name plus a former suffix (e.g. "foo" vs "foo-choices").
const ARGUMENT_ID_NAMESPACE = 'arg-'

export const ARGUMENT_FIELD_ID_PREFIX = `${ARGUMENT_ID_NAMESPACE}field-`

export function argumentFieldId (argumentName) {
  return `${ARGUMENT_FIELD_ID_PREFIX}${argumentName}`
}

export function argumentFieldChoicesId (argumentName) {
  return `${ARGUMENT_ID_NAMESPACE}choices-${argumentName}`
}

export function argumentFieldValueId (argumentName) {
  return `${ARGUMENT_ID_NAMESPACE}value-${argumentName}`
}

export function argumentFieldWrapperId (argumentName) {
  return `${ARGUMENT_ID_NAMESPACE}wrapper-${argumentName}`
}

export function argumentFieldOptionId (argumentName, index) {
  return `${ARGUMENT_ID_NAMESPACE}option-${argumentName}-${index}`
}

export function argumentFieldListboxId (argumentName) {
  return `${ARGUMENT_ID_NAMESPACE}listbox-${argumentName}`
}

export function argumentFieldListboxOptionId (argumentName, index) {
  return `${ARGUMENT_ID_NAMESPACE}listbox-option-${argumentName}-${index}`
}

export function argumentFieldValidationElementId (argument) {
  if (argument.type === 'checklist') {
    return argumentFieldValueId(argument.name)
  }

  return argumentFieldId(argument.name)
}
