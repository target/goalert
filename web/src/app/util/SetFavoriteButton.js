import React from 'react'
import p from 'prop-types'
import IconButton from '@material-ui/core/IconButton'
import FavoriteFilledIcon from '@material-ui/icons/Star'
import FavoriteBorderIcon from '@material-ui/icons/StarBorder'

export function SetFavoriteButton({ typeName, isFavorite, onSubmit }) {
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
            ? `Unset as a Favorite ${typeName}`
            : `Set as a Favorite ${typeName}`
        }
        type='submit'
        color='inherit'
        data-cy='set-fav'
      >
        {isFavorite ? <FavoriteFilledIcon /> : <FavoriteBorderIcon />}
      </IconButton>
    </form>
  )
}

SetFavoriteButton.propTypes = {
  typeName: p.oneOf(['rotation', 'service']),
  onSubmit: p.func,
  isFavorite: p.bool,
}
