export function needsArgumentForm (action) {
  return (action?.arguments?.length > 0) || action?.justification
}
