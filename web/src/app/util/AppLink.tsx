import React, { forwardRef, ForwardRefRenderFunction } from 'react'
import { Link, LinkProps } from 'react-router-dom'
import joinURL from './joinURL'
import { useLocation } from 'react-router'

export interface AppLinkProps extends LinkProps {
  to: string
  newTab?: boolean
}

const AppLink: ForwardRefRenderFunction<HTMLAnchorElement, AppLinkProps> =
  function AppLink(props, ref): JSX.Element {
    const { to: _to, newTab, ...other } = props
    const { pathname } = useLocation()

    if (newTab) {
      other.target = '_blank'
      other.rel = 'noopener noreferrer'
    }

    if (/^(mailto:|https?:\/\/)/.test(_to)) {
      return (
        <a href={_to} ref={ref} {...other}>
          {other.children}
        </a>
      )
    }

    const to = _to.startsWith('/') ? _to : joinURL(pathname, _to)
    return <Link to={to} ref={ref} {...other} />
  }

export default forwardRef(AppLink)
