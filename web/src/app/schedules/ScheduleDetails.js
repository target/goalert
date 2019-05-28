import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import Grid from '@material-ui/core/Grid'
import Switch from '@material-ui/core/Switch'
import DetailsPage from '../details/DetailsPage'
import Query from '../util/Query'
import { UserSelect } from '../selection'
import FilterContainer from '../util/FilterContainer'
import PageActions from '../util/PageActions'
import OtherActions from '../util/OtherActions'
import ScheduleEditDialog from './ScheduleEditDialog'
import ScheduleDeleteDialog from './ScheduleDeleteDialog'
import ScheduleCalendarQuery from './ScheduleCalendarQuery'
import { urlParamSelector } from '../selectors'
import { resetURLParams, setURLParam } from '../actions'
import { connect } from 'react-redux'

const query = gql`
  query($id: ID!) {
    schedule(id: $id) {
      id
      name
      description
      timeZone
    }
  }
`
const partialQuery = gql`
  query($id: ID!) {
    schedule(id: $id) {
      id
      name
      description
    }
  }
`

const mapStateToProps = state => ({
  userFilter: urlParamSelector(state)('userFilter', []),
  activeOnly: urlParamSelector(state)('activeOnly', false),
})

const mapDispatchToProps = dispatch => {
  return {
    setUserFilter: value => dispatch(setURLParam('userFilter', value)),
    setActiveOnly: value => dispatch(setURLParam('activeOnly', value)),
    resetFilter: () =>
      dispatch(
        resetURLParams('userFilter', 'start', 'activeOnly', 'tz', 'duration'),
      ),
  }
}

@connect(
  mapStateToProps,
  mapDispatchToProps,
)
export default class ScheduleDetails extends React.PureComponent {
  static propTypes = {
    scheduleID: p.string.isRequired,
  }

  state = {
    edit: false,
    delete: false,
  }

  render() {
    return (
      <Query
        query={query}
        partialQuery={partialQuery}
        variables={{ id: this.props.scheduleID }}
        render={({ data }) => this.renderPage(data.schedule)}
      />
    )
  }

  renderPage = data => {
    return (
      <React.Fragment>
        <PageActions>
          <FilterContainer onReset={() => this.props.resetFilter()}>
            <Grid item xs={12}>
              <FormControlLabel
                control={
                  <Switch
                    checked={this.props.activeOnly}
                    onChange={e => this.props.setActiveOnly(e.target.checked)}
                    value='activeOnly'
                  />
                }
                label='Active shifts only'
              />
            </Grid>
            <Grid item xs={12}>
              <UserSelect
                label='Filter users...'
                multiple
                value={this.props.userFilter}
                onChange={this.props.setUserFilter}
              />
            </Grid>
          </FilterContainer>
          <OtherActions
            actions={[
              {
                label: 'Edit Schedule',
                onClick: () => this.setState({ edit: true }),
              },
              {
                label: 'Delete Schedule',
                onClick: () => this.setState({ delete: true }),
              },
            ]}
          />
        </PageActions>
        <DetailsPage
          title={data.name}
          details={data.description}
          titleFooter={
            <React.Fragment>Time Zone: {data.timeZone}</React.Fragment>
          }
          links={[
            { label: 'Assignments', url: 'assignments' },
            { label: 'Escalation Policies', url: 'escalation-policies' },
            { label: 'Overrides', url: 'overrides' },
            { label: 'Shifts', url: 'shifts' },
          ]}
          pageFooter={
            <ScheduleCalendarQuery scheduleID={this.props.scheduleID} />
          }
        />
        {this.state.edit && (
          <ScheduleEditDialog
            scheduleID={this.props.scheduleID}
            onClose={() => this.setState({ edit: false })}
          />
        )}
        {this.state.delete && (
          <ScheduleDeleteDialog
            scheduleID={this.props.scheduleID}
            onClose={() => this.setState({ delete: false })}
          />
        )}
      </React.Fragment>
    )
  }
}
