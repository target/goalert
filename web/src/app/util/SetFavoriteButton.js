import React from 'react'
import p from 'prop-types'
import IconButton from '@material-ui/core/IconButton'
import FavoriteFilledIcon from '@material-ui/icons/Star'
import FavoriteBorderIcon from '@material-ui/icons/StarBorder'

export default class SetFavoriteButtonClass extends React.PureComponent {
  static propTypes = {
    type: p.oneOf(['rotation', 'service']),
    onSubmit: p.func,
    isFavorite: p.bool,
  }

  render() {
    const { type, isFavorite, onSubmit } = this.props
    return (
      <form
        onSubmit={e => {
          e.preventDefault()
          onSubmit()
        }}
      >
        <IconButton
          aria-label={
            isFavorite
              ? `Unset as a Favorite ${type}`
              : `Set as a Favorite ${type}`
          }
          type='submit'
          color='inherit'
          data-cy={'set-fav'}
        >
          {isFavorite ? <FavoriteFilledIcon /> : <FavoriteBorderIcon />}
        </IconButton>
      </form>
    )
  }
}

export function SetFavoriteButton({ type, isFavorite, onSubmit }) {
  return (
    <form
      onSubmit={e => {
        e.preventDefault()
        onSubmit()
      }}
    >
      <IconButton
        aria-label={
          isFavorite
            ? `Unset as a Favorite ${type}`
            : `Set as a Favorite ${type}`
        }
        type='submit'
        color='inherit'
        data-cy={'set-fav'}
      >
        {isFavorite ? <FavoriteFilledIcon /> : <FavoriteBorderIcon />}
      </IconButton>
    </form>
  )
}

SetFavoriteButton.propTypes = {
  type: p.oneOf(['rotation', 'service']),
  onSubmit: p.func,
  isFavorite: p.bool,
}
