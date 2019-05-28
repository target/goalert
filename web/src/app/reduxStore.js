import thunk from 'redux-thunk'
import createRootReducer from './reducers'
import { composeWithDevTools } from 'redux-devtools-extension'
import { applyMiddleware, createStore } from 'redux'
import { routerMiddleware } from 'connected-react-router'
import history from './history'

export default createStore(
  createRootReducer(history),
  composeWithDevTools({})(
    applyMiddleware(
      thunk,
      routerMiddleware(history), // for dispatching history actions
    ),
  ),
)
