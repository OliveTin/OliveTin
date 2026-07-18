export function applyArgumentTemplate (template, args) {
  if (!template) {
    return ''
  }

  return template.replace(/\{\{\s*([a-zA-Z0-9_]+)\s*\}\}/g, (_, name) => args[name] ?? '')
}

export function actionRequiresJustification (justification) {
  return (justification ?? '').length > 0
}

export function actionJustificationTemplate (justification) {
  return justification ?? ''
}
