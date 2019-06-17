import React, { Component } from 'react'
import p from 'prop-types'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import Divider from '@material-ui/core/Divider'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import Grid from '@material-ui/core/Grid'
import Hidden from '@material-ui/core/Hidden'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import Switch from '@material-ui/core/Switch'
import Table from '@material-ui/core/Table'
import TableBody from '@material-ui/core/TableBody'
import TableCell from '@material-ui/core/TableCell'
import TableHead from '@material-ui/core/TableHead'
import TableRow from '@material-ui/core/TableRow'
import Typography from '@material-ui/core/Typography'
import withStyles from '@material-ui/core/styles/withStyles'
import isFullScreen from '@material-ui/core/withMobileDialog'
import moment from 'moment'
import Countdown from 'react-count-down'
import { Link } from 'react-router-dom'
import { ScheduleLink, ServiceLink, UserLink } from '../../links'
import { styles } from '../../styles/materialStyles'
import Options from '../../util/Options'
import gql from 'graphql-tag'
import PageActions from '../../util/PageActions'
import Markdown from '../../util/Markdown'

const localStorage = window.localStorage
const exactTimesKey = 'show_exact_times'

const sortTime = (a, b) => {
  const ma = moment(a.timestamp)
  const mb = moment(b.timestamp)
  if (ma.isSame(mb)) return 0
  return ma.isAfter(mb) ? -1 : 1
}

@withStyles(styles)
@isFullScreen()
export default class AlertDetails extends Component {
  static propTypes = {
    loading: p.bool,
    error: p.shape({ message: p.string }),
  }

  constructor(props) {
    super(props)

    // localstorage stores true/false as a string; convert to a bool
    // default to true if localstorage is not set
    let showExactTimes = localStorage.getItem(exactTimesKey) || false
    if (typeof showExactTimes !== 'boolean') {
      showExactTimes = showExactTimes === 'true'
    }

    this.state = {
      fullDescription: false,
      loading: '',
      escalateWaiting: false,
      showExactTimes,
    }
  }

  /*
   * Update state and local storage with new boolean value
   * telling whether or not the show exact times toggle is active
   */
  toggleExactTimes = () => {
    const newVal = !this.state.showExactTimes
    this.setState({
      showExactTimes: newVal,
    })
    localStorage.setItem(exactTimesKey, newVal.toString())
  }

  renderAlertLogEvents() {
    let logs = this.props.data.logs_2.slice(0).sort(sortTime)

    if (logs.length === 0) {
      return (
        <div>
          <Divider />
          <ListItem>
            <ListItemText primary='No events.' />
          </ListItem>
        </div>
      )
    }

    return logs.map((log, index) => {
      let alertTimeStamp = moment(log.timestamp)
        .local()
        .calendar()
      if (this.state.showExactTimes) {
        alertTimeStamp = moment(log.timestamp)
          .local()
          .format('MMM Do YYYY, h:mm:ss a')
      }
      return (
        <div key={index}>
          <Divider />
          <ListItem>
            <ListItemText primary={alertTimeStamp} secondary={log.message} />
          </ListItem>
        </div>
      )
    })
  }

  renderAlertLogs() {
    return (
      <Card className={this.getCardClassName()}>
        <div style={{ display: 'flex' }}>
          <CardContent style={{ flex: 1, paddingBottom: 0 }}>
            <Typography variant='h5'>Event Log</Typography>
          </CardContent>
          <FormControlLabel
            control={
              <Switch
                checked={this.state.showExactTimes}
                onChange={this.toggleExactTimes}
              />
            }
            label='Full Timestamps'
            style={{ padding: '0.5em 0.5em 0 0' }}
          />
        </div>
        <CardContent
          className={this.props.classes.tableCardContent}
          style={{ paddingBottom: 0 }}
        >
          <List>{this.renderAlertLogEvents()}</List>
        </CardContent>
      </Card>
    )
  }

  renderUsers(users, stepID) {
    return users.map((user, i) => {
      const sep = i === 0 ? '' : ', '
      return (
        <span key={stepID + user.id}>
          {sep}
          {UserLink(user)}
        </span>
      )
    })
  }

  renderSchedules(schedules, stepID) {
    return schedules.map((schedule, i) => {
      const sep = i === 0 ? '' : ', '
      return (
        <span key={stepID + schedule.id}>
          {sep}
          {ScheduleLink(schedule)}
        </span>
      )
    })
  }

  /*
   * Returns properties from the escalation policy
   * for easier use in functions.
   */
  epsHelper() {
    const eps = this.props.data.escalation_policy_snapshot
    const alert = this.props.data
    return {
      repeat: eps.repeat,
      numSteps: eps.steps.length,
      steps: eps.steps,
      status: alert.status.toLowerCase(),
      currentLevel: eps.current_level,
      lastEscalation: eps.last_escalation,
    }
  }

  canAutoEscalate() {
    const { repeat, numSteps, status, currentLevel } = this.epsHelper()
    if (status !== 'unacknowledged') return false
    if (repeat === -1) return true
    return currentLevel + 1 < numSteps * (repeat + 1)
  }

  /*
   * Renders a timer that counts down time until the next escalation
   */
  renderTimer(index, delayMinutes) {
    const { currentLevel, numSteps, lastEscalation } = this.epsHelper()
    const prevEscalation = new Date(lastEscalation)

    if (!this.canAutoEscalate()) {
      return <div>{delayMinutes} minutes</div>
    }

    if (currentLevel % numSteps === index) {
      return (
        <Countdown
          options={{
            endDate: new Date(prevEscalation.getTime() + delayMinutes * 60000),
          }}
        />
      )
    }

    return <div />
  }

  renderEscalationPolicySteps() {
    const { steps, status, currentLevel } = this.epsHelper()
    return steps.map((step, index) => {
      const { schedules, delay_minutes: delayMinutes, users } = step

      let usersRender
      if (users.length > 0) {
        usersRender = <div>Users: {this.renderUsers(users, step.id)}</div>
      }

      let schedulesRender
      if (schedules.length > 0) {
        schedulesRender = (
          <div>Schedules: {this.renderSchedules(schedules, step.id)}</div>
        )
      }

      let className
      if (status !== 'closed' && currentLevel % steps.length === index) {
        className = this.props.classes.highlightRow
      }
      return (
        <TableRow key={index} className={className}>
          <TableCell>Step #{index + 1}</TableCell>
          <TableCell>
            {usersRender}
            {schedulesRender}
          </TableCell>
          <TableCell>{this.renderTimer(index, delayMinutes)}</TableCell>
        </TableRow>
      )
    })
  }

  renderEscalationPolicy() {
    const alert = this.props.data

    return (
      <Card className={this.getCardClassName()} style={{ overflowX: 'auto' }}>
        <CardContent>
          <Typography variant='h5'>
            <Link
              to={`/escalation-policies/${alert.service.escalation_policy_id}`}
            >
              Escalation Policy
            </Link>
          </Typography>
        </CardContent>
        <CardContent className={this.props.classes.tableCardContent}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Step</TableCell>
                <TableCell>Alert</TableCell>
                <TableCell>
                  {this.canAutoEscalate()
                    ? 'Time Until Next Escalation'
                    : 'Time Between Escalations'}
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>{this.renderEscalationPolicySteps()}</TableBody>
          </Table>
        </CardContent>
      </Card>
    )
  }

  renderAlertDetails() {
    const alert = this.props.data
    let details = (alert.details || '').trim()
    if (!details) return null

    if (!this.state.fullDescription && details.length > 1000) {
      details = details.slice(0, 1000).trim() + ' ...'
    }

    let expandTextAction = null
    if (details.length > 1000) {
      let text = 'Show Less'

      if (!this.state.fullDescription) {
        text = 'Show More'
      }

      expandTextAction = (
        <Typography
          color='textSecondary'
          onClick={() => {
            this.setState({ fullDescription: !this.state.fullDescription })
          }}
          style={{
            display: 'flex',
            alignItems: 'center',
            cursor: 'pointer',
            justifyContent: 'center',
            textAlign: 'center',
            paddingTop: '1em',
          }}
        >
          {text}
        </Typography>
      )
    }

    return (
      <Grid
        item
        xs={12}
        data-cy='alert-details'
        className={this.props.classes.cardContainer}
      >
        <Card className={this.getCardClassName()}>
          <CardContent>
            <Typography variant='h5'>Details</Typography>
            <Typography
              variant='body1'
              style={{ whiteSpace: 'pre-wrap' }}
              component={Markdown}
              value={details}
            />
            {expandTextAction}
          </CardContent>
        </Card>
      </Grid>
    )
  }

  /*
   * Options to show for alert details menu
   */
  getMenuOptions = () => {
    const {
      escalation_level: escalationLevel,
      number: id,
      status,
    } = this.props.data
    if (status.toLowerCase() === 'closed') return [] // no options to show if alert is already closed

    const updateStatusMutation = gql`
      mutation UpdateAlertStatusMutation($input: UpdateAlertStatusInput!) {
        updateAlertStatus(input: $input) {
          id
          status: status_2
          logs_2 {
            event
            message
            timestamp
          }
        }
      }
    `

    let options = []
    const ack = {
      text: 'Acknowledge',
      mutation: {
        query: updateStatusMutation,
        variables: {
          input: {
            id,
            status_2: 'acknowledged',
          },
        },
      },
    }

    const esc = {
      text: 'Escalate',
      mutation: {
        query: gql`
          mutation EscalateAlertMutation($input: EscalateAlertInput!) {
            escalateAlert(input: $input) {
              id
              status: status_2
              logs_2 {
                event
                message
                timestamp
              }
            }
          }
        `,
        variables: {
          input: {
            id,
            current_escalation_level: escalationLevel,
          },
        },
      },
    }

    const close = {
      text: 'Close',
      mutation: {
        query: updateStatusMutation,
        variables: {
          input: {
            id,
            status_2: 'closed',
          },
        },
      },
    }

    options.push(esc)
    options.push(close)
    if (status.toLowerCase() === 'unacknowledged') options.push(ack)
    return options
  }

  getCardClassName = () => {
    const { classes, fullScreen } = this.props
    return fullScreen ? classes.cardFull : classes.card
  }

  render() {
    const { classes, data: alert } = this.props

    const options = this.getMenuOptions()
    const optionsMenu =
      options.length > 0 ? <Options options={options} /> : null

    return (
      <Grid container spacing={2}>
        <PageActions>{optionsMenu}</PageActions>
        <Grid item xs={12} className={classes.cardContainer}>
          <Card className={this.getCardClassName()}>
            <CardContent data-cy='alert-summary'>
              <Grid container spacing={1}>
                <Grid item xs={12}>
                  <Typography variant='body1'>
                    {ServiceLink(alert.service)}
                  </Typography>
                </Grid>
                <Grid item xs={12}>
                  <Typography variant='h5'>
                    {alert.number}: {alert.summary}
                  </Typography>
                </Grid>
                <Grid item xs={12}>
                  <Typography variant='body1'>
                    {alert.status.toUpperCase()}
                  </Typography>
                </Grid>
              </Grid>
            </CardContent>
          </Card>
        </Grid>
        {this.renderAlertDetails()}
        <Hidden smDown>
          <Grid item xs={12} className={classes.cardContainer}>
            {this.renderEscalationPolicy()}
          </Grid>
        </Hidden>
        <Grid item xs={12} className={classes.cardContainer}>
          {this.renderAlertLogs()}
        </Grid>
      </Grid>
    )
  }
}
