import React from 'react'
import IconButton from '@material-ui/core/IconButton'
import FavoriteFilledIcon from '@material-ui/icons/Star'
import FavoriteBorderIcon from '@material-ui/icons/StarBorder'
import Tooltip from '@material-ui/core/Tooltip'

interface SetFavoriteButtonProps {
  typeName: 'rotation' | 'service' | 'schedule'
  isFavorite?: boolean
  loading: boolean
  onClick: (event: React.FormEvent<HTMLFormElement>) => void
}

export function SetFavoriteButton({
  typeName,
  isFavorite,
  loading,
  onClick,
}: SetFavoriteButtonProps): JSX.Element | null {
  if (loading) {
    return null
  }

  const icon = isFavorite ? <FavoriteFilledIcon /> : <FavoriteBorderIcon />

  const content = (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        onClick(e)
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
        {icon}
      </IconButton>
    </form>
  )

  switch (typeName) {
    case 'service':
      return (
        <Tooltip
          title={
            isFavorite
              ? 'Unfavorite this service to stop seeing its alerts on your homepage'
              : 'Favorite this service to always see its alerts on your homepage'
          }
        >
          {content}
        </Tooltip>
      )
    default:
      return content
  }
}
