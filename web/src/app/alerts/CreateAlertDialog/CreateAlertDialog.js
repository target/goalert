import React, { useState, useEffect } from 'react'
import { Stepper, Step, StepLabel } from '@material-ui/core'

import { useCreateAlerts } from './useCreateAlerts'
import { fieldErrors, allErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import { CreateAlertForm } from './StepContent/CreateAlertForm'
import { CreateAlertReview } from './StepContent/CreateAlertReview'

import _ from 'lodash-es'

const stepTitles = ['Alert Info', 'Service Selection', 'Confirm']

const pluralize = num => (num !== 1 ? 's' : '')

export default function CreateAlertDialog(props) {
  const [step, setStep] = useState(0)
  const [value, setValue] = useState({
    summary: '',
    details: '',
    serviceIDs: [],
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
      ? ` (${failedServices.length} failed)`
      : ''
    reviewTitle = `Successfully created ${
      createdAlertIDs.length
    } alert${pluralize(createdAlertIDs.length)}${failMessage}`

    review = (
      <CreateAlertReview
        createdAlertIDs={_.chain(data)
          .values()
          .filter()
          .map(a => a.id)
          .value()}
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
      onSubmit={() => (hasSubmitted ? props.onClose() : mutate())}
      onNext={currentStep < 2 ? () => setStep(currentStep + 1) : null}
      onBack={currentStep > 0 ? () => setStep(currentStep - 1) : null}
    />
  )
}

// export default function CreateAlertDialog(props) {
//   const width = useWidth()
//   const classes = useStyles()

//   const [activeStep, setActiveStep] = useState(0)
//   const [formFields, setFormFields] = useState({
//     // data for mutation
//     summary: '',
//     details: '',
//     selectedServices: [],

//     // form helper
//     searchQuery: '',
//   })
//   const hasSubmitted = false // based on mutation state
//   const isFinishedSubmitting = false

//   const [
//     createAlerts,
//     { data: alertsCreated, error: alertsFailed, loading: isCreatingAlerts },
//   ] = useCreateAlerts(formFields)

//   const onStepContentChange = e => {
//     setFormFields(prevState => ({ ...prevState, ...e }))
//   }

//   const onLastStep = activeStep === stepTitles.length - 1

//   const onClose = () => {
//     props.handleRequestClose()
//   }

//   const resetForm = () => {
//     setActiveStep(0)
//     setFormFields({
//       summary: '',
//       details: '',
//       selectedServices: [],
//       searchQuery: '',
//     })
//   }
//   const handleNext = () => {
//     if (hasSubmitted) {
//       onClose()
//       return
//     }
//     if (activeStep === 2) {
//       // review step
//       // doSubmitCommit-w/e()
//       return
//     }

//     setActiveStep(activeStep + 1)
//   }

//   const isWideScreen = isWidthUp('md', width)
//   let nextLabel = 'Next'
//   if (hasSubmitted) nextLabel = 'Done'
//   else if (activeStep === 2) nextLabel = 'Submit'

//   return (
//     <Dialog
//       open={props.open}
//       onClose={onLastStep ? null : onClose} // NOTE only close on last step if user hits Done
//       fullScreen={!isWideScreen}
//       fullWidth
//       width='md'
//       PaperProps={{ className: classes.dialog }}
//       onExited={resetForm}
//     >
//       <DialogTitleWrapper fullScreen={!isWideScreen} title='Create New Alert' />
//       {!hasSubmitted && (
//         <Stepper activeStep={activeStep}>
//           {stepTitles.map(
//             label =>
//               label && (
//                 <Step key={label}>
//                   <StepLabel>{label}</StepLabel>
//                 </Step>
//               ),
//           )}
//         </Stepper>
//       )}
//       <DialogContent>
//         <FormContainer
//           onChange={e => onStepContentChange(e)}
//           value={formFields}
//           errors={fieldErrors(alertsFailed)}
//           optionalLabels
//         >
//           <Form id='create-alert-form'>
//             <StepContent
//               activeStep={activeStep}
//               setActiveStep={setActiveStep}
//               formFields={formFields}
//               mutationStatus={{ alertsCreated, alertsFailed, isCreatingAlerts }}
//               onChange={e => onStepContentChange(e)}
//             />
//           </Form>
//         </FormContainer>
//       </DialogContent>
//       <DialogNavigation
//         nextLabel={nextLabel}
//         backLabel={activeStep === 0 ? 'Cancel' : 'Back'}
//         onBack={hasSubmitted ? null : setActiveStep(activeStep - 1)}
//         onNext={handleNext}
//         form={activeStep === 2 ? 'create-alert-form' : null}
//       />
//     </Dialog>
//   )
// }
