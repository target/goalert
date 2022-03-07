import React from 'react'
import Button from '@mui/material/Button'
import IconButton from '@mui/material/IconButton'
import Grid from '@mui/material/Grid'
import Popover from '@mui/material/Popover'
import Typography from '@mui/material/Typography'
import { OpenInNew, Logout } from '@mui/icons-material'
import { useDispatch } from 'react-redux'
import { authLogout } from '../../actions'
import ThemePicker from './ThemePicker'
import { CurrentUserAvatar } from '../../util/avatars'
import AppLink from '../../util/AppLink'
import { useConfigValue, useSessionInfo } from '../../util/RequireConfig'

export default function UserSettingsPopover(): JSX.Element {
  const [feedbackEnabled] = useConfigValue('Feedback.Enable')
  const { userName } = useSessionInfo()
  const firstName = userName?.split(' ')[0]

  const dispatch = useDispatch()
  // eslint-disable-next-line @typescript-eslint/explicit-function-return-type
  const logout = () => dispatch(authLogout(true))

  const [anchorEl, setAnchorEl] = React.useState<HTMLButtonElement | null>(null)
  const open = Boolean(anchorEl)

  return (
    <React.Fragment>
      <IconButton
        size='small'
        onClick={(event) => setAnchorEl(event.currentTarget)}
      >
        <CurrentUserAvatar />
      </IconButton>
      <Popover
        open={open}
        anchorEl={anchorEl}
        onClose={() => setAnchorEl(null)}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'right',
        }}
        PaperProps={{ sx: { p: 2 } }}
      >
        <Grid container direction='column' spacing={2}>
          <Grid item>
            <Typography variant='h5' component='span'>
              Hello, {firstName}!
            </Typography>
          </Grid>
          <Grid item>
            <AppLink to='/profile' style={{ textDecoration: 'none' }}>
              <Button variant='outlined' onClick={() => setAnchorEl(null)}>
                Manage Profile
              </Button>
            </AppLink>
          </Grid>
          <Grid item>
            <ThemePicker />
          </Grid>
          <Grid item>
            {feedbackEnabled && (
              <AppLink
                newTab
                to='https://www.surveygizmo.com/s3/4106900/GoAlert-Feedback'
                style={{ textDecoration: 'none' }}
              >
                <Button variant='outlined' endIcon={<OpenInNew />}>
                  Feedback
                </Button>
              </AppLink>
            )}
          </Grid>
          <Grid item>
            <Button
              variant='text'
              startIcon={<Logout />}
              onClick={(e) => {
                e.preventDefault()
                logout()
              }}
              sx={{ color: (theme) => theme.palette.error.main }}
            >
              Logout
            </Button>
          </Grid>
        </Grid>
      </Popover>
    </React.Fragment>
  )
}
