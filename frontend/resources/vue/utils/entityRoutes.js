export function entityDetailsRoute (entity) {
  return {
    name: 'EntityDetails',
    params: {
      entityType: entity.type,
      entityKey: entity.uniqueKey
    }
  }
}
