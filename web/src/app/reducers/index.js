import { combineReducers } from 'redux'
import auth from './auth'
import main from './main'

export default function createRootReducer() {
  return combineReducers({
    auth, // auth status
    main, // reducer for new user setup flag
  })
}
