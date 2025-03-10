import { AUTH_LOGOUT } from '../actions/auth'
import { Action, State } from './main'

const initialState = (): State => ({
  valid: true,
})

export default function authReducer(
  state = initialState(),
  action: Action,
): State {
  switch (action.type) {
    case AUTH_LOGOUT:
      return { ...state, valid: false }
  }

  return state
}
