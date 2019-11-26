import React from 'react'
import QueryList from './QueryList'

import PageActions from '../util/PageActions'
import p from 'prop-types'

import Search from '../util/Search'
import CreateFAB from './CreateFAB'

export default class SimpleListPage extends React.PureComponent {
  state = {
    create: false,
  }

  static propTypes = {
    createForm: p.element,
    createLabel: p.string,
    queryProps: p.object,
  }

  render() {
    const { createForm, createLabel, ...queryProps } = this.props
    return (
      <React.Fragment>
        <PageActions>
          <Search />
        </PageActions>

        <QueryList {...queryProps} />

        {createForm && (
          <CreateFAB
            onClick={() => this.setState({ create: true })}
            title={`Create ${createLabel}`}
          />
        )}

        {this.state.create &&
          React.cloneElement(createForm, {
            onClose: () => this.setState({ create: false }),
          })}
      </React.Fragment>
    )
  }
}
