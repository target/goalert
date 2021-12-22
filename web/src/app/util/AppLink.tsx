import React, { forwardRef, ForwardRefRenderFunction } from 'react'
import { Link, LinkProps, useLocation } from 'react-router-dom'
import joinURL from './joinURL'
import { OpenInNew } from '@mui/icons-material'

export interface AppLinkProps extends LinkProps {
  to: string
  newTab?: boolean
  icon?: boolean
}

const AppLink: ForwardRefRenderFunction<HTMLAnchorElement, AppLinkProps> =
  function AppLink(props, ref): JSX.Element {
    const { to: _to, newTab, icon, ...other } = props
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

    if (icon) {
      return (
        <div style={{ display: 'flex', alignContent: 'center' }}>
          <Link to={to} ref={ref} {...other} />
          <a href={_to} ref={ref} {...other} style={{ paddingLeft: '2px' }}>
            <OpenInNew fontSize='small' />
          </a>
        </div>
      )
    }

    return <Link to={to} ref={ref} {...other} />
  }

export default forwardRef(AppLink)
