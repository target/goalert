import React, { useState, ReactNode } from 'react'
import p from 'prop-types'
import ButtonGroup from '@mui/material/ButtonGroup'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import FormControlLabel from '@mui/material/FormControlLabel'
import Grid from '@mui/material/Grid'
import Switch from '@mui/material/Switch'
import Table from '@mui/material/Table'
import TableBody from '@mui/material/TableBody'
import TableCell from '@mui/material/TableCell'
import TableHead from '@mui/material/TableHead'
import TableRow from '@mui/material/TableRow'
import Typography from '@mui/material/Typography'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'
import {
  ArrowUpward as EscalateIcon,
  Check as AcknowledgeIcon,
  Close as CloseIcon,
} from '@mui/icons-material'
import { gql, useMutation } from '@apollo/client'
import { DateTime } from 'luxon'
import { ServiceLink } from '../../links'
import { styles as globalStyles } from '../../styles/materialStyles'
import Markdown from '../../util/Markdown'
import AlertDetailLogs from '../AlertDetailLogs'
import AppLink from '../../util/AppLink'
import CardActions from '../../details/CardActions'
import { Alert, EscalationPolicyStep, AlertStatus } from '../../../schema'
import ServiceNotices from '../../services/ServiceNotices'
import { Time } from '../../util/Time'
import AlertFeedback, {
  mutation as undoFeedbackMutation,
} from './AlertFeedback'
import LoadingButton from '../../loading/components/LoadingButton'
import { Notice } from '../../details/Notices'
import { useIsWidthDown } from '../../util/useWidth'
import ReactGA from 'react-ga4'
import { useConfigValue } from '../../util/RequireConfig'
import { renderChipsDest } from '../../escalation-policies/stepUtil'
interface AlertDetailsProps {
  data: Alert
}

interface EscalationPolicyInfo {
  repeatCount?: number
  repeat?: number
  numSteps?: number
  steps?: EscalationPolicyStep[]
  status: AlertStatus
  currentLevel?: number
  lastEscalation?: string
}

const useStyles = makeStyles((theme: Theme) => ({
  card: globalStyles(theme).card,
  cardContainer: globalStyles(theme).cardContainer,
  cardFull: globalStyles(theme).cardFull,
  tableCardContent: globalStyles(theme).tableCardContent,
  epHeader: {
    paddingBottom: 8,
  },
}))

const localStorage = window.localStorage
const exactTimesKey = 'show_exact_times'

const updateStatusMutation = gql`
  mutation UpdateAlertsMutation($input: UpdateAlertsInput!) {
    updateAlerts(input: $input) {
      id
    }
  }
`

export default function AlertDetails(props: AlertDetailsProps): JSX.Element {
  const [analyticsID] = useConfigValue('General.GoogleAnalyticsID') as [string]
  const classes = useStyles()
  const isMobile = useIsWidthDown('sm')

  const alertAction = (action: string, mutation: () => void): void => {
    if (analyticsID) ReactGA.event({ category: 'Alert Action', action })
    mutation()
  }

  const [undoFeedback, undoFeedbackStatus] = useMutation(undoFeedbackMutation, {
    variables: {
      input: {
        alertID: props.data.id,
        noiseReason: '',
      },
    },
  })

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
  function handleToggleExactTimes(): void {
    const newVal = !showExactTimes
    setShowExactTimes(newVal)
    localStorage.setItem(exactTimesKey, newVal.toString())
  }

  /*
   * Returns properties from the escalation policy
   * for easier use in functions.
   */
  function epsHelper(): EscalationPolicyInfo {
    const ep = props.data?.service?.escalationPolicy
    const alert = props.data
    const state = props.data.state

    return {
      repeatCount: state?.repeatCount,
      repeat: ep?.repeat,
      numSteps: ep?.steps.length,
      steps: ep?.steps,
      status: alert.status,
      currentLevel: state?.stepNumber,
      lastEscalation: state?.lastEscalation,
    }
  }

  function canAutoEscalate(): boolean {
    const { currentLevel, status, steps, repeat, repeatCount } = epsHelper()

    if (status !== 'StatusUnacknowledged') {
      return false
    }

    if (
      currentLevel === (steps?.length ?? 0) - 1 &&
      (repeatCount ?? 0) >= (repeat ?? 0)
    ) {
      return false
    }

    return true
  }

  function getNextEscalation(): JSX.Element | string {
    const { currentLevel, lastEscalation, steps } = epsHelper()
    if (!canAutoEscalate()) return 'None'

    const prevEscalation = DateTime.fromISO(lastEscalation ?? '')
    const nextEsclation = prevEscalation.plus({
      minutes: steps ? steps[currentLevel ?? 0].delayMinutes : 0,
    })

    return (
      <Time
        time={nextEsclation}
        format='relative'
        units={['hours', 'minutes', 'seconds']}
        precise
      />
    )
  }

  function renderEscalationPolicySteps(): JSX.Element[] | JSX.Element {
    const { steps, status, currentLevel } = epsHelper()

    if (!steps?.length) {
      return (
        <TableRow>
          <TableCell>No steps</TableCell>
          <TableCell>&mdash;</TableCell>
          <TableCell>&mdash;</TableCell>
        </TableRow>
      )
    }

    return steps.map((step, index) => {
      const selected =
        status !== 'StatusClosed' &&
        (currentLevel ?? 0) % steps.length === index

      return (
        <TableRow key={index} selected={selected}>
          <TableCell>Step #{index + 1}</TableCell>
          <TableCell>{renderChipsDest(step.actions)}</TableCell>
        </TableRow>
      )
    })
  }

  function renderAlertDetails(): ReactNode {
    const alert = props.data
    let details = (alert.details || '').trim()
    if (!details) return null

    if (!fullDescription && details.length > 1000) {
      details = details.slice(0, 1000).trim() + ' ...'
    }

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
        <Card sx={{ width: '100%' }}>
          <CardContent>
            <Typography component='h3' variant='h5'>
              Details
            </Typography>
            <Typography variant='body1' component='div'>
              <Markdown value={details + '\n'} />
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
  function getMenuOptions(): Array<JSX.Element> {
    const { status } = props.data
    if (status === 'StatusClosed') return []
    const isMaintMode = Boolean(props.data?.service?.maintenanceExpiresAt)

    return [
      <ButtonGroup
        key='update-alert-buttons'
        variant='contained'
        aria-label='Update Alert Status Button Group'
      >
        {status === 'StatusUnacknowledged' && (
          <Button
            startIcon={<AcknowledgeIcon />}
            onClick={() => alertAction('alert_acknowledged', ack)}
          >
            Acknowledge
          </Button>
        )}
        <Button
          startIcon={<EscalateIcon />}
          onClick={() => alertAction('alert_escalated', escalate)}
          disabled={isMaintMode}
        >
          Escalate
        </Button>
        <Button
          startIcon={<CloseIcon />}
          onClick={() => alertAction('alert_closed', close)}
        >
          Close
        </Button>
      </ButtonGroup>,
    ]
  }

  const { data: alert } = props

  let extraNotices: Notice[] = alert.pendingNotifications.map((n) => ({
    type: 'WARNING',
    message: `Notification Pending for ${n.destination}`,
    details:
      'This could be due to rate-limiting, processing, or network delays.',
  }))

  const noiseReason = alert?.noiseReason ?? ''
  if (noiseReason !== '') {
    const nrArr = noiseReason.split('|')
    const reasons = nrArr.join(', ')
    extraNotices = [
      ...extraNotices,
      {
        type: 'INFO',
        message: 'This alert has been marked as noise',
        details: `Reason${nrArr.length > 1 ? 's' : ''}: ${reasons}`,
        action: (
          <LoadingButton
            buttonText='Undo'
            aria-label='Reset noise reasons'
            variant='text'
            loading={undoFeedbackStatus.called && undoFeedbackStatus.loading}
            onClick={() => undoFeedback()}
          />
        ),
      },
    ]
  }

  return (
    <Grid container spacing={2}>
      <ServiceNotices
        serviceID={alert?.service?.id ?? ''}
        extraNotices={extraNotices as Notice[]}
      />

      {/* Main Alert Info */}
      <Grid
        item
        lg={isMobile || noiseReason !== '' ? 12 : 8}
        className={classes.cardContainer}
      >
        <Card
          sx={{
            width: '100%',
            height: '100%',
            flexDirection: 'column',
            display: 'flex',
            justifyContent: 'space-between',
          }}
        >
          <CardContent data-cy='alert-summary'>
            <Grid container spacing={1}>
              {alert.service && (
                <Grid item xs={12}>
                  <Typography variant='body1'>
                    {ServiceLink(alert.service)}
                  </Typography>
                </Grid>
              )}
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
          <CardActions primaryActions={getMenuOptions()} />
        </Card>
      </Grid>
      {!noiseReason && (
        <Grid item xs={12} lg={isMobile ? 12 : 4}>
          <AlertFeedback alertID={alert.alertID} />
        </Grid>
      )}
      {renderAlertDetails()}

      {/* Escalation Policy Info */}
      <Grid item xs={12} className={classes.cardContainer}>
        <Card style={{ width: '100%', overflowX: 'auto' }}>
          <CardContent>
            <Typography
              className={classes.epHeader}
              component='h3'
              variant='h5'
            >
              <AppLink
                to={`/escalation-policies/${alert.service?.escalationPolicy?.id}`}
              >
                Escalation Policy
              </AppLink>
            </Typography>
            {alert?.state?.lastEscalation && (
              <React.Fragment>
                <Typography color='textSecondary' variant='caption'>
                  Last Escalated: <Time time={alert.state.lastEscalation} />
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

      {/* Alert Logs */}
      <Grid item xs={12} className={classes.cardContainer}>
        <Card sx={{ width: '100%' }}>
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
