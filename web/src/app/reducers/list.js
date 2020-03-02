import { SET_CHECKED_LIST_ITEMS } from '../actions/list'

const initialState = () => ({
  checkedItems: [],
})

export default function listReducer(state = initialState(), action) {
  switch (action.type) {
    case SET_CHECKED_LIST_ITEMS:
      return {
        ...state,
        checkedItems: action.payload,
      }
  }

  return state
}
