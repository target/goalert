import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import IconButton from '@material-ui/core/IconButton'
import FavoriteFilledIcon from '@material-ui/icons/Star'
import FavoriteBorderIcon from '@material-ui/icons/StarBorder'
import { graphql2Client } from '../../apollo'
import Query from '../../util/Query'

const query = gql`
  query favQuery($id: ID!) {
    schedule(id: $id) {
      id
      isFavorite
    }
  }
`

const mutation = gql`
  mutation setFav($input: SetFavoriteInput!) {
    setFavorite(input: $input)
  }
`

export default class SetFavoriteButton extends React.PureComponent {
  static propTypes = {
    scheduleID: p.string.isRequired,
  }

  renderMutation(isFavorite) {
    return (
      <Mutation
        key='main'
        mutation={mutation}
        client={graphql2Client}
        update={cache => {
          const variables = { id: this.props.scheduleID }
          // get cached schedule
          const { schedule } = cache.readQuery({
            query,
            variables,
          })

          // update schedule
          cache.writeQuery({
            query,
            variables,
            data: {
              schedule: {
                ...schedule,
                isFavorite: !isFavorite,
              },
            },
          })
        }}
      >
        {mutation => this.renderForm(isFavorite, mutation)}
      </Mutation>
    )
  }

  renderForm(isFavorite, mutation) {
    const { scheduleID: id } = this.props
    return (
      <form
        onSubmit={e => {
          e.preventDefault()
          mutation({
            variables: {
              input: {
                target: { id, type: 'schedule' },
                favorite: !isFavorite,
              },
            },
          })
        }}
      >
        <IconButton
          aria-label={
            isFavorite
              ? 'Unset as a Favorite schedule'
              : 'Set as a Favorite schedule'
          }
          type='submit'
          color='inherit'
        >
          {isFavorite ? <FavoriteFilledIcon /> : <FavoriteBorderIcon />}
        </IconButton>
      </form>
    )
  }

  render() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.scheduleID }}
        render={({ data }) => {
          if (!data || !data.schedule) return null

          return this.renderMutation(
            data && data.schedule && data.schedule.isFavorite,
          )
        }}
      />
    )
  }
}
