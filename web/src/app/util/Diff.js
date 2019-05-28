import React from 'react'
import p from 'prop-types'
import { diffChars, diffWords } from 'diff'
import Typography from '@material-ui/core/Typography'
import withStyles from '@material-ui/core/styles/withStyles'

const styles = {
  // bg color for the whole first line
  oldLine: {
    backgroundColor: '#FFEEF0',
  },
  // removed chars darker red
  removed: {
    backgroundColor: '#FCB8C0',
  },
  // bg color for the whole second line
  newLine: {
    backgroundColor: '#E6FFED',
  },
  // added chars darker green
  added: {
    backgroundColor: '#AEF1BE',
  },
  // spacing on +/- symbols in diff
  symbol: {
    minWidth: '2em',
    display: 'inline-block',
  },
}

/*
 * Displays a difference in characters or words from
 * an old and new value
 */
@withStyles(styles)
export default class Diff extends React.PureComponent {
  static propTypes = {
    oldValue: p.string,
    newValue: p.string,
    type: p.oneOf(['chars', 'words']),
  }

  static defaultProps = {
    oldValue: '',
    newValue: '',
    type: 'chars',
  }

  render() {
    const { classes, oldValue, newValue, type } = this.props
    const { oldLine, removed, newLine, added } = classes

    let diff = []
    if (type === 'chars') diff = diffChars(oldValue, newValue)
    if (type === 'words') diff = diffWords(oldValue, newValue)

    const oldWithRemoved = diff.map((part, idx) => {
      if (part.added) return

      return (
        <span key={idx} className={part.removed ? removed : null}>
          {part.value}
        </span>
      )
    })

    const newWithAdded = diff.map((part, idx) => {
      if (part.removed) return

      return (
        <span key={idx} className={part.added ? added : null}>
          {part.value}
        </span>
      )
    })

    const hideRemoved = diff.length === 1 && diff[0].added // net new string
    const hideAdded = diff.length === 1 && diff[0].removed // deleted whole string

    return (
      <React.Fragment>
        {!hideRemoved && (
          <Typography className={oldLine} data-cy='old'>
            <span className={classes.symbol}>&nbsp;-</span>
            {oldWithRemoved}
          </Typography>
        )}
        {!hideAdded && (
          <Typography className={newLine} data-cy='new'>
            <span className={classes.symbol}>&nbsp;+</span>
            {newWithAdded}
          </Typography>
        )}
      </React.Fragment>
    )
  }
}
