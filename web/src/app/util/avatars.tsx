import React, { useEffect, useState } from 'react'
import { Layers, RotateRight, Today, VpnKey, Person } from '@material-ui/icons'
import { useSessionInfo } from './RequireConfig'
import { Avatar, AvatarTypeMap, SvgIconProps } from '@material-ui/core'
import { OverridableComponent } from '@material-ui/core/OverridableComponent'

type IconProps = (props: SvgIconProps) => JSX.Element //type alias
interface AvatarProps extends OverridableComponent<AvatarTypeMap> {
  userID?: string
}

function useValidImage(srcURL?: string) {
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

function useAvatar(
  Fallback: IconProps,
  otherProps: AvatarProps,
  imgSrc?: string,
) {
  const validImage = useValidImage(imgSrc)

  const { userID, ...DOMProps } = otherProps

  return (
    <Avatar
      alt=''
      src={validImage ? imgSrc : undefined}
      data-cy={validImage ? null : 'avatar-fallback'}
      {...DOMProps}
    >
      {validImage ? null : <Fallback fontSize='small' />}
    </Avatar>
  )
}

export function UserAvatar(props: AvatarProps) {
  const { userID } = props
  return useAvatar(Person, props, userID && `/api/v2/user-avatar/${userID}`)
}

export function CurrentUserAvatar(props: AvatarProps) {
  const { ready, userID } = useSessionInfo()
  return useAvatar(Person, props, ready && `/api/v2/user-avatar/${userID}`)
}

export function ServiceAvatar(props: AvatarProps) {
  return useAvatar(VpnKey, props)
}

export function EPAvatar(props: AvatarProps) {
  return useAvatar(Layers, props)
}
export function RotationAvatar(props: AvatarProps) {
  return useAvatar(RotateRight, props)
}
export function ScheduleAvatar(props: AvatarProps) {
  return useAvatar(Today, props)
}
