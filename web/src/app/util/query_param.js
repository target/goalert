const quoteRx = s => (s || '').replace(/[.?*+^$[\]\\(){}|-]/g, '\\$&')

export function getParameterByName(name, url = global.location.href) {
  name = name.replace(/[[\]]/g, '\\$&')
  const rx = new RegExp('[?&]' + quoteRx(name) + '(=([^&#]*)|&|#|$)')
  const m = rx.exec(url)
  if (!m) return null
  if (!m[2]) return ''

  return decodeURIComponent(m[2].replace(/\+/g, ' '))
}

// returns hash of all parameters with keys and values
export function getAllParameters(url = global.location.href) {
  // match and select any parameters in the url
  const rx = /[?&](\w+)=(?:([^&#]*)|&|#|$)/

  const queries = {}
  // find the first match
  let m = rx.exec(url)
  while (m) {
    // while we have a match
    url = url.replace(m[0], '')
    queries[m[1]] = decodeURIComponent(m[2].replace(/\+/g, ' '))
    m = rx.exec(url) // find the next match
  }

  return queries
}

// takes in a var name, var value, and optionally a url to read previous params from.
// returns a string of the params and the maintained hash (DOES NOT RETURN THE PATH)
export function setParameterByName(name, value, url = global.location.href) {
  // fetch all current url queries
  const queries = getAllParameters(url)

  // set new value
  queries[name] = encodeURIComponent(value)

  // rebuild the url -- omit the parameter `name` if value is null
  const queryList = Object.keys(queries)
    .sort((a, b) => (a < b ? -1 : 1))
    .filter(i => !(value === null && i === name))
    .map(query => {
      return query + '=' + queries[query]
    })

  // match against anything that is after the # in the address
  const rx = /(#.*)/
  const m = rx.exec(url)
  let hash = ''
  if (m) hash = m[1]
  const newURL = '?' + queryList.join('&') + hash

  return newURL
}

// clears the parameter given from the current url
export function clearParameter(name, url = global.location.href) {
  const query = setParameterByName(name, null, url)
  return query
}
