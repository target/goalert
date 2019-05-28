import { SET_SHOW_NEW_USER_FORM } from '../actions'
import { getParameterByName } from '../util/query_param'

const initialState = {
  isFirstLogin: getParameterByName('isFirstLogin') === '1',
}

/*
 * Updates state depending on what action type given
 *
 * Returns the immutable final state afterwards (reduce)
 */
export default function mainReducer(state = initialState, action = {}) {
  switch (action.type) {
    case SET_SHOW_NEW_USER_FORM:
      return {
        ...state,
        isFirstLogin: action.payload,
      }
    default:
      return state
  }
}
