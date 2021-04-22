import React, { MouseEventHandler, ReactNode } from 'react'
import statusStyles from '../util/statusStyles'
import { makeStyles } from '@material-ui/core/styles'
import Card from '@material-ui/core/Card'
import CardHeader from '@material-ui/core/CardHeader'
import CardContent from '@material-ui/core/CardContent'
import CardActions from '@material-ui/core/CardActions'
import Button from '@material-ui/core/Button'
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

interface DetailsPageProps {
  title: string

  // content options
  details?: ReactNode
  thumbnail?: ReactNode // placement for an icon or image

  notices?: Array<Notice>
  links?: Array<Link>

  titleFooter?: JSX.Element
  pageFooter?: JSX.Element

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

type Action = {
  label: string
  handleOnClick: MouseEventHandler<HTMLButtonElement>

  icon?: JSX.Element // if true, adds a start icon to a button with text
  secondary?: boolean // if true, renders right-aligned as an icon button
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
  cardActions: {
    // height: '100%',
    alignItems: 'flex-end', // aligns icon buttons to bottom of container

    // moves card actions to bottom if height is stretched
    position: 'absolute',
    bottom: '0',
    width: '-webkit-fill-available',
  },
})

function Action(p: Action): JSX.Element {
  if (p.secondary && p.icon) {
    return <IconButton onClick={p.handleOnClick}>{p.icon}</IconButton>
  }
  return (
    <Button onClick={p.handleOnClick} startIcon={p.icon}>
      {p.label}
    </Button>
  )
}

export default function DetailsPage(p: DetailsPageProps): JSX.Element {
  const classes = useStyles()
  const width = useWidth()

  const linkClassName = (status?: LinkStatus): string => {
    if (status === 'ok') return classes.statusOK
    if (status === 'warn') return classes.statusWarning
    if (status === 'err') return classes.statusError
    return classes.noStatus
  }

  const action = (action: Action | JSX.Element, key: string): JSX.Element => {
    if ('label' in action && 'handleOnClick' in action) {
      return <Action key={key} {...action} />
    }
    return action
  }

  const renderActions = (): JSX.Element => {
    let actions: Array<JSX.Element> = []
    if (p.primaryActions) {
      actions = p.primaryActions.map((a, i) => action(a, 'primary' + i))
    }
    if (p.secondaryActions) {
      actions = [
        ...actions,
        <div key='actions-margin' style={{ margin: '0 auto' }} />,
        ...p.secondaryActions.map((a, i) =>
          action({ ...a, secondary: true }, 'secondary' + i),
        ),
      ]
    }
    return <CardActions className={classes.cardActions}>{actions}</CardActions>
  }

  return (
    <Grid container spacing={2}>
      {/* Notices */}
      {(p.notices?.length ?? 0) > 0 && (
        <Grid item xs={12}>
          <Notices notices={p.notices} />
        </Grid>
      )}

      {/* Primary card */}
      <Grid item xs={12} container spacing={2}>
        <Grid item xs={isDesktopMode(width) && p.links?.length ? 8 : 12}>
          <Card className={classes.primaryCard}>
            <CardHeader
              title={p.title}
              subheader={
                p.markdown ? <Markdown value={p.details} /> : p.details
              }
              avatar={p.thumbnail}
            />

            {p.titleFooter && (
              <CardContent>
                <Typography
                  component='div'
                  variant='subtitle1'
                  data-cy='title-footer'
                >
                  {p.titleFooter}
                </Typography>
              </CardContent>
            )}

            {(p.primaryActions || p.secondaryActions) && renderActions()}
          </Card>
        </Grid>

        {/* Quick Links */}
        <Grid item xs={isDesktopMode(width) && p.links?.length ? 4 : 12}>
          <Card>
            <CardHeader title='Quick Links' />
            {p.links?.length && (
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
            )}
          </Card>
        </Grid>
      </Grid>

      {/* Footer node */}
      {p.pageFooter && (
        <Grid item xs={12}>
          {p.pageFooter}
        </Grid>
      )}
    </Grid>
  )
}

DetailsPage.defaultProps = {
  noMarkdown: false,
}
