/* eslint @typescript-eslint/no-var-requires: 0 */
/* eslint @typescript-eslint/no-require-imports: 0 */
const http = require('http')
const { exec } = require('child_process')

function makeDoCall(pathVal, base = 'http://127.0.0.1:3033') {
  return () =>
    new Promise((resolve, reject) => {
      const path = typeof pathVal === 'function' ? pathVal() : pathVal
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

function pgmocktime(flags) {
  return new Promise((resolve, reject) => {
    exec(
      `go tool pgmocktime -d "$DB" ${flags}`,
      {
        env: {
          PATH: process.env.PATH,
          DB: process.env.CYPRESS_DB_URL,
        },
      },
      (err, stdout, stderr) => {
        if (err) {
          err.message = `${err.message}\n${stderr}`
          reject(err)
          return
        }
        resolve(null)
      },
    )
  })
}

function fastForwardDB(duration) {
  if (!duration) {
    return pgmocktime('--inject --reset')
  }

  return pgmocktime('-a ' + duration)
}
let expFlags = []
let apiOnly = false
function flagsQ() {
  if (expFlags.length === 0) {
    return ''
  }

  return `?extra-arg=--experimental=${expFlags.join(',') + (apiOnly ? '&extra-arg=--api-only' : '')}`
}

let failed = false
module.exports = (on, config) => {
  async function engineCycle() {
    const cycleID = await makeDoCall('/health/engine/cycle', config.baseUrl)()
    await makeDoCall('/signal?sig=SIGUSR2')()
    await makeDoCall('/health/engine?id=' + cycleID, config.baseUrl)()
    return null
  }

  const stopBackend = makeDoCall('/stop')
  const startBackend = makeDoCall(() => '/start' + flagsQ())
  async function fastForward(dur) {
    await fastForwardDB(dur)
    return null
  }

  on('task', {
    'engine:trigger': engineCycle,
    'db:setTimeSpeed': (speed) => pgmocktime('-s ' + speed),
    'db:fastforward': fastForward,
    'db:resettime': () => fastForwardDB(),
    'engine:setexpflags': (flags) => {
      expFlags = flags
      return null
    },
    'engine:setapionly': (val) => {
      apiOnly = val
      return null
    },
    'engine:start': startBackend,
    'engine:stop': stopBackend,
    'check:abort': () => failed,
  })

  on('before:spec', async () => {
    await pgmocktime('--inject --reset')
    await fastForwardDB()
  })

  on('after:spec', (spec, results) => {
    if (results.stats.failures) {
      failed = true
    }
  })
}
