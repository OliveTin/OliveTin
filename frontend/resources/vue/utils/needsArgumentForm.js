import { actionRequiresJustification } from './justificationTemplate.js'

export function needsArgumentForm (action) {
  return (action?.arguments?.length > 0) || actionRequiresJustification(action?.justification)
}
