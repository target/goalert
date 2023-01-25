import { test } from '@playwright/test'
import http from 'http'

const base = 'http://127.0.0.1:6133'
function fetch(path: string): Promise<string> {
  return new Promise((resolve, reject) => {
    http.get(base + path, (res) => {
      if (res.statusCode !== 200) {
        reject(
          new Error(`request failed: ${res.statusCode}; url=${base + path}`),
        )
        return
      }
      let data = ''
      res.on('data', (chunk) => {
        data += chunk
      })
      res.on('end', () => {
        resolve(data)
      })

      res.on('error', (err) => {
        reject(err)
      })
    })
  })
}

export function configureExpFlags(flags: string[]): void {
  test.describe.configure({ mode: 'serial' })

  test.beforeAll(async () => {
    await fetch('/stop')
    await fetch('/start?extra-arg=--experimental&extra-arg=' + flags.join(','))
  })
  test.afterAll(async () => {
    await fetch('/stop')
    await fetch('/start')
  })
}
