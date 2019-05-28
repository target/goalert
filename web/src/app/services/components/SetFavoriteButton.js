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
    service(id: $id) {
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
    serviceID: p.string.isRequired,
  }

  renderMutation(isFavorite) {
    return (
      <Mutation
        key='main'
        mutation={mutation}
        client={graphql2Client}
        update={cache => {
          const variables = { id: this.props.serviceID }
          // get cached service
          const { service } = cache.readQuery({
            query,
            variables,
          })

          // update service
          cache.writeQuery({
            query,
            variables,
            data: {
              service: {
                ...service,
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
    const { serviceID: id } = this.props
    return (
      <form
        onSubmit={e => {
          e.preventDefault()
          mutation({
            variables: {
              input: {
                target: { id, type: 'service' },
                favorite: !isFavorite,
              },
            },
          })
        }}
      >
        <IconButton
          aria-label={
            isFavorite
              ? 'Unset as a Favorite Service'
              : 'Set as a Favorite Service'
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
        variables={{ id: this.props.serviceID }}
        render={({ data }) => {
          if (!data || !data.service) return null

          return this.renderMutation(
            data && data.service && data.service.isFavorite,
          )
        }}
      />
    )
  }
}
