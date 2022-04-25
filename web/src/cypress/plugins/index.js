/* eslint @typescript-eslint/no-var-requires: 0 */
const http = require('http')

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

let failed = false
module.exports = (on) => {
  on('task', {
    'engine:trigger': makeDoCall('/signal?sig=SIGUSR2'),
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
