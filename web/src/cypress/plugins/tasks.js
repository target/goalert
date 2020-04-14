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

module.exports = {
  'engine:trigger': makeDoCall('/signal?sig=SIGUSR2'),
  'engine:start': makeDoCall('/start'),
  'engine:stop': makeDoCall('/stop'),
}
