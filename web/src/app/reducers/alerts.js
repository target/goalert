import {
  SET_ALERTS_CHECKED,
  SET_ALERTS_ACTION_COMPLETE,
  SET_ALERTS,
} from '../actions'

const initialState = () => {
  return {
    actionComplete: false,
    alerts: [],
    checkedAlerts: [],
  }
}

/*
 * Updates state depending on what action type given
 *
 * Returns the immutable final state afterwards (reduce)
 */
export default function alertsReducer(state = initialState(), action) {
  switch (action.type) {
    case SET_ALERTS_CHECKED: {
      return {
        ...state,
        checkedAlerts: action.payload,
      }
    }
    case SET_ALERTS_ACTION_COMPLETE: {
      return {
        ...state,
        actionComplete: action.payload,
      }
    }
    case SET_ALERTS: {
      return {
        ...state,
        alerts: action.payload,
      }
    }
  }

  return state
}
