import React from 'react'
import p from 'prop-types'
import Query from '../util/Query'
import gql from 'graphql-tag'
import FlatList from '../lists/FlatList'
import { ScheduleTZFilter } from './ScheduleTZFilter'
import { Grid, Card } from '@material-ui/core'
import FilterContainer from '../util/FilterContainer'
import PageActions from '../util/PageActions'
import { connect } from 'react-redux'
import { urlParamSelector } from '../selectors'
import { startCase, sortBy } from 'lodash-es'
import { RotationAvatar, UserAvatar } from '../util/avatar'
import OtherActions from '../util/OtherActions'
import SpeedDial from '../util/SpeedDial'
import { AccountPlus, AccountMultiplePlus } from 'mdi-material-ui'
import ScheduleRuleCreateDialog from './ScheduleRuleCreateDialog'
import { ruleSummary } from './util'
import ScheduleRuleEditDialog from './ScheduleRuleEditDialog'
import ScheduleRuleDeleteDialog from './ScheduleRuleDeleteDialog'
import { resetURLParams } from '../actions'

const query = gql`
  query scheduleRules($id: ID!) {
    schedule(id: $id) {
      id
      timeZone
      targets {
        target {
          id
          type
          name
        }
        rules {
          id
          start
          end
          weekdayFilter
        }
      }
    }
  }
`

@connect(
  state => ({ zone: urlParamSelector(state)('tz', 'local') }),
  dispatch => ({ resetFilter: () => dispatch(resetURLParams('tz')) }),
)
export default class ScheduleRuleList extends React.PureComponent {
  static propTypes = {
    scheduleID: p.string.isRequired,
  }

  state = {
    editTarget: null,
    deleteTarget: null,
    createType: null,
  }

  render() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.scheduleID }}
        render={({ data }) =>
          this.renderList(data.schedule.targets, data.schedule.timeZone)
        }
      />
    )
  }
  getHeaderNote() {
    const zone = this.props.zone
    return `Showing times in ${zone === 'local' ? 'local time' : zone}.`
  }
  renderList(targets, timeZone) {
    const items = []

    let lastType
    sortBy(targets, ['target.type', 'target.name']).forEach(tgt => {
      const { name, id, type } = tgt.target
      if (type !== lastType) {
        items.push({ subHeader: startCase(type + 's') })
        lastType = type
      }

      items.push({
        title: name,
        url: (type === 'rotation' ? '/rotations/' : '/users/') + id,
        subText: ruleSummary(tgt.rules, timeZone, this.props.zone),
        icon:
          type === 'rotation' ? <RotationAvatar /> : <UserAvatar userID={id} />,
        action: (
          <OtherActions
            actions={[
              {
                label: 'Edit',
                onClick: () => this.setState({ editTarget: { type, id } }),
              },
              {
                label: 'Delete',
                onClick: () => this.setState({ deleteTarget: { type, id } }),
              },
            ]}
          />
        ),
      })
    })

    return (
      <React.Fragment>
        <PageActions>
          <FilterContainer onReset={() => this.props.resetFilter()}>
            <Grid item xs={12}>
              <ScheduleTZFilter scheduleTimeZone={timeZone} />
            </Grid>
          </FilterContainer>
          <SpeedDial
            label='Add Assignment'
            actions={[
              {
                label: 'Add Rotation',
                onClick: () => this.setState({ createType: 'rotation' }),
                icon: <AccountMultiplePlus />,
              },
              {
                label: 'Add User',
                onClick: () => this.setState({ createType: 'user' }),
                icon: <AccountPlus />,
              },
            ]}
          />
        </PageActions>
        <Card style={{ width: '100%' }}>
          <FlatList headerNote={this.getHeaderNote()} items={items} />
        </Card>

        {this.state.createType && (
          <ScheduleRuleCreateDialog
            targetType={this.state.createType}
            scheduleID={this.props.scheduleID}
            onClose={() => this.setState({ createType: null })}
          />
        )}
        {this.state.editTarget && (
          <ScheduleRuleEditDialog
            target={this.state.editTarget}
            scheduleID={this.props.scheduleID}
            onClose={() => this.setState({ editTarget: null })}
          />
        )}
        {this.state.deleteTarget && (
          <ScheduleRuleDeleteDialog
            target={this.state.deleteTarget}
            scheduleID={this.props.scheduleID}
            onClose={() => this.setState({ deleteTarget: null })}
          />
        )}
      </React.Fragment>
    )
  }
}
