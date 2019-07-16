import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import CardHeader from '@material-ui/core/CardHeader'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemAvatar from '@material-ui/core/ListItemAvatar'
import ListItemText from '@material-ui/core/ListItemText'
import Typography from '@material-ui/core/Typography'
import { UserAvatar } from '../../util/avatar'
import { Link } from 'react-router-dom'
import Spinner from '../../loading/components/Spinner'
import { withStyles } from '@material-ui/core'
import { styles as globalStyles } from '../../styles/materialStyles'

const styles = theme => {
  const { cardHeader } = globalStyles(theme)

  return {
    cardHeader,
  }
}

@withStyles(styles)
export default class OnCallForService extends Component {
  static propTypes = {
    onCallUsers: p.arrayOf(
      p.shape({
        stepNumber: p.number.isRequired,
        userID: p.string.isRequired,
        userName: p.string.isRequired,
      }),
    ),
  }

  /*
   * Handles if a user is on multiple steps
   * creates a dict for each user on call
   * and creates an array of steps each user
   * is on
   */
  getUsersDict = users => {
    if (!users) return {}
    let usersDict = {}

    users.forEach(x => {
      // if duplicate found add step # to steps array of existing key
      if (x.userID in usersDict) {
        usersDict[x.userID]['steps'].push(`#${x.stepNumber + 1}`)
        return
      }

      // not in dict yet, create key and add single step #
      usersDict[x.userID] = {
        name: x.userName,
        steps: [`#${x.stepNumber + 1}`],
      }
    })

    return usersDict
  }

  /*
   * Given an array of steps, returns a string with proper sentence grammar
   */
  stepsText = steps => {
    if (steps.length === 1) {
      return steps[0]
    }

    const copy = [...steps]
    const last = copy.pop()
    return copy.join(', ') + (steps.length > 3 ? ', and ' : ' and ') + last
  }

  renderUsers() {
    const usersDict = this.getUsersDict(this.props.onCallUsers)

    if (!Object.keys(usersDict).length) {
      return (
        <Typography variant='caption'>
          No users on call for this service
        </Typography>
      )
    }

    return (
      <List>
        {Object.keys(usersDict).map(id => {
          const step = usersDict[id]['steps'].length > 1 ? 'Steps' : 'Step'

          return (
            <ListItem key={id} button component={Link} to={`/users/${id}`}>
              <ListItemAvatar>
                <UserAvatar userID={id} />
              </ListItemAvatar>
              <ListItemText
                primary={usersDict[id]['name']}
                secondary={`${step} ${this.stepsText(usersDict[id]['steps'])}`}
              />
            </ListItem>
          )
        })}
      </List>
    )
  }

  render() {
    let content = this.props.loading ? <Spinner /> : this.renderUsers()
    let { classes } = this.props

    return (
      <Card>
        <CardHeader
          className={classes.cardHeader}
          component='h3'
          title='On Call Users'
        />
        <CardContent>{content}</CardContent>
      </Card>
    )
  }
}
