import React, { Component } from 'react'
import Typography from '@material-ui/core/Typography'
import {
  SentimentDissatisfied,
  SentimentVeryDissatisfied,
} from '@material-ui/icons'

export class PageNotFound extends Component {
  render() {
    return (
      <div style={{ textAlign: 'center' }}>
        <SentimentDissatisfied style={{ height: '33vh', width: '33vw' }} />
        <Typography variant='h5'>
          Sorry, the page you were trying to reach could not be found.
        </Typography>
      </div>
    )
  }
}

export class ObjectNotFound extends Component {
  render() {
    return (
      <div style={{ textAlign: 'center' }}>
        <SentimentDissatisfied style={{ height: '33vh', width: '33vw' }} />
        <Typography variant='h5'>
          Sorry, the {this.props.type || 'thing'} you were looking for could not
          be found.
        </Typography>
        <Typography variant='caption'>
          Someone may have deleted it, or it never existed.
        </Typography>
      </div>
    )
  }
}

export class GenericError extends Component {
  render() {
    let errorText
    if (this.props.error) {
      errorText = <Typography variant='caption'>{this.props.error}</Typography>
    }
    return (
      <div style={{ textAlign: 'center' }}>
        <SentimentVeryDissatisfied style={{ height: '33vh', width: '33vw' }} />
        <Typography variant='h5'>Sorry, an error occurred.</Typography>
        {errorText}
      </div>
    )
  }
}
