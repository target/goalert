import methods from './methods'

self.onmessage = (e) => {
  const method = e.data.method as keyof typeof methods

  if (!(method in methods)) {
    // Shouldn't happen, but ensures that we can't unknowingly
    // call a method that doesn't exist, or is on the prototype.
    throw new Error('Invalid method')
  }

  const result = methods[method](e.data.arg)
  self.postMessage(result)
}
