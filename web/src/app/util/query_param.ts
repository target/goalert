export function getParameterByName(
  name: string,
  url: string = global.location.href,
): string | null {
  return new URL(url).searchParams.get(name)
}
