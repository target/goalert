import methods from './methods'

self.onmessage = (e) => {
  const method = e.data.method as keyof typeof methods
  const result = methods[method](e.data.arg)
  self.postMessage(result)
}
