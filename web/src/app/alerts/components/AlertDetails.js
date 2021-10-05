import React, { useState } from 'react'
import p from 'prop-types'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import FormControlLabel from '@mui/material/FormControlLabel'
import Grid from '@mui/material/Grid'
import Hidden from '@mui/material/Hidden'
import Switch from '@mui/material/Switch'
import Table from '@mui/material/Table'
import TableBody from '@mui/material/TableBody'
import TableCell from '@mui/material/TableCell'
import TableHead from '@mui/material/TableHead'
import TableRow from '@mui/material/TableRow'
import Typography from '@mui/material/Typography'
import Countdown from 'react-countdown-now'
import {
  ArrowUpward as EscalateIcon,
  Check as AcknowledgeIcon,
  Close as CloseIcon,
} from '@mui/icons-material'
import { gql, useMutation } from '@apollo/client'
import { RotationLink, ScheduleLink, ServiceLink, UserLink } from '../../links'
import { styles } from '../../styles/materialStyles'
import Markdown from '../../util/Markdown'
import AlertDetailLogs from '../AlertDetailLogs'
import AppLink from '../../util/AppLink'
import makeStyles from '@mui/styles/makeStyles'
import { useIsWidthDown } from '../../util/useWidth'
import _ from 'lodash'
import CardActions from '../../details/CardActions'
import Notices from '../../details/Notices'

const useStyles = makeStyles((theme) => {
  return styles(theme)
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
function AlertDetails(props) {
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

  function renderAlertLogs() {
    return (
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
    )
  }

  function renderRotations(rotations, stepID) {
    return _.sortBy(rotations, 'name').map((rotation, i) => {
      const sep = i === 0 ? '' : ', '
      return (
        <span key={stepID + rotation.id}>
          {sep}
          {RotationLink(rotation)}
        </span>
      )
    })
  }

  function renderSchedules(schedules, stepID) {
    return _.sortBy(schedules, 'name').map((schedule, i) => {
      const sep = i === 0 ? '' : ', '
      return (
        <span key={stepID + schedule.id}>
          {sep}
          {ScheduleLink(schedule)}
        </span>
      )
    })
  }

  function renderUsers(users, stepID) {
    return _.sortBy(users, 'name').map((user, i) => {
      const sep = i === 0 ? '' : ', '
      return (
        <span key={stepID + user.id}>
          {sep}
          {UserLink(user)}
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
      repeat: state?.repeatCount,
      numSteps: ep.steps.length,
      steps: ep.steps,
      status: alert.status,
      currentLevel: state?.stepNumber,
      lastEscalation: state?.lastEscalation,
    }
  }

  function canAutoEscalate() {
    const { repeat, numSteps, status, currentLevel } = epsHelper()
    if (status !== 'StatusUnacknowledged') return false
    if (repeat === -1) return true
    return currentLevel + 1 < numSteps * (repeat + 1)
  }

  /*
   * Renders a timer that counts down time until the next escalation
   */
  function renderTimer(index, delayMinutes) {
    const { currentLevel, numSteps, lastEscalation } = epsHelper()
    const prevEscalation = new Date(lastEscalation)

    if (currentLevel % numSteps === index && canAutoEscalate()) {
      return (
        <Countdown
          date={new Date(prevEscalation.getTime() + delayMinutes * 60000)}
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
    return <Typography>&mdash;</Typography>
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
      const { delayMinutes, id, targets } = step

      const rotations = targets.filter((t) => t.type === 'rotation')
      const schedules = targets.filter((t) => t.type === 'schedule')
      const users = targets.filter((t) => t.type === 'user')

      let rotationsRender
      if (rotations.length > 0) {
        rotationsRender = <div>Rotations: {renderRotations(rotations, id)}</div>
      }

      let schedulesRender
      if (schedules.length > 0) {
        schedulesRender = <div>Schedules: {renderSchedules(schedules, id)}</div>
      }

      let usersRender
      if (users.length > 0) {
        usersRender = <div>Users: {renderUsers(users, id)}</div>
      }

      let className
      if (status !== 'closed' && currentLevel % steps.length === index) {
        className = classes.highlightRow
      }

      return (
        <TableRow key={index} className={className}>
          <TableCell>Step #{index + 1}</TableCell>
          <TableCell>
            {!targets.length && <Typography>&mdash;</Typography>}
            {rotationsRender}
            {schedulesRender}
            {usersRender}
          </TableCell>
          <TableCell>{renderTimer(index, delayMinutes)}</TableCell>
        </TableRow>
      )
    })
  }

  function renderEscalationPolicy() {
    const alert = props.data

    return (
      <Card className={getCardClassName()} style={{ overflowX: 'auto' }}>
        <CardContent>
          <Typography component='h3' variant='h5'>
            <AppLink
              to={`/escalation-policies/${alert.service.escalationPolicy.id}`}
            >
              Escalation Policy
            </AppLink>
          </Typography>
        </CardContent>
        <CardContent className={classes.tableCardContent}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Step</TableCell>
                <TableCell>Alert</TableCell>
                <TableCell>
                  {canAutoEscalate()
                    ? 'Time Until Next Escalation'
                    : 'Time Between Escalations'}
                </TableCell>
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
    )
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
      <Hidden mdDown>
        <Grid item xs={12} className={classes.cardContainer}>
          {renderEscalationPolicy()}
        </Grid>
      </Hidden>
      <Grid item xs={12} className={classes.cardContainer}>
        {renderAlertLogs()}
      </Grid>
    </Grid>
  )
}

AlertDetails.propTypes = {
  error: p.shape({ message: p.string }),
}

export default AlertDetails
