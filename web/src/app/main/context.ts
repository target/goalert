import React from 'react'

export const AppContext = React.createContext({
  theme: '',
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  setTheme: (theme: string): void => {},
})
AppContext.displayName = 'AppContext'
