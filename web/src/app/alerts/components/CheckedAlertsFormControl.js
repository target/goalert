/* eslint @typescript-eslint/camelcase: 0 */
import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import { setCheckedAlerts, setAlertsActionComplete } from '../../actions'
import { bindActionCreators } from 'redux'
import { connect } from 'react-redux'
import Checkbox from '@material-ui/core/Checkbox'
import Grid from '@material-ui/core/Grid'
import IconButton from '@material-ui/core/IconButton'
import Tooltip from '@material-ui/core/Tooltip'
import ArrowDropDown from '@material-ui/icons/ArrowDropDown'
import AcknowledgeIcon from '@material-ui/icons/Check'
import CloseIcon from '@material-ui/icons/Close'
import EscalateIcon from '@material-ui/icons/ArrowUpward'
import withStyles from '@material-ui/core/styles/withStyles'
import { styles as globalStyles } from '../../styles/materialStyles'
import Icon from '@material-ui/core/Icon'
import OtherActions from '../../util/OtherActions'
import classnames from 'classnames'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import UpdateAlertsSnackbar from './UpdateAlertsSnackbar'
import withWidth from '@material-ui/core/withWidth'
import { alertFilterSelector } from '../../selectors'

const updateAlerts = gql`
  mutation UpdateAlertsMutation($input: UpdateAlertsInput!) {
    updateAlerts(input: $input) {
      alertID
      id
    }
  }
`

const escalateAlerts = gql`
  mutation EscalateAlertsMutation($input: [Int!]) {
    escalateAlerts(input: $input) {
      alertID
      id
    }
  }
`

const styles = theme => ({
  ...globalStyles(theme),
  checkboxGridContainer: {
    padding: '14px',
    paddingBottom: 0,
  },
  hover: {
    '&:hover': {
      cursor: 'pointer',
    },
  },
  icon: {
    alignItems: 'center',
    display: 'flex',
  },
  popper: {
    opacity: 1,
  },
  hidden: {
    visibility: 'hidden',
  },
})

const mapStateToProps = state => ({
  actionComplete: state.alerts.actionComplete,
  checkedAlerts: state.alerts.checkedAlerts,
  filter: alertFilterSelector(state),
})

const mapDispatchToProps = dispatch =>
  bindActionCreators(
    {
      setCheckedAlerts,
      setAlertsActionComplete,
    },
    dispatch,
  )

@withStyles(styles)
@withWidth()
@connect(mapStateToProps, mapDispatchToProps)
export default class CheckedAlertsFormControl extends Component {
  static propTypes = {
    cardClassName: p.string,
    data: p.shape({
      alert2: p.shape({
        items: p.array,
        total_count: p.number,
      }),
    }),
  }

  state = {
    errorMessage: '',
    updateMessage: '',
  }

  visibleAlertIDs = () => {
    if (!this.props.data.alerts2) return []
    return this.props.data.alerts2.items
      .filter(a => a.status !== 'closed')
      .map(a => a.number)
  }

  checkedAlertIDs = () => {
    const alerts = {}
    this.visibleAlertIDs().forEach(id => {
      alerts[id] = true
    })

    return this.props.checkedAlerts.filter(id => alerts[id])
  }

  areAllChecked = () => {
    return this.visibleAlertIDs().length === this.checkedAlertIDs().length
  }

  areNoneChecked = () => {
    return this.checkedAlertIDs().length === 0
  }

  // handle resetting selected alerts when visiting another route
  componentWillUnmount() {
    this.props.setCheckedAlerts([])
  }

  setNone = () => {
    this.props.setCheckedAlerts([])
  }

  setAll = () => {
    this.props.setCheckedAlerts(this.visibleAlertIDs())
  }

  handleSelectAllChange = () => {
    if (this.areNoneChecked()) return this.setAll()

    return this.setNone()
  }

  updateAlerts = (newStatus, mutation) => {
    return mutation({
      variables: {
        input: {
          alertIDs: this.checkedAlertIDs(),
          newStatus,
        },
      },
    })
  }

  getSelectOptions = () => {
    return [
      {
        label: 'All',
        onClick: this.setAll,
      },
      {
        label: 'None',
        onClick: this.setNone,
      },
    ]
  }

  onUpdate = numUpdated => {
    this.props.setAlertsActionComplete(true)
    this.setState({
      updateMessage: `${numUpdated} of ${
        this.checkedAlertIDs().length
      } alerts updated`,
    })
    this.props.setCheckedAlerts([])
    this.props.refetch()
  }

  getAckButton = () => {
    return (
      <Mutation
        mutation={updateAlerts}
        onError={err => {
          this.props.setAlertsActionComplete(true)
          this.setState({ errorMessage: err.message })
        }}
        update={(cache, { data }) =>
          this.onUpdate(data.updateAlerts ? data.updateAlerts.length : 0)
        }
      >
        {mutation => (
          <Tooltip
            title='Acknowledge'
            placement='bottom'
            classes={{ popper: this.props.classes.popper }}
          >
            <IconButton
              aria-label='Acknowledge Selected Alerts'
              data-cy='acknowledge'
              onClick={() => this.updateAlerts('StatusAcknowledged', mutation)}
            >
              <AcknowledgeIcon />
            </IconButton>
          </Tooltip>
        )}
      </Mutation>
    )
  }

  getCloseButton = () => {
    return (
      <Mutation
        mutation={updateAlerts}
        onError={err => {
          this.props.setAlertsActionComplete(true)
          this.setState({ errorMessage: err.message })
        }}
        update={(cache, { data }) => this.onUpdate(data.updateAlerts.length)}
      >
        {mutation => (
          <Tooltip
            title='Close'
            placement='bottom'
            classes={{ popper: this.props.classes.popper }}
          >
            <IconButton
              aria-label='Close Selected Alerts'
              data-cy='close'
              onClick={() => this.updateAlerts('StatusClosed', mutation)}
            >
              <CloseIcon />
            </IconButton>
          </Tooltip>
        )}
      </Mutation>
    )
  }

  getEscalateButton = () => {
    return (
      <Mutation
        mutation={escalateAlerts}
        onError={err => {
          this.props.setAlertsActionComplete(true)
          this.setState({ errorMessage: err.message })
        }}
        update={(cache, { data }) => this.onUpdate(data.escalateAlerts.length)}
      >
        {mutation => (
          <Tooltip
            title='Escalate'
            placement='bottom'
            classes={{ popper: this.props.classes.popper }}
          >
            <IconButton
              aria-label='Escalate Selected Alerts'
              data-cy='escalate'
              onClick={() =>
                mutation({
                  variables: {
                    input: this.checkedAlertIDs(),
                  },
                })
              }
            >
              <EscalateIcon />
            </IconButton>
          </Tooltip>
        )}
      </Mutation>
    )
  }

  renderActionButtons = () => {
    const { checkedAlerts, classes, filter } = this.props
    if (!checkedAlerts.length) return null

    let ack = null
    let close = null
    let escalate = null
    if (
      filter === 'active' ||
      filter === 'unacknowledged' ||
      filter === 'all'
    ) {
      ack = this.getAckButton()
    }

    if (filter !== 'closed') {
      close = this.getCloseButton()
      escalate = this.getEscalateButton()
    }

    return (
      <Grid item className={classes.icon}>
        {ack}
        {close}
        {escalate}
      </Grid>
    )
  }

  render() {
    const { actionComplete, classes } = this.props
    const { errorMessage, updateMessage } = this.state

    return (
      <React.Fragment>
        <UpdateAlertsSnackbar
          errorMessage={errorMessage}
          numberChecked={this.checkedAlertIDs().length}
          onClose={() => this.props.setAlertsActionComplete(false)}
          onExited={() => {
            this.setState({ errorMessage: '', updateMessage: '' })
          }}
          open={actionComplete}
          updateMessage={updateMessage}
        />
        <Grid container className={classes.checkboxGridContainer}>
          <Grid item>
            <Checkbox
              checked={!this.areNoneChecked()}
              data-cy='select-all'
              indeterminate={!this.areNoneChecked() && !this.areAllChecked()}
              tabIndex={-1}
              onChange={this.handleSelectAllChange}
            />
          </Grid>
          <Grid
            item
            className={classnames(classes.hover, classes.icon)}
            data-cy='checkboxes-menu'
          >
            <OtherActions
              icon={
                <Icon>
                  <ArrowDropDown />
                </Icon>
              }
              actions={this.getSelectOptions()}
              placement='right'
            />
          </Grid>
          {this.renderActionButtons()}
        </Grid>
      </React.Fragment>
    )
  }
}
