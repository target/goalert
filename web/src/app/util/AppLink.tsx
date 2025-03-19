import React, { forwardRef } from 'react'
import { LinkProps } from '@mui/material'
import MUILink from '@mui/material/Link'

import { Link, LinkProps as WLinkProps, useLocation } from 'wouter'
import joinURL from './joinURL'

type MergedLinkProps = Omit<LinkProps & WLinkProps, 'to' | 'href'>

export interface AppLinkProps extends MergedLinkProps {
  to: string
  newTab?: boolean
  onClick?: React.MouseEventHandler<HTMLAnchorElement> // use explicit anchor elem
}

interface WrapLinkProps {
  to: string
  children: React.ReactNode
}

const WrapLink = forwardRef(function WrapLink(
  props: WrapLinkProps,
  ref: React.Ref<HTMLAnchorElement>,
) {
  const { to, children, ...rest } = props
  return (
    <Link to={to} ref={ref} {...rest}>
      {children}
    </Link>
  )
})

const AppLink = forwardRef<HTMLAnchorElement, AppLinkProps>((props, ref) => {
  let { to, newTab, ...other } = props
  const [location] = useLocation()

  if (newTab) {
    other.target = '_blank'
    other.rel = 'noopener noreferrer'
  }

  const external = /^(tel:|mailto:|blob:|https?:\/\/)/.test(to)

  // handle relative URLs
  if (!external && !to.startsWith('/')) {
    to = joinURL(location, to)
  }

  return (
    <MUILink
      {...other}
      ref={ref}
      to={to}
      href={to}
      component={external || newTab ? 'a' : WrapLink}
    />
  )
})

AppLink.displayName = 'AppLink'
export default AppLink

// forwardRef required to shut console up
export const AppLinkListItem = forwardRef<HTMLAnchorElement, AppLinkProps>(
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  (props, _) => (
    <li>
      <AppLink {...props} />
    </li>
  ),
)
AppLinkListItem.displayName = 'AppLinkListItem'
