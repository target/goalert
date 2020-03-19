// eslint-disable-next-line @typescript-eslint/no-unused-vars
namespace Cypress {
  interface Chainable {
    /** Creates a new user profile. */
    createUser: typeof createUser

    /** Creates multiple new user profiles. */
    createManyUsers: typeof createManyUsers

    /**
     * Resets the test user profile, including any existing contact methods.
     */
    resetProfile: typeof resetProfile

    /** Adds a contact method. If userID is missing, the test user's will be used. */
    addContactMethod: typeof addContactMethod

    /** Adds a notification rule. If userID is missing, the test user's will be used. */
    addNotificationRule: typeof addNotificationRule
  }
}

type UserRole = 'user' | 'admin'
interface Profile {
  id: string
  name: string
  email: string
  role: UserRole
}

interface UserOptions {
  name?: string
  email?: string
  role?: UserRole
}

type ContactMethodType = 'SMS' | 'VOICE'
interface ContactMethod {
  id: string
  userID: string
  name: string
  type: ContactMethodType
  value: string
}

interface ContactMethodOptions {
  userID?: string
  name?: string
  type?: ContactMethodType
  value?: string
}

interface NotificationRule {
  id: string
  userID: string
  contactMethodID: string
  contactMethod: ContactMethod
  delayMinutes: number
}

interface NotificationRuleOptions {
  userID?: string
  delayMinutes?: number
  contactMethodID?: string
  contactMethod?: ContactMethodOptions
}
