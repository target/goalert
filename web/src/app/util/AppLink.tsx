import React, { forwardRef, ForwardRefRenderFunction } from 'react'
import { LinkProps } from '@mui/material'
import MUILink from '@mui/material/Link'

import { Link, LinkProps as WLinkProps } from 'wouter'
import joinURL from './joinURL'

type MergedLinkProps = Omit<LinkProps & WLinkProps, 'to' | 'href'>

export interface AppLinkProps extends MergedLinkProps {
  to: string
  newTab?: boolean
  onClick?: React.MouseEventHandler<HTMLAnchorElement> // use explicit anchor elem
}

function WrapLink(props, ref) {
  return (
    <Link to={props.to}>
      <a ref={ref} {...props} />
    </Link>
  )
}

const AppLink: ForwardRefRenderFunction<HTMLAnchorElement, AppLinkProps> =
  function AppLink(props, ref): JSX.Element {
    let { to, newTab, ...other } = props

    if (newTab) {
      other.target = '_blank'
      other.rel = 'noopener noreferrer'
    }

    const external = /^(tel:|mailto:|https?:\/\/)/.test(to)

    // handle relative URLs
    if (!external && !to.startsWith('/')) {
      to = joinURL(window.location.pathname, to)
    }

    return (
      <MUILink
        ref={ref}
        to={to}
        href={to}
        component={external ? 'a' : forwardRef(WrapLink)}
        {...other}
      />
    )
  }

export default forwardRef(AppLink)
