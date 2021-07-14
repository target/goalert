import React, { forwardRef, ForwardRefRenderFunction } from 'react'
import {
  Link as RouterLink,
  LinkProps as RouterLinkProps,
} from 'react-router-dom'
import Link, { LinkProps } from '@material-ui/core/Link'
import { useSelector } from 'react-redux'
import { urlPathSelector } from '../selectors'
import joinURL from './joinURL'

export interface AppLinkProps extends RouterLinkProps {
  to: string
  newTab?: boolean
}

export interface MuiLinkProps extends LinkProps {
  to: string
  newTab?: boolean
}

function usePath(to: string): string {
  const path = useSelector(urlPathSelector)
  return to.startsWith('/') ? to : joinURL(path, to)
}

const AppLink: ForwardRefRenderFunction<HTMLAnchorElement, AppLinkProps> =
  function AppLink(props, ref): JSX.Element {
    const { to: _to, newTab, ...other } = props
    const path = usePath(_to)

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

    return <RouterLink to={path} ref={ref} {...other} />
  }

export function MuiLink(props: MuiLinkProps): JSX.Element {
  const { to: _to, newTab, ...other } = props
  const path = usePath(_to)

  return (
    <Link component={RouterLink} to={path} {...other}>
      {props.children}
    </Link>
  )
}

export default forwardRef(AppLink)
