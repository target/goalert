import plugin from 'cypress-plugin-retries/lib/plugin'
import fs from 'fs'

export default on => {
  plugin(on)
  on('task', {
    'engine:trigger': () => {
      const data = fs.readFileSync('backend.pid')
      process.kill(parseInt(data.toString(), 10), 'SIGUSR2')
      return null
    },
  })
}
