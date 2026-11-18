import * as React from 'react'
import Breadcrumbs from '@mui/material/Breadcrumbs'
import { ChevronRight } from '@mui/icons-material'
import { useLocation } from 'wouter'
import { Theme } from '@mui/material'
import BreadCrumb from './BreadCrumb'

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore type definition is broken for this file
import { useIsWidthDown } from '../../util/useWidth'

// todo: not needed once appbar is using same color prop for dark/light modes
export const getContrastColor = (theme: Theme): string => {
  return theme.palette.getContrastText(
    theme.palette.mode === 'dark'
      ? theme.palette.background.paper
      : theme.palette.primary.main,
  )
}

export default function ToolbarPageTitle(): React.JSX.Element {
  const isMobile = useIsWidthDown('md')

  const [path] = useLocation()
  const parts = React.useMemo(() => path.split('/').slice(1), [path])

  return (
    <Breadcrumbs
      maxItems={isMobile ? 2 : undefined}
      // stay on one line and truncate the last crumb, so a long page title
      // does not push the toolbar actions off-screen on narrow viewports
      sx={{
        minWidth: 0,
        '& .MuiBreadcrumbs-ol': { flexWrap: 'nowrap' },
        '& .MuiBreadcrumbs-li': { minWidth: 0 },
        '& .MuiBreadcrumbs-li:not(:last-child), & .MuiBreadcrumbs-separator': {
          flexShrink: 0,
        },
      }}
      separator={
        <ChevronRight
          sx={{
            color: getContrastColor,
          }}
        />
      }
    >
      {parts.map((part, idx) => (
        <BreadCrumb
          key={part + idx}
          crumb={part}
          urlParts={parts}
          index={idx}
          link={path
            .split('/')
            .slice(0, idx + 2)
            .join('/')}
        />
      ))}
    </Breadcrumbs>
  )
}
