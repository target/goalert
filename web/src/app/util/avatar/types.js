import React, { useEffect, useState } from 'react'
import { Layers, RotateRight, Today, VpnKey, Person } from '@material-ui/icons'
import { useSessionInfo } from '../RequireConfig'
import { Avatar } from '@material-ui/core'
import { absURLSelector } from '../../selectors'
import { useSelector } from 'react-redux'

function useValidImage(srcURL) {
  const [valid, setValid] = useState(false)

  useEffect(() => {
    setValid(false)
    if (!srcURL) return

    const image = new Image()
    image.onload = () => setValid(true)
    image.src = srcURL

    return () => {
      image.onload = null
    }
  }, [srcURL])

  return valid
}

function useAvatar(Fallback, otherProps, imgSrc) {
  const validImage = useValidImage(imgSrc)

  return (
    <Avatar
      alt=''
      src={validImage ? imgSrc : null}
      data-cy={validImage ? null : 'avatar-fallback'}
      {...otherProps}
    >
      {validImage ? null : <Fallback fontSize='small' />}
    </Avatar>
  )
}

export function UserAvatar({ userID, ...otherProps }) {
  const absURL = useSelector(absURLSelector)
  return useAvatar(
    Person,
    otherProps,
    userID ? absURL(`/api/v2/user-avatar/${userID}`) : null,
  )
}

export function CurrentUserAvatar(otherProps) {
  const { userID } = useSessionInfo()
  return <UserAvatar userID={userID} {...otherProps} />
}

export function ServiceAvatar(props) {
  return useAvatar(VpnKey, props)
}

export function EPAvatar(props) {
  return useAvatar(Layers, props)
}
export function RotationAvatar(props) {
  return useAvatar(RotateRight, props)
}
export function ScheduleAvatar(props) {
  return useAvatar(Today, props)
}
