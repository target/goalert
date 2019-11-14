import React, { useState, useEffect } from 'react'
import {
  makeStyles,
  Button,
  Grid,
  Stepper,
  Step,
  StepLabel,
  Typography,
  Link,
} from '@material-ui/core'
import OpenInNewIcon from '@material-ui/icons/OpenInNew'

import { useCreateAlerts } from './useCreateAlerts'
import { fieldErrors, allErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import { CreateAlertForm } from './StepContent/CreateAlertForm'
import { CreateAlertReview } from './StepContent/CreateAlertReview'

import _ from 'lodash-es'

const stepTitles = ['Alert Info', 'Service Selection', 'Confirm']
const pluralize = num => (num !== 1 ? 's' : '')

const useStyles = makeStyles(theme => ({
  dialog: {
    [theme.breakpoints.up('md')]: {
      height: '65vh',
    },
  },
  flexGrow: {
    flexGrow: 1,
  },
}))

export default function CreateAlertDialog(props) {
  const classes = useStyles()
  const [step, setStep] = useState(0)
  const [value, setValue] = useState({
    summary: '',
    details: '',
    serviceIDs: props.serviceID ? [props.serviceID] : [],
  })
  const [mutate, { data, loading, error }, getSvcID] = useCreateAlerts(value)

  const fieldErrs = fieldErrors(error)
  const hasValidationError = fieldErrs.some(e =>
    ['summary', 'details'].includes(e.field),
  )

  useEffect(() => {
    if (hasValidationError) {
      setStep(0)
    }
  }, [hasValidationError])

  const hasSubmitted = Boolean(data) && !hasValidationError
  const currentStep = loading ? 2 : step

  let review, reviewTitle
  if (hasSubmitted) {
    const createdAlertIDs = _.chain(data)
      .values()
      .filter()
      .map(a => a.id)
      .value()

    const failedServices = allErrors(error).map(e => ({
      id: getSvcID(e.path),
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
          <Link
            href={`/alerts?allServices=1&filter=all&search=${encodeURIComponent(
              value.summary,
            )}`}
            target='_blank'
            rel='noopener noreferrer'
          >
            <Button
              variant='contained'
              color='primary'
              size='small'
              endIcon={<OpenInNewIcon />}
            >
              Monitor Alerts
            </Button>
          </Link>
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
      alert={hasSubmitted}
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
            onChange={newValue => setValue(newValue)}
            disabled={loading}
            errors={fieldErrors(error)}
          />
        )
      }
      PaperProps={{ className: classes.dialog }}
      onSubmit={() => (hasSubmitted ? props.onClose() : mutate())}
      onNext={currentStep < 2 ? () => setStep(currentStep + 1) : null}
      onBack={currentStep > 0 ? () => setStep(currentStep - 1) : null}
    />
  )
}
