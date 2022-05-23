import React, { useState } from 'react'
import UserContactMethodCreateDialog from '../../users/UserContactMethodCreateDialog'
import UserContactMethodVerificationDialog from '../../users/UserContactMethodVerificationDialog'
import { useSessionInfo, useConfigValue } from '../../util/RequireConfig'
import { useResetURLParams, useURLParam } from '../../actions'

export default function NewUserSetup() {
  const [isFirstLogin] = useURLParam('isFirstLogin', '')
  const clearIsFirstLogin = useResetURLParams('isFirstLogin')
  const [contactMethodID, setContactMethodID] = useState('')
  const { userID, ready } = useSessionInfo()
  const [disclaimer] = useConfigValue('General.NotificationDisclaimer')

  if (!isFirstLogin || !ready) {
    return null
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
    <UserContactMethodCreateDialog
      title='Welcome to GoAlert!'
      subtitle='To get started, please enter a contact method.'
      disclaimer={disclaimer}
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
