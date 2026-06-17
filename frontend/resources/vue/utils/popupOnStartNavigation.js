export function shouldSuppressPopupOnStartNavigation (router) {
  const path = router?.currentRoute?.value?.path ?? window.location.pathname
  return path === '/logs' || path === '/logs/queue'
}
