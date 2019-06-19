import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import Checkbox from '@material-ui/core/Checkbox'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction'
import Typography from '@material-ui/core/Typography'
import moment from 'moment'
import withStyles from '@material-ui/core/styles/withStyles'
import { connect } from 'react-redux'
import { withRouter } from 'react-router-dom'
import { setCheckedAlerts } from '../../actions'
import { bindActionCreators } from 'redux'
import statusStyles from '../../util/statusStyles'
import { alertFilterSelector } from '../../selectors'

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
@connect(
  mapStateToProps,
  mapDispatchToProps,
)
@withRouter
export default class AlertsListDataWrapper extends Component {
  static propTypes = {
    alert: p.object.isRequired,
    onServicePage: p.bool,
  }

  componentWillMount() {
    moment.updateLocale('en', {
      relativeTime: {
        future: 'in %s',
        past: '%s ago',
        s: '< 1m',
        m: '1m',
        mm: '%dm',
        h: '1h',
        hh: '%dh',
        d: '1d',
        dd: '%dd',
        M: '1mo',
        MM: '%dmo',
        y: '1y',
        yy: '%dy',
      },
    })
  }

  componentWillReceiveProps(nextProps, nextContext) {
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
    const { alert, checkedAlerts, classes, history, onServicePage } = this.props

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
        onChange={() => this.toggleChecked(alert.number)}
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
      <ListItem button className={statusClass}>
        {checkbox}
        <div
          className={classes.listItem}
          onClick={() => history.push(`/alerts/${alert.number}`)}
        >
          <ListItemText disableTypography style={{ paddingRight: '2.75em' }}>
            <Typography>
              <b>{alert.number}: </b>
              {alert.status.toUpperCase()}
            </Typography>
            {onServicePage ? null : (
              <Typography variant='caption'>{alert.service.name}</Typography>
            )}
            <Typography variant='caption' noWrap>
              {alert.summary}
            </Typography>
          </ListItemText>
          <ListItemSecondaryAction>
            <ListItemText disableTypography>
              <Typography variant='caption'>
                {moment(alert.created_at)
                  .local()
                  .fromNow()}
              </Typography>
            </ListItemText>
          </ListItemSecondaryAction>
        </div>
      </ListItem>
    )
  }
}
