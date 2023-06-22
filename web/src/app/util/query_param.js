export function getParameterByName(name, url = global.location.href) {
  return new URL(url).searchParams.get(name)
}

// returns hash of all parameters with keys and values
export function getAllParameters(url = global.location.href) {
  const q = {}
  for (const [key, value] of new URL(url).searchParams) {
    q[key] = value
  }

  return q
}

// takes in a var name, var value, and optionally a url to read previous params from.
// returns a string of the params and the maintained hash (DOES NOT RETURN THE PATH)
export function setParameterByName(name, value, url = global.location.href) {
  const u = new URL(url)
  u.searchParams.set(name, value)
  return u.toString()
}

// clears the parameter given from the current url
export function clearParameter(name, url = global.location.href) {
  const query = setParameterByName(name, null, url)
  return query
}
