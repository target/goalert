// TODO: move to ../reducers and define rest of state
interface ReduxState {
  auth: {
    valid: boolean
  }
}

export const authSelector = (state: ReduxState): boolean => state.auth.valid
