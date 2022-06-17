import _ from 'lodash'

import methods from './methods'

self.onmessage = (e) => {
  const result = methods[e.data.method](e.data.arg)
  self.postMessage(result)
}
