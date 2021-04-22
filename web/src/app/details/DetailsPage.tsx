import React, { ReactNode } from 'react'
import statusStyles from '../util/statusStyles'
import { makeStyles } from '@material-ui/core/styles'
import Card from '@material-ui/core/Card'
import CardHeader from '@material-ui/core/CardHeader'
import CardContent from '@material-ui/core/CardContent'
import Grid from '@material-ui/core/Grid'
import Typography from '@material-ui/core/Typography'
import { ChevronRight } from '@material-ui/icons'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction'
import IconButton from '@material-ui/core/IconButton'

import Notices, { Notice } from './Notices'
import Markdown from '../util/Markdown'
import AppLink from '../util/AppLink'
import useWidth from '../util/useWidth'
import CardActions, { Action } from './CardActions'

interface DetailsPageProps {
  title: string

  // content options
  details?: ReactNode
  thumbnail?: ReactNode // placement for an icon or image

  notices?: Array<Notice>
  links?: Array<Link>

  headerContent?: string | JSX.Element
  primaryContent?: JSX.Element

  primaryActions?: Array<Action | JSX.Element>
  secondaryActions?: Array<Action | JSX.Element>

  // api options
  markdown?: boolean // enables markdown support for details. default: false
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
  primaryCard: {
    height: '100%', // align with quick links if shorter in height
    position: 'relative', // allows card actions to remain at bottom, if height is stretched
  },
  titleFooterContent: {
    paddingTop: 0,
  },
})

export default function DetailsPage(p: DetailsPageProps): JSX.Element {
  const classes = useStyles()
  const width = useWidth()

  const linkClassName = (status?: LinkStatus): string => {
    if (status === 'ok') return classes.statusOK
    if (status === 'warn') return classes.statusWarning
    if (status === 'err') return classes.statusError
    return classes.noStatus
  }

  return (
    <Grid container spacing={2}>
      {/* Notices */}
      {(p.notices?.length ?? 0) > 0 && (
        <Grid item xs={12}>
          <Notices notices={p.notices} />
        </Grid>
      )}

      {/* Header card */}
      <Grid item xs={12} container spacing={2}>
        <Grid item xs={isDesktopMode(width) && p.links?.length ? 8 : 12}>
          <Card className={classes.primaryCard}>
            <CardHeader
              title={p.title}
              subheader={
                p.markdown ? <Markdown value={p.details} /> : p.details
              }
              avatar={p.thumbnail}
              titleTypographyProps={{
                variant: 'h5',
                component: 'h2',
              }}
              subheaderTypographyProps={{
                variant: 'body1',
              }}
            />

            {p.headerContent && (
              <CardContent className={classes.titleFooterContent}>
                <Typography
                  component='div'
                  variant='subtitle1'
                  color='textSecondary'
                  data-cy='title-footer'
                >
                  {p.headerContent}
                </Typography>
              </CardContent>
            )}

            <CardActions
              primaryActions={p.primaryActions}
              secondaryActions={p.secondaryActions}
            />
          </Card>
        </Grid>

        {/* Quick Links */}
        {p.links?.length && (
          <Grid item xs={isDesktopMode(width) && p.links?.length ? 4 : 12}>
            <Card>
              <CardHeader title='Quick Links' />
              <List data-cy='route-links'>
                {p.links.map((li, idx) => (
                  <ListItem
                    key={idx}
                    className={linkClassName(li.status)}
                    component={AppLink}
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
                    <ListItemSecondaryAction>
                      <IconButton component={AppLink} to={li.url}>
                        <ChevronRight />
                      </IconButton>
                    </ListItemSecondaryAction>
                  </ListItem>
                ))}
              </List>
            </Card>
          </Grid>
        )}
      </Grid>

      {/* Primary Content */}
      {p.primaryContent && (
        <Grid item xs={12}>
          {p.primaryContent}
        </Grid>
      )}
    </Grid>
  )
}

DetailsPage.defaultProps = {
  noMarkdown: false,
}
