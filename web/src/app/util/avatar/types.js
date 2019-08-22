import React, { useEffect, useState } from 'react'
import { Layers, RotateRight, Today, VpnKey, Person } from '@material-ui/icons'
import { useSessionInfo } from '../RequireConfig'
import { Avatar } from '@material-ui/core'

function useImage(srcURL) {
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

  if (!valid) return null

  return srcURL
}

function avatar(Fallback, otherProps, imgSrc) {
  const src = useImage(imgSrc)

  return (
    <Avatar
      alt=''
      src={src || null}
      data-cy={src ? null : 'avatar-fallback'}
      {...otherProps}
    >
      {src ? null : <Fallback />}
    </Avatar>
  )
}

export function UserAvatar({ userID, ...otherProps }) {
  return avatar(
    Person,
    otherProps,
    userID ? `/api/v2/user-avatar/${userID}` : null,
  )
}

export function CurrentUserAvatar(otherProps) {
  const { userID } = useSessionInfo()
  return <UserAvatar userID={userID} {...otherProps} />
}

export function ServiceAvatar(props) {
  return avatar(VpnKey, props)
}

export function EPAvatar(props) {
  return avatar(Layers, props)
}
export function RotationAvatar(props) {
  return avatar(RotateRight, props)
}
export function ScheduleAvatar(props) {
  return avatar(Today, props)
}
