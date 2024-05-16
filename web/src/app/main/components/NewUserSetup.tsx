import React, { useState } from 'react'
import UserContactMethodVerificationDialog from '../../users/UserContactMethodVerificationDialog'
import { useSessionInfo } from '../../util/RequireConfig'
import { useResetURLParams, useURLParam } from '../../actions'
import UserContactMethodCreateDialogDest from '../../users/UserContactMethodCreateDialogDest'

export default function NewUserSetup(): JSX.Element {
  const [isFirstLogin] = useURLParam('isFirstLogin', '')
  const clearIsFirstLogin = useResetURLParams('isFirstLogin')
  const [contactMethodID, setContactMethodID] = useState('')
  const { userID, ready } = useSessionInfo()

  if (!isFirstLogin || !ready) {
    return <React.Fragment />
  }
  if (contactMethodID) {
    return (
      <UserContactMethodVerificationDialog
        contactMethodID={contactMethodID}
        onClose={clearIsFirstLogin}
      />
    )
  }

  return (
    <UserContactMethodCreateDialogDest
      title='Welcome to GoAlert!'
      subtitle='To get started, please enter a contact method.'
      userID={userID}
      onClose={(contactMethodID) => {
        if (contactMethodID) {
          setContactMethodID(contactMethodID)
        } else {
          clearIsFirstLogin()
        }
      }}
    />
  )
}
