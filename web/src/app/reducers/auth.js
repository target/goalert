import { AUTH_LOGOUT } from '../actions/auth'

const initialState = () => ({
  valid: true,
})

export default function authReducer(state = initialState(), action) {
  switch (action.type) {
    case AUTH_LOGOUT:
      return { ...state, valid: false }
  }

  return state
}
