import React, { cloneElement, forwardRef } from 'react'
import { makeStyles } from '@material-ui/core/styles'
import { isWidthDown } from '@material-ui/core'
import Card from '@material-ui/core/Card'
import CardHeader from '@material-ui/core/CardHeader'
import CardContent from '@material-ui/core/CardContent'
import Grid from '@material-ui/core/Grid'
import Typography from '@material-ui/core/Typography'
import { ChevronRight } from '@material-ui/icons'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import { ReactNode } from 'react-markdown'

import Notices, { Notice } from './Notices'
import Markdown from '../util/Markdown'
import CardActions, { Action } from './CardActions'
import AppLink, { AppLinkProps } from '../util/AppLink'
import useWidth from '../util/useWidth'
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

function isDesktopMode(width: string): boolean {
  return width === 'md' || width === 'lg' || width === 'xl'
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
  const width = useWidth()

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
      <Grid item xs={12} lg={isDesktopMode(width) && p.links?.length ? 8 : 12}>
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
      {links.length && (
        <Grid item xs={12} lg={isDesktopMode(width) && links.length ? 4 : 12}>
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
                      isDesktopMode(width) ? undefined : { variant: 'h5' }
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
          className={
            isWidthDown('sm', width) ? classes.smPageBottom : undefined
          }
          item
          xs={12}
        >
          {p.pageContent}
        </Grid>
      )}
    </Grid>
  )
}
