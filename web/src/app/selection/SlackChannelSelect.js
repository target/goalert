import React from 'react'
import gql from 'graphql-tag'
import QuerySelect from './QuerySelect'
import withStyles from '@material-ui/core/styles/withStyles'
import except from 'except'

const query = gql`
  query($input: SlackChannelSearchOptions) {
    slackChannels(input: $input) {
      nodes {
        id
        name
      }
    }
  }
`

const valueQuery = gql`
  query($id: ID!) {
    slackChannel(id: $id) {
      id
      name
    }
  }
`

const styles = {
  slackButton: {
    textTransform: 'none',
    backgroundColor: 'white',
    border: '1px solid',
    borderColor: 'lightgrey',
    width: 'fit-content',
  },
  slackIcon: {
    marginRight: '0.5em',
  },
}

@withStyles(styles)
export class SlackChannelSelect extends React.PureComponent {
  render = () => (
    <QuerySelect
      {...except(this.props, 'classes')}
      query={query}
      valueQuery={valueQuery}
      defaultQueryVariables={{ input: { first: 5 } }}
    />
  )
}
