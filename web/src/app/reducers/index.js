import { combineReducers } from 'redux'
import auth from './auth'
import main from './main'
import { connectRouter } from 'connected-react-router'

export default history =>
  combineReducers({
    router: connectRouter(history),

    auth, // auth status
    main, // reducer for new user setup flag
  })
