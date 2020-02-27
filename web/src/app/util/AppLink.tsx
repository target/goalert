import React from 'react'
import { Link, LinkProps } from 'react-router-dom'
import { useSelector } from 'react-redux'
import { urlPathSelector } from '../selectors'
import joinURL from './joinURL'

interface AppLinkProps extends LinkProps {
  to: string
}

export function AppLink(props: AppLinkProps) {
  const { to: _to, ...other } = props
  const path = useSelector(urlPathSelector)

  if (/^(mailto:|https?:\/\/)/.test(_to)) {
    return <a href={_to} {...other} />
  }

  const to = _to.startsWith('/') ? _to : joinURL(path, _to)
  return <Link to={to} {...other} />
}
