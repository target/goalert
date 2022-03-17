import React, { useContext } from 'react'
import Grid from '@mui/material/Grid'
import ToggleButton from '@mui/material/ToggleButton'
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup'
import Typography from '@mui/material/Typography'
import DarkModeIcon from '@mui/icons-material/DarkMode'
import SystemIcon from '@mui/icons-material/SettingsBrightness'
import LightModeIcon from '@mui/icons-material/LightMode'
import { CirclePicker as ColorPicker, ColorResult } from 'react-color'
import { sourceColors, ThemeContext } from '../../theme/themeConfig'

export default function ThemePicker(): JSX.Element {
  const { themeMode, setThemeMode, sourceColor, setSourceColor } =
    useContext(ThemeContext)

  return (
    <Grid container direction='column' spacing={2}>
      <Grid item>
        <Typography variant='body1'>Theme Mode</Typography>
      </Grid>
      <Grid item>
        <ToggleButtonGroup color='primary' value={themeMode}>
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
      </Grid>
      <Grid item>
        <ColorPicker
          color={sourceColor}
          colors={sourceColors}
          onChange={(color: ColorResult) => setSourceColor(color.hex)}
        />
      </Grid>
    </Grid>
  )
}
