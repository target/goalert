import React from 'react'
import QueryList from './QueryList'

import PageActions from '../util/PageActions'

import Search from '../util/Search'
import CreateFAB from './CreateFAB'

export default class SimpleListPage extends React.PureComponent {
  state = {
    create: false,
  }

  render() {
    const { createForm, ...queryProps } = this.props
    return (
      <React.Fragment>
        <PageActions>
          <Search />
        </PageActions>

        <QueryList {...queryProps} />

        {createForm && (
          <CreateFAB onClick={() => this.setState({ create: true })} />
        )}

        {this.state.create &&
          React.cloneElement(createForm, {
            onClose: () => this.setState({ create: false }),
          })}
      </React.Fragment>
    )
  }
}
