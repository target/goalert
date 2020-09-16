import tasks from './plugins/tasks'

export default (on) => {
  on('task', tasks)
}
