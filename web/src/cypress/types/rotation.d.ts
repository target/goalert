declare namespace Cypress {
  interface Chainable {
    /**
     * Creates a new rotation.
     */
    createRotation: typeof createRotation

    /** Delete the rotation with the specified ID */
    deleteRotation: typeof deleteRotation
  }
}

type RotationType = 'hourly' | 'daily' | 'weekly'
interface Rotation {
  id: string
  name: string
  description: string
  timeZone: string
  shiftLength: number
  type: RotationType
  start: string
  users: Array<{
    id: string
    name: string
    email: string
  }>
}

interface RotationOptions {
  name?: string
  description?: string
  timeZone?: string
  shiftLength?: number
  type?: RotationType
  start?: string
  favorite?: boolean

  /** Number of participants to add to the rotation. */
  count?: number
}
