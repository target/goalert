import React, { useState } from 'react'
import p from 'prop-types'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import Grid from '@material-ui/core/Grid'
import Hidden from '@material-ui/core/Hidden'
import Switch from '@material-ui/core/Switch'
import Table from '@material-ui/core/Table'
import TableBody from '@material-ui/core/TableBody'
import TableCell from '@material-ui/core/TableCell'
import TableHead from '@material-ui/core/TableHead'
import TableRow from '@material-ui/core/TableRow'
import Typography from '@material-ui/core/Typography'
import {
  ArrowUpward as EscalateIcon,
  Check as AcknowledgeIcon,
  Close as CloseIcon,
} from '@material-ui/icons'
import { gql, useMutation } from '@apollo/client'
import {
  RotationLink,
  ScheduleLink,
  ServiceLink,
  SlackChannelLink,
  UserLink,
} from '../../links'
import { styles } from '../../styles/materialStyles'
import Markdown from '../../util/Markdown'
import AlertDetailLogs from '../AlertDetailLogs'
import AppLink from '../../util/AppLink'
import { makeStyles } from '@material-ui/core'
import { useIsWidthDown } from '../../util/useWidth'
import _ from 'lodash'
import CardActions from '../../details/CardActions'
import Notices from '../../details/Notices'
import { DateTime } from 'luxon'
import Countdown from 'react-countdown'

const useStyles = makeStyles((theme) => {
  return {
    ...styles(theme),
    epHeader: {
      paddingBottom: 8,
    },
  }
})

const localStorage = window.localStorage
const exactTimesKey = 'show_exact_times'

const updateStatusMutation = gql`
  mutation UpdateAlertsMutation($input: UpdateAlertsInput!) {
    updateAlerts(input: $input) {
      id
    }
  }
`

export default function AlertDetails(props) {
  const classes = useStyles()
  const fullScreen = useIsWidthDown('md')

  const [ack] = useMutation(updateStatusMutation, {
    variables: {
      input: {
        alertIDs: [props.data.id],
        newStatus: 'StatusAcknowledged',
      },
    },
  })
  const [close] = useMutation(updateStatusMutation, {
    variables: {
      input: {
        alertIDs: [props.data.id],
        newStatus: 'StatusClosed',
      },
    },
  })
  const [escalate] = useMutation(
    gql`
      mutation EscalateAlertMutation($input: [Int!]) {
        escalateAlerts(input: $input) {
          id
        }
      }
    `,
    {
      variables: {
        input: [props.data.id],
      },
    },
  )

  // localstorage stores true/false as a string; convert to a bool
  // default to true if localstorage is not set
  let _showExactTimes = localStorage.getItem(exactTimesKey) || false
  if (typeof _showExactTimes !== 'boolean') {
    _showExactTimes = _showExactTimes === 'true'
  }

  const [fullDescription, setFullDescription] = useState(false)
  const [showExactTimes, setShowExactTimes] = useState(_showExactTimes)

  /*
   * Update state and local storage with new boolean value
   * telling whether or not the show exact times toggle is active
   */
  function handleToggleExactTimes() {
    const newVal = !showExactTimes
    setShowExactTimes(newVal)
    localStorage.setItem(exactTimesKey, newVal.toString())
  }

  function getCardClassName() {
    return fullScreen ? classes.cardFull : classes.card
  }

  function renderTargets(targets, stepID) {
    return _.sortBy(targets, 'name').map((target, i) => {
      const separator = i === 0 ? '' : ', '

      let link
      const t = target.type
      if (t === 'rotation') link = RotationLink(target)
      else if (t === 'schedule') link = ScheduleLink(target)
      else if (t === 'slackChannel') link = SlackChannelLink(target)
      else if (t === 'user') link = UserLink(target)
      else link = target.name

      return (
        <span key={stepID + target.id}>
          {separator}
          {link}
        </span>
      )
    })
  }

  /*
   * Returns properties from the escalation policy
   * for easier use in functions.
   */
  function epsHelper() {
    const ep = props.data.service.escalationPolicy
    const alert = props.data
    const state = props.data.state

    return {
      repeatCount: state?.repeatCount,
      repeat: ep.repeat,
      numSteps: ep.steps.length,
      steps: ep.steps,
      status: alert.status,
      currentLevel: state?.stepNumber,
      lastEscalation: state?.lastEscalation,
    }
  }

  function canAutoEscalate() {
    const { currentLevel, status, steps, repeat, repeatCount } = epsHelper()

    if (status !== 'StatusUnacknowledged') {
      return false
    }

    if (currentLevel === steps.length - 1 && repeat === repeatCount) {
      return false
    }

    return true
  }

  function getNextEscalation() {
    const { currentLevel, lastEscalation, steps } = epsHelper()
    const prevEscalation = new Date(lastEscalation)

    if (canAutoEscalate()) {
      return (
        <Countdown
          date={
            new Date(
              prevEscalation.getTime() +
                steps[currentLevel].delayMinutes * 60000,
            )
          }
          renderer={(props) => {
            const { hours, minutes, seconds } = props

            const hourTxt = parseInt(hours)
              ? `${hours} hour${parseInt(hours) === 1 ? '' : 's'} `
              : ''
            const minTxt = parseInt(minutes)
              ? `${minutes} minute${parseInt(minutes) === 1 ? '' : 's'} `
              : ''
            const secTxt = `${seconds} second${
              parseInt(seconds) === 1 ? '' : 's'
            }`

            return hourTxt + minTxt + secTxt
          }}
        />
      )
    }

    return 'None'
  }

  function renderEscalationPolicySteps() {
    const { steps, status, currentLevel } = epsHelper()

    if (!steps.length) {
      return (
        <TableRow>
          <TableCell>No steps</TableCell>
          <TableCell>&mdash;</TableCell>
          <TableCell>&mdash;</TableCell>
        </TableRow>
      )
    }

    return steps.map((step, index) => {
      const { id, targets } = step

      const rotations = targets.filter((t) => t.type === 'rotation')
      const schedules = targets.filter((t) => t.type === 'schedule')
      const users = targets.filter((t) => t.type === 'user')
      const slackChannels = targets.filter((t) => t.type === 'slackChannel')

      let className
      if (status !== 'closed' && currentLevel % steps.length === index) {
        className = classes.highlightRow
      }

      return (
        <TableRow key={index} className={className}>
          <TableCell>Step #{index + 1}</TableCell>
          <TableCell>
            {!targets.length && <Typography>&mdash;</Typography>}
            {rotations.length > 0 && (
              <div>Rotations: {renderTargets(rotations, id)}</div>
            )}
            {schedules.length > 0 && (
              <div>Schedules: {renderTargets(schedules, id)}</div>
            )}
            {slackChannels.length > 0 && (
              <div>Slack Channels: {renderTargets(slackChannels, id)}</div>
            )}
            {users.length > 0 && <div>Users: {renderTargets(users, id)}</div>}
          </TableCell>
        </TableRow>
      )
    })
  }

  function renderAlertDetails() {
    const alert = props.data
    let details = (alert.details || '').trim()
    if (!details) return null

    if (!fullDescription && details.length > 1000) {
      details = details.slice(0, 1000).trim() + ' ...'
    }
    if (details.split('```').length % 2 === 0) details += '\n```'

    let expandTextAction = null
    if (details.length > 1000) {
      let text = 'Show Less'

      if (!fullDescription) {
        text = 'Show More'
      }

      expandTextAction = (
        <Typography
          color='textSecondary'
          onClick={() => setFullDescription(!fullDescription)}
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
        className={classes.cardContainer}
      >
        <Card className={getCardClassName()}>
          <CardContent>
            <Typography component='h3' variant='h5'>
              Details
            </Typography>
            <Typography
              variant='body1'
              component='div'
              style={{ whiteSpace: 'pre-wrap' }}
            >
              <Markdown value={details} />
            </Typography>
            {expandTextAction}
          </CardContent>
        </Card>
      </Grid>
    )
  }

  /*
   * Options to show for alert details menu
   */
  function getMenuOptions() {
    const { status } = props.data
    let options = []

    if (status === 'StatusClosed') return options
    if (status === 'StatusUnacknowledged') {
      options = [
        {
          icon: <AcknowledgeIcon />,
          label: 'Acknowledge',
          handleOnClick: () => ack(),
        },
      ]
    }

    // only remaining status is acknowledged, show remaining buttons
    return [
      ...options,
      {
        icon: <CloseIcon />,
        label: 'Close',
        handleOnClick: () => close(),
      },
      {
        icon: <EscalateIcon />,
        label: 'Escalate',
        handleOnClick: () => escalate(),
      },
    ]
  }

  const { data: alert } = props

  const notices = alert.pendingNotifications.map((n) => ({
    type: 'WARNING',
    message: `Notification Pending for ${n.destination}`,
    details:
      'This could be due to rate-limiting, processing, or network delays.',
  }))

  return (
    <Grid container spacing={2} justifyContent='center'>
      <Grid item className={getCardClassName()}>
        <Notices notices={notices} />
      </Grid>

      {/* Main Alert Info */}
      <Grid item xs={12} className={classes.cardContainer}>
        <Card className={getCardClassName()}>
          <CardContent data-cy='alert-summary'>
            <Grid container spacing={1}>
              <Grid item xs={12}>
                <Typography variant='body1'>
                  {ServiceLink(alert.service)}
                </Typography>
              </Grid>
              <Grid item xs={12}>
                <Typography component='h2' variant='h5'>
                  {alert.alertID}: {alert.summary}
                </Typography>
              </Grid>
              <Grid item xs={12}>
                <Typography variant='body1' data-cy='alert-status'>
                  {alert.status.toUpperCase().replace('STATUS', '')}
                </Typography>
              </Grid>
            </Grid>
          </CardContent>
          <CardActions secondaryActions={getMenuOptions()} />
        </Card>
      </Grid>
      {renderAlertDetails()}

      {/* Escalation Policy Info */}
      <Hidden smDown>
        <Grid item xs={12} className={classes.cardContainer}>
          <Card className={getCardClassName()} style={{ overflowX: 'auto' }}>
            <CardContent>
              <Typography
                className={classes.epHeader}
                component='h3'
                variant='h5'
              >
                <AppLink
                  to={`/escalation-policies/${alert.service.escalationPolicy.id}`}
                >
                  Escalation Policy
                </AppLink>
              </Typography>
              {alert.state !== null && (
                <React.Fragment>
                  <Typography color='textSecondary' variant='caption'>
                    Last Escalated:{' '}
                    {DateTime.fromISO(alert.state.lastEscalation).toFormat(
                      'fff',
                    )}
                  </Typography>
                  <br />
                  <Typography color='textSecondary' variant='caption'>
                    Next Escalation: {getNextEscalation()}
                  </Typography>
                </React.Fragment>
              )}
            </CardContent>
            <CardContent className={classes.tableCardContent}>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Step</TableCell>
                    <TableCell>Alert</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>{renderEscalationPolicySteps()}</TableBody>
              </Table>
            </CardContent>
            <CardContent>
              <Typography color='textSecondary' variant='caption'>
                Visit this escalation policy for more information.
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Hidden>

      {/* Alert Logs */}
      <Grid item xs={12} className={classes.cardContainer}>
        <Card className={getCardClassName()}>
          <div style={{ display: 'flex' }}>
            <CardContent style={{ flex: 1, paddingBottom: 0 }}>
              <Typography component='h3' variant='h5'>
                Event Log
              </Typography>
            </CardContent>
            <FormControlLabel
              control={
                <Switch
                  checked={showExactTimes}
                  onChange={handleToggleExactTimes}
                />
              }
              label='Full Timestamps'
              style={{ padding: '0.5em 0.5em 0 0' }}
            />
          </div>
          <CardContent
            className={classes.tableCardContent}
            style={{ paddingBottom: 0 }}
          >
            <AlertDetailLogs
              alertID={props.data.alertID}
              showExactTimes={showExactTimes}
            />
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  )
}

AlertDetails.propTypes = {
  error: p.shape({ message: p.string }),
}
