/* eslint @typescript-eslint/no-var-requires: 0 */
const http = require('http')
const { exec } = require('child_process')

function makeDoCall(path) {
  return () =>
    new Promise((resolve, reject) => {
      http.get('http://127.0.0.1:3033' + path, (res) => {
        if (res.statusCode !== 200) {
          reject(new Error('request failed: ' + res.statusCode))
          return
        }
        resolve(null)
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
module.exports = (on) => {
  on('task', {
    'engine:trigger': makeDoCall('/signal?sig=SIGUSR2'),
    'db:fastforward': fastForwardDB,
    'db:resettime': () => {
      durations = []
      return fastForwardDB()
    },
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
