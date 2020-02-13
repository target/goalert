import React from 'react'
import p from 'prop-types'
import PageActions from '../util/PageActions'
import { Grid, FormControlLabel, Switch } from '@material-ui/core'
import QueryList from '../lists/QueryList'
import gql from 'graphql-tag'
import { UserAvatar } from '../util/avatar'
import OtherActions from '../util/OtherActions'
import FilterContainer from '../util/FilterContainer'
import { UserSelect } from '../selection'
import { connect } from 'react-redux'
import { setURLParam, resetURLParams } from '../actions'
import { urlParamSelector } from '../selectors'
import { ScheduleTZFilter } from './ScheduleTZFilter'
import ScheduleOverrideCreateDialog from './ScheduleOverrideCreateDialog'
import ScheduleNewOverrideFAB from './ScheduleNewOverrideFAB'
import ScheduleOverrideDeleteDialog from './ScheduleOverrideDeleteDialog'
import { formatOverrideTime } from './util'
import ScheduleOverrideEditDialog from './ScheduleOverrideEditDialog'

// the query name `scheduleOverrides` is used for refetch queries
const query = gql`
  query scheduleOverrides($input: UserOverrideSearchOptions) {
    userOverrides(input: $input) {
      nodes {
        id
        start
        end
        addUser {
          id
          name
        }
        removeUser {
          id
          name
        }
      }

      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`

const mapStateToProps = state => {
  return {
    userFilter: urlParamSelector(state)('userFilter', []),
    showPast: urlParamSelector(state)('showPast', false),
    zone: urlParamSelector(state)('tz', 'local'),
  }
}
const mapDispatchToProps = dispatch => {
  return {
    setZone: value => dispatch(setURLParam('tz', value, 'local')),
    setUserFilter: value => dispatch(setURLParam('userFilter', value)),
    setShowPast: value => dispatch(setURLParam('showPast', value)),
    resetFilter: () => dispatch(resetURLParams('userFilter', 'showPast', 'tz')),
  }
}

@connect(mapStateToProps, mapDispatchToProps)
export default class ScheduleOverrideList extends React.PureComponent {
  static propTypes = {
    scheduleID: p.string.isRequired,
  }

  state = {
    editID: null,
    deleteID: null,
    create: null,
  }

  render() {
    const { zone } = this.props

    const subText = n => {
      const timeStr = formatOverrideTime(n.start, n.end, zone)
      if (n.addUser && n.removeUser) {
        // replace
        return `Replaces ${n.removeUser.name} from ${timeStr}`
      }
      if (n.addUser) {
        // add
        return `Added from ${timeStr}`
      }
      // remove
      return `Removed from ${timeStr}`
    }

    const zoneText = zone === 'local' ? 'local time' : zone
    const hasUsers = Boolean(this.props.userFilter.length)
    const note = this.props.showPast
      ? `Showing all overrides${
          hasUsers ? ' for selected users' : ''
        } in ${zoneText}.`
      : `Showing active and future overrides${
          hasUsers ? ' for selected users' : ''
        } in ${zoneText}.`

    return (
      <React.Fragment>
        <PageActions>
          <FilterContainer onReset={() => this.props.resetFilter()}>
            <Grid item xs={12}>
              <FormControlLabel
                control={
                  <Switch
                    checked={this.props.showPast}
                    onChange={e => this.props.setShowPast(e.target.checked)}
                    value='showPast'
                  />
                }
                label='Show past overrides'
              />
            </Grid>
            <Grid item xs={12}>
              <ScheduleTZFilter scheduleID={this.props.scheduleID} />
            </Grid>
            <Grid item xs={12}>
              <UserSelect
                label='Filter users...'
                multiple
                value={this.props.userFilter}
                onChange={value => this.props.setUserFilter(value)}
              />
            </Grid>
          </FilterContainer>
          <ScheduleNewOverrideFAB
            onClick={variant => this.setState({ create: variant })}
          />
        </PageActions>
        <QueryList
          listHeader={note}
          noSearch
          noPlaceholder
          query={query}
          mapDataNode={n => ({
            title: n.addUser ? n.addUser.name : n.removeUser.name,
            subText: subText(n),
            icon: (
              <UserAvatar userID={n.addUser ? n.addUser.id : n.removeUser.id} />
            ),
            action: (
              <OtherActions
                actions={[
                  {
                    label: 'Edit',
                    onClick: () => this.setState({ editID: n.id }),
                  },
                  {
                    label: 'Delete',
                    onClick: () => this.setState({ deleteID: n.id }),
                  },
                ]}
              />
            ),
          })}
          variables={{
            input: {
              scheduleID: this.props.scheduleID,
              start: this.props.showPast ? null : new Date().toISOString(),
              filterAnyUserID: this.props.userFilter,
            },
          }}
        />
        {this.state.create && (
          <ScheduleOverrideCreateDialog
            scheduleID={this.props.scheduleID}
            variant={this.state.create}
            onClose={() => this.setState({ create: null })}
          />
        )}
        {this.state.deleteID && (
          <ScheduleOverrideDeleteDialog
            overrideID={this.state.deleteID}
            onClose={() => this.setState({ deleteID: null })}
          />
        )}
        {this.state.editID && (
          <ScheduleOverrideEditDialog
            overrideID={this.state.editID}
            onClose={() => this.setState({ editID: null })}
          />
        )}
      </React.Fragment>
    )
  }
}
