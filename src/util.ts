export function isRelevantRequest(url: string): boolean {
  let pathname: string

  try {
    const parsedUrl = new URL(url)
    pathname = parsedUrl.pathname
    url = parsedUrl.toString()
  } catch (_) {
    return false
  }
  console.log('isRelevantRequest', url)
  if (!url.startsWith('wss://wss-primary.slack.com')) {
    return false
  }

  return true
}