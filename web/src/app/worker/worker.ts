import methods from './methods'

self.onmessage = (e) => {
  const methodName = e.data.method as keyof typeof methods

  if (!(methodName in methods)) {
    // Shouldn't happen, but ensures that we can't unknowingly
    // call a method that doesn't exist, or is on the prototype.
    throw new Error('Invalid method')
  }

  const method = methods[methodName]
  if (typeof method !== 'function') {
    throw new Error('Method is not a function')
  }

  const result = method(e.data.arg)
  self.postMessage(result)
}
