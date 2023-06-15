import React, { useRef, useState } from 'react'
import { PropTypes as p } from 'prop-types'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import CardHeader from '@mui/material/CardHeader'
import Dialog from '@mui/material/Dialog'
import List from '@mui/material/List'
import Typography from '@mui/material/Typography'
import makeStyles from '@mui/styles/makeStyles'
import { Add } from '@mui/icons-material'
import { DragDropContext, Droppable, Draggable } from 'react-beautiful-dnd'
import { gql, useMutation } from '@apollo/client'
import CreateFAB from '../lists/CreateFAB'
import PolicyStepCreateDialog from './PolicyStepCreateDialog'
import { useResetURLParams, useURLParam } from '../actions'
import PolicyStep from './PolicyStep'
import DialogTitleWrapper from '../dialogs/components/DialogTitleWrapper'
import DialogContentError from '../dialogs/components/DialogContentError'
import { policyStepsQuery } from './PolicyStepsQuery'
import { useIsWidthDown } from '../util/useWidth'

const useStyles = makeStyles((theme) => ({
  dndDragging: {
    backgroundColor: theme.palette.background.default,
  },
  paddingTop: {
    paddingTop: '1em',
  },
}))

const mutation = gql`
  mutation UpdateEscalationPolicyMutation(
    $input: UpdateEscalationPolicyInput!
  ) {
    updateEscalationPolicy(input: $input)
  }
`

function PolicyStepsCard(props) {
  const classes = useStyles()
  const { escalationPolicyID, repeat, steps } = props

  const isMobile = useIsWidthDown('md')
  const stepNumParam = 'createStep'
  const [createStep, setCreateStep] = useURLParam(stepNumParam, false)
  const resetCreateStep = useResetURLParams(stepNumParam)

  const oldID = useRef(null)
  const oldIdx = useRef(null)
  const newIdx = useRef(null)

  const [error, setError] = useState(null)

  function arrayMove(arr) {
    const el = arr[oldIdx.current]
    arr.splice(oldIdx.current, 1)
    arr.splice(newIdx.current, 0, el)
  }

  function onStepUpdate(cache, data) {
    // mutation returns true on a success
    if (
      !data.updateEscalationPolicy ||
      oldIdx.current == null ||
      newIdx.current == null
    ) {
      return
    }

    // variables for query to read/write from the cache
    const variables = {
      id: escalationPolicyID,
    }

    // get the current state of the steps in the cache
    const { escalationPolicy } = cache.readQuery({
      query: policyStepsQuery,
      variables,
    })

    // get steps from cache
    const steps = escalationPolicy.steps.slice()

    // if optimistic cache update was successful, return out
    if (steps[newIdx.current].id === oldID.current) return

    // re-order escalationPolicy.steps array
    arrayMove(steps)

    // write new steps order to cache
    cache.writeQuery({
      query: policyStepsQuery,
      variables,
      data: {
        escalationPolicy: {
          ...escalationPolicy,
          steps,
        },
      },
    })
  }

  const [updateEscalationPolicy] = useMutation(mutation, {
    onCompleted: () => {
      oldID.current = null
      oldIdx.current = null
      newIdx.current = null
    },
    onError: (err) => setError(err),
    update: (cache, { data }) => onStepUpdate(cache, data),
    optimisticResponse: { updateEscalationPolicy: true },
  })

  function handleDragStart() {
    // adds a little vibration if the browser supports it
    if (window.navigator.vibrate) {
      window.navigator.vibrate(100)
    }
  }

  // update step order on ui and send out mutation
  function onDragEnd(result) {
    // dropped outside the list
    if (!result.destination) {
      return
    }

    // map ids to swap elements
    const sids = steps.map((s) => s.id)
    oldID.current = result.draggableId
    oldIdx.current = sids.indexOf(oldID.current)
    newIdx.current = result.destination.index

    // re-order sids array
    arrayMove(sids)

    return updateEscalationPolicy({
      variables: {
        input: {
          id: escalationPolicyID,
          stepIDs: sids,
        },
      },
    })
  }

  function renderRepeatText() {
    if (!steps.length) {
      return null
    }

    let text = ''
    if (repeat === 0) text = 'Do not repeat'
    else if (repeat === 1) text = 'Repeat once'
    else text = `Repeat ${repeat} times`

    return (
      <Typography variant='subtitle1' component='p'>
        {text}
      </Typography>
    )
  }

  function renderNoSteps() {
    return (
      <Typography className={classes.paddingTop} variant='caption'>
        No steps currently on this Escalation Policy
      </Typography>
    )
  }

  /*
   * Renders the steps list with the drag and drop context
   *
   * Each step will have a grid containing the step number,
   * targets (rendered as mui chips), and the delay length
   * until the next escalation.
   */
  function renderStepsList() {
    if (!steps.length) {
      return renderNoSteps()
    }

    return (
      <React.Fragment>
        <Typography component='p' variant='subtitle1'>
          Notify the following:
        </Typography>
        <DragDropContext
          key='drag-context'
          onDragStart={handleDragStart}
          onDragEnd={(res) => onDragEnd(res)}
        >
          <Droppable droppableId='droppable'>
            {(provided) => (
              <div ref={provided.innerRef} {...provided.droppableProps}>
                <List data-cy='steps-list'>
                  {steps.map((step, index) => (
                    <Draggable
                      key={step.id}
                      draggableId={step.id}
                      index={index}
                    >
                      {(provided, snapshot) => {
                        // light grey background while dragging
                        const draggingBackground = snapshot.isDragging
                          ? classes.dndDragging
                          : null

                        return (
                          <div
                            ref={provided.innerRef}
                            {...provided.draggableProps}
                            {...provided.dragHandleProps}
                            className={draggingBackground}
                          >
                            <PolicyStep
                              escalationPolicyID={escalationPolicyID}
                              index={index}
                              repeat={repeat}
                              step={step}
                              steps={steps}
                              selected={snapshot.isDragging}
                            />
                          </div>
                        )
                      }}
                    </Draggable>
                  ))}
                  {provided.placeholder}
                </List>
              </div>
            )}
          </Droppable>
        </DragDropContext>
      </React.Fragment>
    )
  }

  const { message: errMsg } = error || {}

  return (
    <React.Fragment>
      {isMobile && (
        <CreateFAB onClick={() => setCreateStep(true)} title='Create Step' />
      )}
      {createStep && (
        <PolicyStepCreateDialog
          escalationPolicyID={escalationPolicyID}
          onClose={resetCreateStep}
        />
      )}
      <Card>
        <CardContent>
          <CardHeader
            title='Escalation Steps'
            component='h3'
            sx={{ padding: 0, margin: 0 }}
            action={
              !isMobile && (
                <Button
                  variant='contained'
                  onClick={() => setCreateStep(true)}
                  startIcon={<Add />}
                >
                  Create Step
                </Button>
              )
            }
          />
          {renderStepsList()}
          {renderRepeatText()}
        </CardContent>
      </Card>
      <Dialog open={Boolean(error)} onClose={() => setError(null)}>
        <DialogTitleWrapper
          fullScreen={useIsWidthDown('md')}
          title='An error occurred'
        />
        <DialogContentError error={errMsg} />
      </Dialog>
    </React.Fragment>
  )
}

PolicyStepsCard.propTypes = {
  escalationPolicyID: p.string.isRequired,
  repeat: p.number.isRequired, // # of times EP repeats escalation process
  steps: p.arrayOf(
    p.shape({
      id: p.string.isRequired,
      delayMinutes: p.number.isRequired,
      targets: p.arrayOf(
        p.shape({
          id: p.string.isRequired,
          name: p.string.isRequired,
          type: p.string.isRequired,
        }),
      ).isRequired,
    }),
  ).isRequired,
}

export default PolicyStepsCard
