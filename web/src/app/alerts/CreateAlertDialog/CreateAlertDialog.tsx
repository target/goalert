import React, { useState, useEffect } from 'react'
import {
  Button,
  Grid,
  Stepper,
  Step,
  StepLabel,
  Typography,
  Theme,
} from '@mui/material'
import { makeStyles } from '@mui/styles'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'
import _ from 'lodash'
import { useCreateAlerts } from './useCreateAlerts'
import { fieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import { CreateAlertForm } from './StepContent/CreateAlertForm'
import { CreateAlertReview } from './StepContent/CreateAlertReview'
import AppLink from '../../util/AppLink'

export interface Value {
  summary: string
  details: string
  serviceIDs: string[]
}

const pluralize = (num: number): string => (num !== 1 ? 's' : '')

const useStyles = makeStyles((theme: Theme) => ({
  dialog: {
    [theme.breakpoints.up('md')]: {
      height: '65vh',
    },
  },
  flexGrow: {
    flexGrow: 1,
  },
}))

export default function CreateAlertDialog(props: {
  onClose: () => void
  serviceID?: string
}): React.JSX.Element {
  const classes = useStyles()
  const [step, setStep] = useState(0)
  const serviceID = props.serviceID

  const [value, setValue] = useState<Value>({
    summary: '',
    details: '',
    serviceIDs: serviceID ? [serviceID] : [],
  })
  const [mutate, { data, loading, error }, getSvcID] = useCreateAlerts(value)

  const fieldErrs = fieldErrors(error)
  const hasValidationError = fieldErrs.some((e) =>
    ['summary', 'details'].includes(e.field),
  )

  useEffect(() => {
    if (hasValidationError) {
      setStep(0)
    }
  }, [hasValidationError])

  const hasCompleted = Boolean(data) && !hasValidationError
  const currentStep = loading ? 2 : step

  const stepTitles = serviceID
    ? ['Alert Info', 'Confirm']
    : ['Alert Info', 'Service Selection', 'Confirm']

  const onNext = (): void => {
    if (currentStep === 0 && serviceID) {
      setStep(currentStep + 2)
    } else if (currentStep < 2) {
      setStep(currentStep + 1)
    }
  }

  let review, reviewTitle
  if (hasCompleted) {
    const createdAlertIDs = _.chain(data)
      .values()
      .filter()
      .map((a) => a.id)
      .value()

    const failedServices = fieldErrors(error).map((e) => ({
      id: getSvcID(e.path ? e.path[1] : ''),
      message: e.message,
    }))

    const failMessage = failedServices.length
      ? `(${failedServices.length} failed)`
      : ''

    const titleMessage = `Successfully created ${
      createdAlertIDs.length
    } alert${pluralize(createdAlertIDs.length)} ${failMessage}`

    reviewTitle = (
      <Grid container>
        <Grid item>
          <Typography>{titleMessage}</Typography>
        </Grid>
        <Grid item className={classes.flexGrow} />
        <Grid item>
          <Button
            variant='contained'
            size='small'
            component={AppLink}
            endIcon={<OpenInNewIcon />}
            to={`/alerts?allServices=1&filter=all&search=${encodeURIComponent(
              value.summary,
            )}`}
            newTab
          >
            Monitor Alerts
          </Button>
        </Grid>
      </Grid>
    )

    review = (
      <CreateAlertReview
        createdAlertIDs={createdAlertIDs}
        failedServices={failedServices}
      />
    )
  }

  return (
    <FormDialog
      title='Create New Alert'
      alert={hasCompleted}
      primaryActionLabel={hasCompleted ? 'Done' : null}
      onClose={props.onClose}
      loading={loading}
      subTitle={
        reviewTitle || (
          <Stepper activeStep={currentStep}>
            {stepTitles.map((title, idx) => (
              <Step key={idx}>
                <StepLabel>{title}</StepLabel>
              </Step>
            ))}
          </Stepper>
        )
      }
      form={
        review || (
          <CreateAlertForm
            activeStep={currentStep}
            value={value}
            onChange={(newValue: Value) => setValue(newValue)}
            disabled={loading}
            errors={fieldErrors(error)}
          />
        )
      }
      PaperProps={{ className: classes.dialog }}
      onSubmit={() => (hasCompleted ? props.onClose() : mutate())}
      disableNext={currentStep === 2}
      onNext={onNext}
      onBack={currentStep > 0 ? () => setStep(currentStep - 1) : null}
    />
  )
}
