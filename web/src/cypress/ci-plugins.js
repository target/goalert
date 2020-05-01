import plugin from 'cypress-plugin-retries/lib/plugin'
import tasks from './plugins/tasks'

export default (on) => {
  plugin(on)
  on('task', tasks)
}
