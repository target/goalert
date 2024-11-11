import { check } from 'k6'
import http from 'k6/http'

const TEST_USER = 'admin'
const TEST_PASS = 'admin123'

// login will return the session token
export function login(
  host = 'http://localhost:3030',
  user = TEST_USER,
  pass = TEST_PASS,
): string {
  let res = http.post(
    host + '/api/v2/identity/providers/basic?noRedirect=1',
    {
      username: user,
      password: pass,
    },
    {
      headers: {
        referer: host,
      },
    },
  )

  if (res.status !== 200) {
    throw new Error(`Unexpected status code: ${res.status}\n` + res.body)
  }

  return res.body as string
}
