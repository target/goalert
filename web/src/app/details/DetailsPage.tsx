import React from 'react'
import statusStyles from '../util/statusStyles'
import { makeStyles } from '@material-ui/core/styles'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import Divider from '@material-ui/core/Divider'
import Grid from '@material-ui/core/Grid'
import Typography from '@material-ui/core/Typography'
import { ChevronRight } from '@material-ui/icons'
import Hidden from '@material-ui/core/Hidden'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction'
import ListSubheader from '@material-ui/core/ListSubheader'
import IconButton from '@material-ui/core/IconButton'
import Notices, { Notice } from './Notices'
import Markdown from '../util/Markdown'
import AppLink from '../util/AppLink'
import useWidth from '../util/useWidth'

function isDesktopMode(width: string): boolean {
  return width === 'md' || width === 'lg' || width === 'xl'
}

const useLinkStyles = makeStyles(() => statusStyles)
const useStyles = makeStyles((theme) => ({
  iconContainer: {
    [theme.breakpoints.down('sm')]: { float: 'top' },
    [theme.breakpoints.up('md')]: { float: 'left' },
    margin: 20,
  },
  linksContainer: {
    display: 'flex',
  },
  linksList: {
    width: '100%',
  },
  linksSubheader: {
    margin: 0,
    fontSize: 'larger',
  },
  spacing: {
    '&:not(:first-child)': {
      marginTop: 8,
    },
    '&:not(:last-child)': {
      marginBottom: 8,
    },
    '&:last-child': {
      marginBottom: 64,
    },
  },
  title: {
    fontSize: '1.5rem',
  },
}))

type LinkStatus = 'ok' | 'warn' | 'err'
interface DetailsLinkProps {
  url: string
  label: string
  subText?: JSX.Element
  status?: LinkStatus
}

function DetailsLink(p: DetailsLinkProps): JSX.Element {
  const classes = useLinkStyles()
  const width = useWidth()

  let cn = classes.noStatus
  if (status === 'ok') cn = classes.statusOK
  if (status === 'warn') cn = classes.statusWarning
  if (status === 'err') cn = classes.statusError

  return (
    <ListItem className={cn} component={AppLink} to={p.url} button>
      <ListItemText
        primary={p.label}
        primaryTypographyProps={
          isDesktopMode(width) ? undefined : { variant: 'h5' }
        }
        secondary={p.subText}
      />
      <ListItemSecondaryAction>
        <IconButton component={AppLink} to={p.url}>
          <ChevronRight />
        </IconButton>
      </ListItemSecondaryAction>
    </ListItem>
  )
}

interface DetailsPageProps {
  title: string
  details?: string
  icon?: JSX.Element
  links?: Array<DetailsLinkProps>
  notices?: Array<Notice>
  titleFooter?: JSX.Element
  pageFooter?: JSX.Element

  noMarkdown?: boolean
}

export default function DetailsPage(p: DetailsPageProps): JSX.Element {
  const classes = useStyles()
  const width = useWidth()

  let links = null
  if (p.links?.length) {
    links = (
      <List
        data-cy='route-links'
        className={classes.linksList}
        subheader={
          isDesktopMode(width) ? (
            <ListSubheader
              className={classes.linksSubheader}
              component='h2'
              color='primary'
            >
              Quick Links
            </ListSubheader>
          ) : undefined
        }
      >
        {isDesktopMode(width) ? <Divider /> : null}
        {p.links.map((li, idx) => (
          <DetailsLink key={idx} {...li} />
        ))}
      </List>
    )
  }

  return (
    <Grid container>
      {(p.notices?.length ?? 0) > 0 && (
        <Grid className={classes.spacing} item xs={12}>
          <Notices notices={p.notices} />
        </Grid>
      )}
      <Grid className={classes.spacing} item xs={12}>
        <Card>
          <CardContent>
            <Grid container spacing={2}>
              <Grid item xs={isDesktopMode(width) && links ? 8 : 12}>
                {p.icon && (
                  <div className={classes.iconContainer}>{p.icon}</div>
                )}
                <Typography
                  data-cy='details-heading'
                  className={classes.title}
                  component='h2'
                >
                  {p.title}
                </Typography>
                <Typography
                  data-cy='details'
                  variant='subtitle1'
                  component='div'
                >
                  {p.noMarkdown ? p.details : <Markdown value={p.details} />}
                </Typography>
                {p.titleFooter && (
                  <Typography
                    component='div'
                    variant='subtitle1'
                    data-cy='title-footer'
                  >
                    {p.titleFooter}
                  </Typography>
                )}
              </Grid>
              {links && (
                <Hidden smDown>
                  <Grid className={classes.linksContainer} item xs={4}>
                    <Divider orientation='vertical' />
                    {links}
                  </Grid>
                </Hidden>
              )}
            </Grid>
          </CardContent>
        </Card>
      </Grid>
      <Hidden mdUp>
        <Grid className={classes.spacing} item xs={12}>
          <Card>{links}</Card>
        </Grid>
      </Hidden>
      {p.pageFooter && (
        <Grid className={classes.spacing} item xs={12}>
          {p.pageFooter}
        </Grid>
      )}
    </Grid>
  )
}

DetailsPage.defaultProps = {
  noMarkdown: false,
}
