import React, { useContext } from 'react'
import ToggleButton from '@mui/material/ToggleButton'
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup'
import Typography from '@mui/material/Typography'
import DarkModeIcon from '@mui/icons-material/DarkMode'
import SystemIcon from '@mui/icons-material/SettingsBrightness'
import LightModeIcon from '@mui/icons-material/LightMode'
import { ThemeContext } from '../themeConfig'

export default function ThemePicker(): JSX.Element {
  const { themeMode, setThemeMode } = useContext(ThemeContext)

  return (
    <div>
      <Typography variant='body1' sx={{ paddingBottom: '0.5rem' }}>
        Theme Mode
      </Typography>
      <ToggleButtonGroup value={themeMode}>
        <ToggleButton value='light' onClick={() => setThemeMode('light')}>
          <LightModeIcon sx={{ paddingRight: 1 }} />
          Light
        </ToggleButton>
        <ToggleButton value='system' onClick={() => setThemeMode('system')}>
          <SystemIcon sx={{ paddingRight: 1 }} />
          System
        </ToggleButton>
        <ToggleButton value='dark' onClick={() => setThemeMode('dark')}>
          <DarkModeIcon sx={{ paddingRight: 1 }} />
          Dark
        </ToggleButton>
      </ToggleButtonGroup>
    </div>
  )
}
