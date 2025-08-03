export function getVersionMacro () {
  const type = process.env.GITHUB_REF_TYPE;
  const name = process.env.GITHUB_REF_NAME;

  if (type === 'tag') {
    return name;
  }

  return 'dev';
}
