import React from 'react'
import p from 'prop-types'
import { connect } from 'react-redux'
import { absURLSelector } from '../selectors/url'
import statusStyles from '../util/statusStyles'
import withStyles from '@material-ui/core/styles/withStyles'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import Grid from '@material-ui/core/Grid'
import Typography from '@material-ui/core/Typography'
import { Link } from 'react-router-dom'
import { ChevronRight } from '@material-ui/icons'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction'
import IconButton from '@material-ui/core/IconButton'
import Markdown from '../util/Markdown'

const styles = theme => ({
  ...statusStyles,
  spacing: {
    '&:not(:first-child)': {
      marginTop: 8,
    },
    marginBottom: 8,
  },
  iconContainer: {
    [theme.breakpoints.down('sm')]: { float: 'top' },
    [theme.breakpoints.up('md')]: { float: 'left' },
    margin: 20,
  },
  mainHeading: {
    fontSize: '1.5rem',
  },
})

const mapStateToProps = state => {
  return {
    absURL: absURLSelector(state),
  }
}

@withStyles(styles)
@connect(mapStateToProps)
export default class DetailsPage extends React.PureComponent {
  static propTypes = {
    title: p.string.isRequired,
    details: p.string.isRequired,

    icon: p.node,
    links: p.arrayOf(
      p.shape({
        label: p.string.isRequired,
        url: p.string.isRequired,
        status: p.oneOf(['ok', 'warn', 'err']),
        subText: p.node,
      }),
    ),

    titleFooter: p.any,
    pageFooter: p.any,
  }

  renderLink = ({ url, label, status, subText }, idx) => {
    const { classes, absURL } = this.props
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
      <ListItem
        key={idx}
        component={Link}
        to={absURL(url)}
        button
        className={itemClass}
      >
        <ListItemText
          secondary={subText}
          primary={label}
          primaryTypographyProps={{ variant: 'h5' }}
        />
        <ListItemSecondaryAction>
          <IconButton>
            <ChevronRight />
          </IconButton>
        </ListItemSecondaryAction>
      </ListItem>
    )
  }

  renderLinks() {
    const { links } = this.props

    if (!links || !links.length) return null

    return (
      <Grid item xs={12} className={this.props.classes.spacing}>
        <Card>
          <List>{links.map(this.renderLink)}</List>
        </Card>
      </Grid>
    )
  }

  render() {
    const {
      title,
      details,
      icon,
      titleFooter,
      pageFooter,
      classes,
    } = this.props
    return (
      <Grid item container>
        <Grid item xs={12} className={classes.spacing}>
          <Card>
            <CardContent>
              {icon && <div className={classes.iconContainer}>{icon}</div>}
              <Typography
                data-cy='details-heading'
                className={classes.mainHeading}
                component='h2'
              >
                {title}
              </Typography>
              <Typography data-cy='details' variant='subtitle1' component='div'>
                <Markdown value={details} />
              </Typography>
              {titleFooter && (
                <Typography
                  component='p'
                  variant='subtitle1'
                  data-cy='title-footer'
                >
                  {titleFooter}
                </Typography>
              )}
            </CardContent>
          </Card>
        </Grid>

        {this.renderLinks()}

        {pageFooter && (
          <Grid item xs={12} className={classes.spacing}>
            {pageFooter}
          </Grid>
        )}
      </Grid>
    )
  }
}
