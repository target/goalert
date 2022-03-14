import React, { forwardRef, ForwardRefRenderFunction } from 'react'
import { LinkProps } from '@mui/material'
import Link from '@mui/material/Link'
import {
  Link as RRLink,
  LinkProps as RRLinkProps,
  useLocation,
} from 'react-router-dom'
import joinURL from './joinURL'

type MergedLinkProps = Omit<LinkProps & RRLinkProps, 'to' | 'href'>

export interface AppLinkProps extends MergedLinkProps {
  to: string
  newTab?: boolean
  onClick?: React.MouseEventHandler<HTMLAnchorElement> // use explicit anchor elem
}

const AppLink: ForwardRefRenderFunction<HTMLAnchorElement, AppLinkProps> =
  function AppLink(props, ref): JSX.Element {
    let { to, newTab, ...other } = props
    const { pathname } = useLocation()

    if (newTab) {
      other.target = '_blank'
      other.rel = 'noopener noreferrer'
    }

    const external = /^(tel:|mailto:|https?:\/\/)/.test(to)

    // handle relative URLs
    if (!external && !to.startsWith('/')) {
      to = joinURL(pathname, to)
    }

    return (
      <Link
        ref={ref}
        to={to}
        href={to}
        component={external ? 'a' : RRLink}
        color='secondary'
        {...other}
      />
    )
  }

export default forwardRef(AppLink)
