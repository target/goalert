import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import Checkbox from '@material-ui/core/Checkbox'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction'
import Typography from '@material-ui/core/Typography'
import withStyles from '@material-ui/core/styles/withStyles'
import { connect } from 'react-redux'
import { setCheckedAlerts } from '../../actions'
import { bindActionCreators } from 'redux'
import statusStyles from '../../util/statusStyles'
import { alertFilterSelector } from '../../selectors'
import { formatTimeSince } from '../../util/timeFormat'
import { AppLink } from '../../util/AppLink'

const styles = {
  checkBox: {
    marginRight: '27px',
    padding: '4px', // match form control padding
  },
  iconButton: {
    width: 'fit-content',
  },
  listItem: {
    width: '100%',
  },
  summaryText: {
    marginLeft: '0.5em',
  },
  ...statusStyles,
}

const mapStateToProps = state => ({
  allChecked: state.alerts.allChecked,
  checkedAlerts: state.alerts.checkedAlerts,
  filter: alertFilterSelector(state),
})

const mapDispatchToProps = dispatch =>
  bindActionCreators(
    {
      setCheckedAlerts,
    },
    dispatch,
  )

@withStyles(styles)
@connect(mapStateToProps, mapDispatchToProps)
export default class AlertsListDataWrapper extends Component {
  static propTypes = {
    alert: p.object.isRequired,
    onServicePage: p.bool,
  }

  componentWillReceiveProps(nextProps) {
    const { allChecked, checkedAlerts, setCheckedAlerts } = this.props
    const { alert, allChecked: nextChecked } = nextProps

    if (!allChecked && nextChecked) {
      setCheckedAlerts([...checkedAlerts, alert.number])
    }
  }

  toggleChecked = id => {
    const { checkedAlerts: _checkedAlerts, setCheckedAlerts } = this.props
    const checkedAlerts = _checkedAlerts.slice() // copy array

    if (checkedAlerts.includes(id)) {
      const idx = checkedAlerts.indexOf(id)
      checkedAlerts.splice(idx, 1) // removes at index
      setCheckedAlerts(checkedAlerts)
    } else {
      checkedAlerts.push(id)
      setCheckedAlerts(checkedAlerts)
    }
  }

  render() {
    const { alert, checkedAlerts, classes, onServicePage } = this.props

    const checkbox = (
      <Checkbox
        className={classes.checkBox}
        classes={{
          root: classes.iconButton,
        }}
        checked={
          checkedAlerts.includes(alert.number) && alert.status !== 'closed'
        }
        disabled={alert.status === 'closed'}
        data-cy={'alert-' + alert.number}
        disableRipple
        tabIndex={-1}
        onClick={e => {
          e.stopPropagation()
          e.preventDefault()
          this.toggleChecked(alert.number)
        }}
      />
    )

    let statusClass
    switch (alert.status.toLowerCase()) {
      case 'unacknowledged':
        statusClass = classes.statusError
        break
      case 'acknowledged':
        statusClass = classes.statusWarning
        break
      default:
        statusClass = classes.noStatus
        break
    }

    return (
      <ListItem
        button
        className={statusClass}
        component={AppLink}
        to={`/alerts/${alert.number}`}
      >
        {checkbox}

        <ListItemText disableTypography style={{ paddingRight: '2.75em' }}>
          <Typography>
            <b>{alert.number}: </b>
            {alert.status.toUpperCase()}
          </Typography>
          {onServicePage ? null : (
            <Typography variant='caption'>
              {alert.service.name + ':'}
            </Typography>
          )}
          <Typography variant='caption' noWrap className={classes.summaryText}>
            {alert.summary}
          </Typography>
        </ListItemText>
        <ListItemSecondaryAction>
          <ListItemText disableTypography>
            <Typography variant='caption'>
              {formatTimeSince(alert.created_at)}
            </Typography>
          </ListItemText>
        </ListItemSecondaryAction>
      </ListItem>
    )
  }
}
