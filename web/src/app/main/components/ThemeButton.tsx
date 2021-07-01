import React, { useContext } from 'react'
import IconButton from '@material-ui/core/IconButton'
import makeStyles from '@material-ui/core/styles/makeStyles'
import DarkModeIcon from '@material-ui/icons/Brightness4'
import LightModeIcon from '@material-ui/icons/BrightnessHigh'
import { AppContext } from '../context'

const useStyles = makeStyles((theme) => ({
  icon: {
    color: theme.palette.getContrastText(theme.palette.primary.main),
  },
}))

export default function ThemeButton(): JSX.Element {
  const classes = useStyles()
  const { theme, setTheme } = useContext(AppContext)

  return (
    <IconButton onClick={() => setTheme(theme)}>
      {theme === 'dark' ? (
        <DarkModeIcon className={classes.icon} />
      ) : (
        <LightModeIcon className={classes.icon} />
      )}
    </IconButton>
  )
}
