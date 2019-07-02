import React from 'react'
import ReactDOM from 'react-dom'

const GracefulUnmountContext = React.createContext({
  onRender: () => {},
  onMount: () => {},
  onUnmount: () => {},
})
GracefulUnmountContext.displayName = 'GracefulUnmountContext'

let id = 1
export class GracefulUnmounter extends React.PureComponent {
  id = id++

  componentDidMount() {
    this.props.onMount(this.id, this.props.render)
  }

  componentWillUnmount() {
    this.props.onUnmount(this.id)
  }

  componentDidUpdate() {
    this.props.onUpdate(this.id, this.props.render)
  }

  render() {
    return null
  }
}

export class GracefulUnmounterProvider extends React.PureComponent {
  items = []

  onExited = id => {
    this.items = this.items.filter(i => i.id !== id)
    this.forceUpdate()
  }

  onUnmount = id => {
    this.items.find(i => i.id === id).isUnmounting = true
    this.forceUpdate()
  }

  onMount = (id, render) => {
    this.items.push({
      id,
      isUnmounting: false,
      render,
    })
    this.forceUpdate()
  }

  onUpdate = (id, render) => {
    const item = this.items.find(i => i.id === id)
    if (item.render === render) return
    item.render = render
    this.forceUpdate()
  }

  renderItems() {
    return this.items
      .filter(item => item)
      .map(item =>
        item.render({
          key: 'graceful_' + item.id,
          isUnmounting: item.isUnmounting,
          onExited: () => this.onExited(item.id),
        }),
      )
  }

  value = {
    onMount: this.onMount,
    onUnmount: this.onUnmount,
    onUpdate: this.onUpdate,
  }

  render() {
    return (
      <React.Fragment>
        {ReactDOM.createPortal(
          this.renderItems(),
          document.getElementById('graceful-unmount'),
          'container',
        )}
        <GracefulUnmountContext.Provider value={this.value}>
          {this.props.children}
        </GracefulUnmountContext.Provider>
      </React.Fragment>
    )
  }
}

export default function gracefulUnmount() {
  return Component =>
    function GracefulUnmount(props) {
      return (
        <GracefulUnmountContext.Consumer>
          {ctxProps => (
            <GracefulUnmounter
              {...ctxProps}
              component={Component}
              componentProps={props}
              render={gracefulProps => (
                <Component {...gracefulProps} {...props} />
              )}
            />
          )}
        </GracefulUnmountContext.Consumer>
      )
    }
}
