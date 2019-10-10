import React from 'react'
import p from 'prop-types'
import QueryList from './QueryList'
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
    const { createForm, createLabel, ...rest } = this.props
    return (
      <React.Fragment>
        <QueryList {...rest} />

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
