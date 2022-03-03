import React from 'react'
import { CurrentUserAvatar } from '../../util/avatars'
// import { useSessionInfo } from '../../util/RequireConfig'

export default function UserSettingsPopover(): JSX.Element {
  //   const { userID } = useSessionInfo()
  //   const dispatch = useDispatch()
  //   const logout = () => dispatch(authLogout(true))

  return (
    <React.Fragment>
      <CurrentUserAvatar />
      {/* <Config>
        {(cfg) =>
          cfg['Feedback.Enable'] &&
          renderFeedback(
            cfg['Feedback.OverrideURL'] ||
              'https://www.surveygizmo.com/s3/4106900/GoAlert-Feedback',
          )
        }
      </Config> */}
      {/* {renderSidebarLink(LogoutIcon, '/api/v2/identity/logout', 'Logout', {
        onClick: (e) => {
          e.preventDefault()
          logout()
        },
      })} */}
    </React.Fragment>
  )
}
