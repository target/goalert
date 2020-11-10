interface ReduxState {
  valid: boolean
}

export const authSelector = (state: ReduxState): boolean => state.valid
