import React, { useState } from 'react'
import UserContactMethodCreateDialog from '../../users/UserContactMethodCreateDialog'
import UserContactMethodVerificationDialog from '../../users/UserContactMethodVerificationDialog'
import { useDispatch, useSelector } from 'react-redux'
import { urlParamSelector } from '../../selectors'
import { useSessionInfo, useConfigValue } from '../../util/RequireConfig'
import { resetURLParams } from '../../actions'

export default function NewUserSetup() {
  const urlParams = useSelector(urlParamSelector)
  const isFirstLogin = urlParams('isFirstLogin', '')
  const dispatch = useDispatch()
  const clearIsFirstLogin = () => dispatch(resetURLParams('isFirstLogin'))
  const [contactMethodID, setContactMethodID] = useState('')
  const { userID, ready } = useSessionInfo()
  const [disclaimer] = useConfigValue('General.NotificationDisclaimer')

  if (!isFirstLogin || !ready) {
    return null
  }
  if (contactMethodID) {
    return (
      <UserContactMethodVerificationDialog
        contactMethodID={contactMethodID.contactMethodID}
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
      onClose={result => {
        if (result && result.contactMethodID) {
          setContactMethodID(result)
        } else {
          clearIsFirstLogin()
        }
      }}
    />
  )
}
