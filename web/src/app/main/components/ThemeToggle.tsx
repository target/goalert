import React, { useContext } from 'react'
import IconButton from '@mui/material/IconButton'
import makeStyles from '@mui/styles/makeStyles'
import DarkModeIcon from '@mui/icons-material/Brightness4'
import LightModeIcon from '@mui/icons-material/BrightnessHigh'
import { ThemeContext } from '../themeConfig'

const useStyles = makeStyles(() => ({
  icon: {
    color: 'white',
  },
}))

export default function ThemeToggle(): JSX.Element {
  const classes = useStyles()
  const { themeMode, toggleThemeMode } = useContext(ThemeContext)

  return (
    <IconButton onClick={() => toggleThemeMode()}>
      {themeMode === 'dark' ? (
        <DarkModeIcon className={classes.icon} />
      ) : (
        <LightModeIcon className={classes.icon} />
      )}
    </IconButton>
  )
}
