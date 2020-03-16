import { createBrowserHistory } from 'history'
import { pathPrefix } from './env'

export default createBrowserHistory({ basename: pathPrefix })
