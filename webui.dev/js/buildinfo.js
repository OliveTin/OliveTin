export function getVersionMacro() {
  return process.env.GITHUB_REF_NAME || 'dev'
}
