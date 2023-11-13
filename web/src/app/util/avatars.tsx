import React, { useEffect, useState } from 'react'
import { Layers, RotateRight, Today, VpnKey, Person } from '@mui/icons-material'
import { useSessionInfo } from './RequireConfig'
import { Avatar, SvgIconProps, AvatarProps, Skeleton } from '@mui/material'
import { pathPrefix } from '../env'

interface UserAvatarProps extends AvatarProps {
  userID: string
}

function useValidImage(srcURL?: string): boolean {
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
  Fallback: React.FC<SvgIconProps>,
  otherProps: AvatarProps,
  loading = false,
  imgSrc?: string,
): React.ReactNode {
  const validImage = useValidImage(imgSrc)

  const av = (
    <Avatar
      alt=''
      src={validImage ? imgSrc : undefined}
      data-cy={validImage ? null : 'avatar-fallback'}
      {...otherProps}
    >
      {validImage ? null : <Fallback color='primary' />}
    </Avatar>
  )

  if (loading) {
    return <Skeleton variant='circular'>{av}</Skeleton>
  }

  return av
}

export function UserAvatar(props: UserAvatarProps): React.ReactNode {
  const { userID, ...otherProps } = props
  return useAvatar(
    Person,
    otherProps as AvatarProps,
    false,
    pathPrefix + `/api/v2/user-avatar/${userID}`,
  )
}

export function CurrentUserAvatar(props: AvatarProps): React.ReactNode {
  const { ready, userID } = useSessionInfo()
  return useAvatar(
    Person,
    props,
    !ready,
    ready ? pathPrefix + `/api/v2/user-avatar/${userID}` : undefined,
  )
}

export function ServiceAvatar(props: AvatarProps): React.ReactNode {
  return useAvatar(VpnKey, props)
}

export function EPAvatar(props: AvatarProps): React.ReactNode {
  return useAvatar(Layers, props)
}

export function RotationAvatar(props: AvatarProps): React.ReactNode {
  return useAvatar(RotateRight, props)
}

export function ScheduleAvatar(props: AvatarProps): React.ReactNode {
  return useAvatar(Today, props)
}
