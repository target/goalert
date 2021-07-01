import React, { useContext } from 'react'
import IconButton from '@material-ui/core/IconButton'
import DarkModeIcon from '@material-ui/icons/Brightness4'
import LightModeIcon from '@material-ui/icons/BrightnessHigh'
import { AppContext } from '../context'

export default function ThemeButton(): JSX.Element {
  const { theme, setTheme } = useContext(AppContext)

  return (
    <IconButton onClick={() => setTheme(theme)}>
      {theme === 'dark' ? <DarkModeIcon /> : <LightModeIcon />}
    </IconButton>
  )
}

// function getThemePreference(): string {
//   const theme = localStorage.getItem('theme')
//   if (!theme) {
//     const mq = window.matchMedia('(prefers-color-scheme: dark)')
//     return mq.matches ? 'dark' : 'light'
//   }
//   return theme
// }

// export default function ThemeProvider(): JSX.Element {
//   const [theme, setTheme] = useState(getThemePreference())

//   useEffect(() => {
//       localStorage.setItem('theme', theme)
//   }, [theme])

//   if (!theme) {
//     const mq = window.matchMedia('(prefers-color-scheme: dark)')
//     theme = mq.matches ? 'dark' : 'light'
//     localStorage.setItem('theme', theme)
//   }

//   return null
// }
