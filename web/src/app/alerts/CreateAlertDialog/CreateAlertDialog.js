import React, { useState, useEffect } from 'react'
import p from 'prop-types'
import {
  makeStyles,
  Button,
  Grid,
  Stepper,
  Step,
  StepLabel,
  Typography,
} from '@material-ui/core'
import OpenInNewIcon from '@material-ui/icons/OpenInNew'
import _ from 'lodash-es'

import { useCreateAlerts } from './useCreateAlerts'
import { fieldErrors, allErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import { CreateAlertForm } from './StepContent/CreateAlertForm'
import { CreateAlertReview } from './StepContent/CreateAlertReview'
import { AppLink } from '../../util/AppLink'

const pluralize = (num) => (num !== 1 ? 's' : '')

const useStyles = makeStyles((theme) => ({
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

  const stepTitles = props.serviceID
    ? ['Alert Info', 'Confirm']
    : ['Alert Info', 'Service Selection', 'Confirm']

  const onNext = () => {
    if (currentStep === 0 && props.serviceID) {
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

    const failedServices = allErrors(error).map((e) => ({
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
          <Button
            variant='contained'
            color='primary'
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
            onChange={(newValue) => setValue(newValue)}
            disabled={loading}
            errors={fieldErrors(error)}
          />
        )
      }
      PaperProps={{ className: classes.dialog }}
      onSubmit={() => (hasCompleted ? props.onClose() : mutate())}
      onNext={currentStep < 2 ? onNext : null}
      onBack={currentStep > 0 ? () => setStep(currentStep - 1) : null}
    />
  )
}

CreateAlertDialog.propTypes = {
  onClose: p.func.isRequired,
  serviceID: p.string,
}
