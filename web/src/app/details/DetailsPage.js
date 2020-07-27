import React from 'react'
import p from 'prop-types'
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
import Markdown from '../util/Markdown'
import { AppLink } from '../util/AppLink'
import useWidth from '../util/useWidth'

function isDesktopMode(width) {
  return width === 'md' || width === 'lg' || width === 'xl'
}

const useLinkStyles = makeStyles(() => statusStyles)
const useStyles = makeStyles((theme) => ({
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
  iconContainer: {
    [theme.breakpoints.down('sm')]: { float: 'top' },
    [theme.breakpoints.up('md')]: { float: 'left' },
    margin: 20,
  },
  listSubheader: {
    margin: 0,
    fontSize: 'larger',
  },
  mainHeading: {
    fontSize: '1.5rem',
  },
  quickLinksContainer: {
    display: 'flex',
  },
  quickLinksList: {
    width: '100%',
  },
}))

function DetailsLink({ url, label, status, subText }) {
  const classes = useLinkStyles()
  const width = useWidth()

  let itemClass = classes.noStatus
  switch (status) {
    case 'ok':
      itemClass = classes.statusOK
      break
    case 'warn':
      itemClass = classes.statusWarning
      break
    case 'err':
      itemClass = classes.statusError
      break
  }

  return (
    <ListItem component={AppLink} to={url} button className={itemClass}>
      <ListItemText
        secondary={subText}
        primary={label}
        primaryTypographyProps={isDesktopMode(width) ? null : { variant: 'h5' }}
      />
      <ListItemSecondaryAction>
        <IconButton component={AppLink} to={url}>
          <ChevronRight />
        </IconButton>
      </ListItemSecondaryAction>
    </ListItem>
  )
}

DetailsLink.propTypes = {
  label: p.string.isRequired,
  url: p.string.isRequired,
  status: p.oneOf(['ok', 'warn', 'err']),
  subText: p.node,
}

export default function DetailsPage(props) {
  const classes = useStyles()
  const width = useWidth()

  const { title, details, icon, titleFooter, pageFooter } = props

  let links = null
  if (props.links && props.links.length) {
    links = (
      <List
        data-cy='route-links'
        className={classes.quickLinksList}
        subheader={
          isDesktopMode(width) ? (
            <ListSubheader
              className={classes.listSubheader}
              component='h2'
              color='primary'
            >
              Quick Links
            </ListSubheader>
          ) : null
        }
      >
        {isDesktopMode(width) ? <Divider /> : null}
        {props.links.map((li, idx) => (
          <DetailsLink key={idx} {...li} />
        ))}
      </List>
    )
  }

  return (
    <Grid container>
      <Grid item xs={12} className={classes.spacing}>
        <Card>
          <CardContent>
            <Grid container spacing={2}>
              <Grid item xs={isDesktopMode(width) && links ? 8 : 12}>
                {icon && <div className={classes.iconContainer}>{icon}</div>}
                <Typography
                  data-cy='details-heading'
                  className={classes.mainHeading}
                  component='h2'
                >
                  {title}
                </Typography>
                <Typography
                  data-cy='details'
                  variant='subtitle1'
                  component='div'
                >
                  <Markdown value={details} />
                </Typography>
                {titleFooter && (
                  <Typography
                    component='div'
                    variant='subtitle1'
                    data-cy='title-footer'
                  >
                    {titleFooter}
                  </Typography>
                )}
              </Grid>
              {links && (
                <Hidden smDown>
                  <Grid className={classes.quickLinksContainer} item xs={4}>
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
        <Grid item xs={12} className={classes.spacing}>
          <Card>{links}</Card>
        </Grid>
      </Hidden>
      {pageFooter && (
        <Grid item xs={12} className={classes.spacing}>
          {pageFooter}
        </Grid>
      )}
    </Grid>
  )
}
DetailsPage.propTypes = {
  title: p.string,
  details: p.string,

  icon: p.node,
  links: p.arrayOf(p.shape(DetailsLink.propTypes)),

  titleFooter: p.any,
  pageFooter: p.any,
}
