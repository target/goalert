/* eslint @typescript-eslint/no-var-requires: 0 */
const http = require('http')
const { exec } = require('child_process')

function makeDoCall(path, base = 'http://127.0.0.1:3033') {
  return () =>
    new Promise((resolve, reject) => {
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

function execQuery(query) {
  return new Promise((resolve, reject) => {
    exec(
      `psql-lite -d "$DB" -c "$QUERY"`,
      {
        env: {
          PATH: process.env.PATH,
          DB: process.env.CYPRESS_DB_URL,
          QUERY: query,
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

let durations = []

function fastForwardDB(duration) {
  if (duration) durations.push(duration)

  const durStr = durations.map((d) => `'${d}'::interval`).join(' + ')

  const query = `
    create schema if not exists testing_overrides;
    alter database postgres set search_path = "$user", public, testing_overrides, pg_catalog;
		create or replace function testing_overrides.now()
		returns timestamp with time zone
		as $$
			begin
			return (pg_catalog.now()${durStr ? ` + ${durStr}` : ''});
			end;
		$$ language plpgsql;
  `

  return execQuery(query)
}

let failed = false
module.exports = (on, config) => {
  async function engineCycle() {
    const cycleID = await makeDoCall('/health/engine/cycle', config.baseUrl)()
    await makeDoCall('/signal?sig=SIGUSR2')()
    await makeDoCall('/health/engine?id=' + cycleID, config.baseUrl)()
    return null
  }
  async function fastForward(dur) {
    await fastForwardDB(dur)
    await engineCycle()
    return null
  }

  on('task', {
    'engine:trigger': engineCycle,
    'db:fastforward': fastForward,
    'db:resettime': () => fastForwardDB(),
    'engine:start': makeDoCall('/start'),
    'engine:stop': makeDoCall('/stop'),
    'check:abort': () => failed,
  })

  on('after:spec', (spec, results) => {
    if (results.stats.failures) {
      failed = true
    }
  })
}
