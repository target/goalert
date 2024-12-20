import { thunk } from 'redux-thunk'
import createRootReducer from './reducers'
import { composeWithDevTools } from 'redux-devtools-extension'
import { applyMiddleware, createStore } from 'redux'

export default createStore(
  createRootReducer(),
  composeWithDevTools({})(applyMiddleware(thunk)),
)
