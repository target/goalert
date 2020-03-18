import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import Dialog from '@material-ui/core/Dialog'
import List from '@material-ui/core/List'
import Typography from '@material-ui/core/Typography'
import { withStyles } from '@material-ui/core/styles/index'
import withWidth, { isWidthDown } from '@material-ui/core/withWidth'
import { DragDropContext, Droppable, Draggable } from 'react-beautiful-dnd'
import { styles as globalStyles } from '../styles/materialStyles'
import gql from 'graphql-tag'
import PolicyStep from './PolicyStep'
import { Mutation } from 'react-apollo'
import DialogTitleWrapper from '../dialogs/components/DialogTitleWrapper'
import DialogContentError from '../dialogs/components/DialogContentError'
import { policyStepsQuery } from './PolicyStepsQuery'

const styles = theme => {
  const { dndDragging } = globalStyles(theme)

  return {
    dndDragging,
    paddingTop: {
      paddingTop: '1em',
    },
  }
}

@withWidth()
@withStyles(styles)
export default class PolicyStepsCard extends Component {
  static propTypes = {
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

  state = {
    error: null,
  }

  oldID = null
  oldIdx = null
  newIdx = null

  arrayMove = arr => {
    const el = arr[this.oldIdx]
    arr.splice(this.oldIdx, 1)
    arr.splice(this.newIdx, 0, el)
  }

  onMutationUpdate = (cache, data) => {
    // mutation returns true on a success
    if (
      !data.updateEscalationPolicy ||
      this.oldIdx == null ||
      this.newIdx == null
    ) {
      return
    }

    // variables for query to read/write from the cache
    const variables = {
      id: this.props.escalationPolicyID,
    }

    // get the current state of the steps in the cache
    const { escalationPolicy } = cache.readQuery({
      query: policyStepsQuery,
      variables,
    })

    // get steps from cache
    const steps = escalationPolicy.steps.slice()

    // if optimistic cache update was successful, return out
    if (steps[this.newIdx].id === this.oldID) return

    // re-order escalationPolicy.steps array
    this.arrayMove(steps)

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

  handleDragStart = () => {
    // adds a little vibration if the browser supports it
    if (window.navigator.vibrate) {
      window.navigator.vibrate(100)
    }
  }

  // update step order on ui and send out mutation
  onDragEnd = (result, mutation) => {
    // dropped outside the list
    if (!result.destination) {
      return
    }

    // map ids to swap elements
    const sids = this.props.steps.map(s => s.id)
    this.oldID = result.draggableId
    this.oldIdx = sids.indexOf(this.oldID)
    this.newIdx = result.destination.index

    // re-order sids array
    this.arrayMove(sids)

    mutation({
      variables: {
        input: {
          id: this.props.escalationPolicyID,
          stepIDs: sids,
        },
      },
    })
  }

  renderRepeatText = () => {
    const { repeat, steps } = this.props

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

  renderNoSteps = () => {
    return (
      <Typography className={this.props.classes.paddingTop} variant='caption'>
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
  renderStepsList = () => {
    const { classes, escalationPolicyID, repeat, steps } = this.props

    if (!steps.length) {
      return this.renderNoSteps()
    }

    return (
      <React.Fragment>
        <Typography component='p' variant='subtitle1'>
          Notify the following:
        </Typography>
        <Mutation
          mutation={gql`
            mutation UpdateEscalationPolicyMutation(
              $input: UpdateEscalationPolicyInput!
            ) {
              updateEscalationPolicy(input: $input)
            }
          `}
          onCompleted={() => {
            this.oldID = null
            this.oldIdx = null
            this.newIdx = null
          }}
          optimisticResponse={{
            updateEscalationPolicy: true,
          }}
          onError={error => this.setState({ error })}
          update={(cache, { data }) => this.onMutationUpdate(cache, data)}
        >
          {mutation => (
            <DragDropContext
              key='drag-context'
              onDragStart={this.handleDragStart}
              onDragEnd={res => this.onDragEnd(res, mutation)}
            >
              <Droppable droppableId='droppable'>
                {provided => (
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
          )}
        </Mutation>
      </React.Fragment>
    )
  }

  render() {
    const { message: error } = this.state.error || {}

    return (
      <React.Fragment>
        <Card>
          <CardContent>
            <Typography variant='h5' component='h3'>
              Escalation Steps
            </Typography>
            {this.renderStepsList()}
            {this.renderRepeatText()}
          </CardContent>
        </Card>
        <Dialog
          open={Boolean(this.state.error)}
          onClose={() => this.setState({ error: null })}
        >
          <DialogTitleWrapper
            fullScreen={isWidthDown('md', this.props.width)}
            title='An error occurred'
          />
          <DialogContentError error={error} />
        </Dialog>
      </React.Fragment>
    )
  }
}
