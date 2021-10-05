import React, { cloneElement, forwardRef } from 'react'
import makeStyles from '@mui/styles/makeStyles'
import Card from '@mui/material/Card'
import CardHeader from '@mui/material/CardHeader'
import CardContent from '@mui/material/CardContent'
import Grid from '@mui/material/Grid'
import Typography from '@mui/material/Typography'
import { ChevronRight } from '@mui/icons-material'
import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemText from '@mui/material/ListItemText'
import { ReactNode } from 'react-markdown'

import Notices, { Notice } from './Notices'
import Markdown from '../util/Markdown'
import CardActions, { Action } from './CardActions'
import AppLink, { AppLinkProps } from '../util/AppLink'
import { useIsWidthDown } from '../util/useWidth'
import statusStyles from '../util/statusStyles'

interface DetailsPageProps {
  title: string

  // optional content
  avatar?: JSX.Element // placement for an icon or image
  subheader?: string | JSX.Element
  details?: string
  notices?: Array<Notice>
  links?: Array<Link>
  pageContent?: JSX.Element
  primaryActions?: Array<Action | JSX.Element>
  secondaryActions?: Array<Action | JSX.Element>
}

type LinkStatus = 'ok' | 'warn' | 'err'
type Link = {
  url: string
  label: string
  subText?: string
  status?: LinkStatus
}

const useStyles = makeStyles({
  ...statusStyles,
  flexHeight: {
    flexGrow: 1,
  },
  fullHeight: {
    height: '100%', // align height of the first row of cards together
  },
  headerContent: {
    paddingTop: 0,
  },
  quickLinks: {
    paddingBottom: 8,
  },
  smPageBottom: {
    marginBottom: 64,
  },
})

const LIApplink = forwardRef<HTMLAnchorElement, AppLinkProps>(
  function LIApplink(props, ref): JSX.Element {
    return (
      <li>
        <AppLink ref={ref} {...props} />
      </li>
    )
  },
)

export default function DetailsPage(p: DetailsPageProps): JSX.Element {
  const classes = useStyles()
  const isMobile = useIsWidthDown('sm')

  const linkClassName = (status?: LinkStatus): string => {
    if (status === 'ok') return classes.statusOK
    if (status === 'warn') return classes.statusWarning
    if (status === 'err') return classes.statusError
    return classes.noStatus
  }

  const avatar = (): ReactNode => {
    if (!p.avatar) return null
    return cloneElement(p.avatar, {
      style: { width: 56, height: 56 },
    })
  }

  const links = (p.links || []).filter((l) => l)

  return (
    <Grid container spacing={2}>
      {/* Notices */}
      {Boolean(p.notices?.length) && (
        <Grid item xs={12}>
          <Notices notices={p.notices} />
        </Grid>
      )}

      {/* Header card */}
      <Grid item xs={12} lg={!isMobile && p.links?.length ? 8 : 12}>
        <Card className={classes.fullHeight}>
          <Grid
            className={classes.fullHeight}
            item
            xs
            container
            direction='column'
          >
            <Grid item>
              <CardHeader
                title={p.title}
                subheader={p.subheader}
                avatar={avatar()}
                titleTypographyProps={{
                  'data-cy': 'title',
                  variant: 'h5',
                  component: 'h2',
                }}
                subheaderTypographyProps={{
                  'data-cy': 'subheader',
                  variant: 'body1',
                }}
              />
            </Grid>

            {p.details && (
              <Grid item>
                <CardContent className={classes.headerContent}>
                  <Typography
                    component='div'
                    variant='subtitle1'
                    color='textSecondary'
                    data-cy='details'
                  >
                    <Markdown value={p.details} />
                  </Typography>
                </CardContent>
              </Grid>
            )}

            <Grid className={classes.flexHeight} item />
            {(p.primaryActions?.length || p.secondaryActions?.length) && (
              <Grid item>
                <CardActions
                  primaryActions={p.primaryActions}
                  secondaryActions={p.secondaryActions}
                />
              </Grid>
            )}
          </Grid>
        </Card>
      </Grid>

      {/* Quick Links */}
      {links.length > 0 && (
        <Grid item xs={12} lg={!isMobile && links.length ? 4 : 12}>
          <Card className={classes.fullHeight}>
            <CardHeader
              title='Quick Links'
              titleTypographyProps={{
                variant: 'h5',
                component: 'h2',
              }}
            />
            <List data-cy='route-links' className={classes.quickLinks} dense>
              {links.map((li, idx) => (
                <ListItem
                  key={idx}
                  className={linkClassName(li.status)}
                  component={LIApplink}
                  to={li.url}
                  button
                >
                  <ListItemText
                    primary={li.label}
                    primaryTypographyProps={
                      !isMobile ? undefined : { variant: 'h5' }
                    }
                    secondary={li.subText}
                  />
                  <ChevronRight />
                </ListItem>
              ))}
            </List>
          </Card>
        </Grid>
      )}

      {/* Primary Page Content */}
      {p.pageContent && (
        <Grid
          className={!isMobile ? classes.smPageBottom : undefined}
          item
          xs={12}
        >
          {p.pageContent}
        </Grid>
      )}
    </Grid>
  )
}
