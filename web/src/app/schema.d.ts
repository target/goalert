interface Query {
  user?: User
  users: UserConnection
  alert?: Alert
  alerts: AlertConnection
  service?: Service
  integrationKey?: IntegrationKey
  heartbeatMonitor?: HeartbeatMonitor
  services: ServiceConnection
  rotation?: Rotation
  rotations: RotationConnection
  schedule?: Schedule
  userCalendarSubscription?: UserCalendarSubscription
  schedules: ScheduleConnection
  escalationPolicy?: EscalationPolicy
  escalationPolicies: EscalationPolicyConnection
  authSubjectsForProvider: AuthSubjectConnection
  timeZones: TimeZoneConnection
  labels: LabelConnection
  labelKeys: StringConnection
  labelValues: StringConnection
  userOverrides: UserOverrideConnection
  userOverride?: UserOverride
  config: ConfigValue
  configHints: ConfigHint
  systemLimits: SystemLimit
  userContactMethod?: UserContactMethod
  slackChannels: SlackChannelConnection
  slackChannel?: SlackChannel
}

interface SlackChannelSearchOptions {
  first?: number
  after?: string
  search?: string
  omit?: string
}

interface SlackChannel {
  id: string
  name: string
}

interface SlackChannelConnection {
  nodes: SlackChannel
  pageInfo: PageInfo
}

interface SystemLimit {
  id: SystemLimitID
  description: string
  value: number
}

interface SystemLimitInput {
  id: SystemLimitID
  value: number
}

interface ConfigValue {
  id: string
  description: string
  value: string
  type: ConfigType
  password: boolean
}

interface ConfigHint {
  id: string
  value: string
}

enum ConfigType {
  string = 'string',
  stringList = 'stringList',
  integer = 'integer',
  boolean = 'boolean',
}

enum SystemLimitID {
  CalendarSubscriptionsPerUser = 'CalendarSubscriptionsPerUser',
  NotificationRulesPerUser = 'NotificationRulesPerUser',
  ContactMethodsPerUser = 'ContactMethodsPerUser',
  EPStepsPerPolicy = 'EPStepsPerPolicy',
  EPActionsPerStep = 'EPActionsPerStep',
  ParticipantsPerRotation = 'ParticipantsPerRotation',
  RulesPerSchedule = 'RulesPerSchedule',
  IntegrationKeysPerService = 'IntegrationKeysPerService',
  UnackedAlertsPerService = 'UnackedAlertsPerService',
  TargetsPerSchedule = 'TargetsPerSchedule',
  HeartbeatMonitorsPerService = 'HeartbeatMonitorsPerService',
  UserOverridesPerSchedule = 'UserOverridesPerSchedule',
}

interface UserOverrideSearchOptions {
  first?: number
  after?: string
  omit?: string
  scheduleID?: string
  filterAddUserID?: string
  filterRemoveUserID?: string
  filterAnyUserID?: string
  start?: ISOTimestamp
  end?: ISOTimestamp
}

interface UserOverrideConnection {
  nodes: UserOverride
  pageInfo: PageInfo
}

interface UserOverride {
  id: string
  start: ISOTimestamp
  end: ISOTimestamp
  addUserID: string
  removeUserID: string
  addUser?: User
  removeUser?: User
  target: Target
}

interface LabelSearchOptions {
  first?: number
  after?: string
  search?: string
  uniqueKeys?: boolean
  omit?: string
}

interface LabelKeySearchOptions {
  first?: number
  after?: string
  search?: string
  omit?: string
}

interface LabelValueSearchOptions {
  key: string
  first?: number
  after?: string
  search?: string
  omit?: string
}

interface LabelConnection {
  nodes: Label
  pageInfo: PageInfo
}

interface StringConnection {
  nodes: string
  pageInfo: PageInfo
}

interface Mutation {
  addAuthSubject: boolean
  deleteAuthSubject: boolean
  updateUser: boolean
  testContactMethod: boolean
  updateAlerts?: Alert
  updateRotation: boolean
  escalateAlerts?: Alert
  setFavorite: boolean
  updateService: boolean
  updateEscalationPolicy: boolean
  updateEscalationPolicyStep: boolean
  deleteAll: boolean
  createAlert?: Alert
  createService?: Service
  createEscalationPolicy?: EscalationPolicy
  createEscalationPolicyStep?: EscalationPolicyStep
  createRotation?: Rotation
  createIntegrationKey?: IntegrationKey
  createHeartbeatMonitor?: HeartbeatMonitor
  setLabel: boolean
  createSchedule?: Schedule
  createUserCalendarSubscription: UserCalendarSubscription
  updateUserCalendarSubscription: boolean
  updateScheduleTarget: boolean
  createUserOverride?: UserOverride
  createUserContactMethod?: UserContactMethod
  createUserNotificationRule?: UserNotificationRule
  updateUserContactMethod: boolean
  sendContactMethodVerification: boolean
  verifyContactMethod: boolean
  updateSchedule: boolean
  updateUserOverride: boolean
  updateHeartbeatMonitor: boolean
  setConfig: boolean
  setSystemLimits: boolean
}

interface CreateAlertInput {
  summary: string
  details?: string
  serviceID: string
}

interface CreateUserCalendarSubscriptionInput {
  name: string
  reminderMinutes?: number
  scheduleID: string
  disabled?: boolean
}

interface UpdateUserCalendarSubscriptionInput {
  id: string
  name?: string
  reminderMinutes?: number
  disabled?: boolean
}

interface UserCalendarSubscription {
  id: string
  name: string
  reminderMinutes: number
  scheduleID: string
  schedule?: Schedule
  lastAccess: ISOTimestamp
  disabled: boolean
  url?: string
}

interface ConfigValueInput {
  id: string
  value: string
}

interface UpdateUserOverrideInput {
  id: string
  start?: ISOTimestamp
  end?: ISOTimestamp
  addUserID?: string
  removeUserID?: string
}

interface CreateUserOverrideInput {
  scheduleID?: string
  start: ISOTimestamp
  end: ISOTimestamp
  addUserID?: string
  removeUserID?: string
}

interface CreateScheduleInput {
  name: string
  description?: string
  timeZone: string
  favorite?: boolean
  targets?: ScheduleTargetInput
  newUserOverrides?: CreateUserOverrideInput
}

interface ScheduleTargetInput {
  scheduleID?: string
  target?: TargetInput
  newRotation?: CreateRotationInput
  rules: ScheduleRuleInput
}

interface ScheduleRuleInput {
  id?: string
  start?: ClockTime
  end?: ClockTime
  weekdayFilter?: boolean
}

interface SetLabelInput {
  target?: TargetInput
  key: string
  value: string
}

interface TimeZoneSearchOptions {
  first?: number
  after?: string
  search?: string
  omit?: string
}

interface TimeZoneConnection {
  nodes: TimeZone
  pageInfo: PageInfo
}

interface TimeZone {
  id: string
}

interface CreateServiceInput {
  name: string
  description?: string
  favorite?: boolean
  escalationPolicyID?: string
  newEscalationPolicy?: CreateEscalationPolicyInput
  newIntegrationKeys?: CreateIntegrationKeyInput
  labels?: SetLabelInput
  newHeartbeatMonitors?: CreateHeartbeatMonitorInput
}

interface CreateEscalationPolicyInput {
  name: string
  description?: string
  repeat?: number
  steps?: CreateEscalationPolicyStepInput
}

interface CreateEscalationPolicyStepInput {
  escalationPolicyID?: string
  delayMinutes: number
  targets?: TargetInput
  newRotation?: CreateRotationInput
  newSchedule?: CreateScheduleInput
}

interface EscalationPolicyStep {
  id: string
  stepNumber: number
  delayMinutes: number
  targets: Target
  escalationPolicy?: EscalationPolicy
}

interface UpdateScheduleInput {
  id: string
  name?: string
  description?: string
  timeZone?: string
}

interface UpdateServiceInput {
  id: string
  name?: string
  description?: string
  escalationPolicyID?: string
}

interface UpdateEscalationPolicyInput {
  id: string
  name?: string
  description?: string
  repeat?: number
  stepIDs?: string
}

interface UpdateEscalationPolicyStepInput {
  id: string
  delayMinutes?: number
  targets?: TargetInput
}

interface SetFavoriteInput {
  target: TargetInput
  favorite: boolean
}

interface EscalationPolicyConnection {
  nodes: EscalationPolicy
  pageInfo: PageInfo
}

interface AlertConnection {
  nodes: Alert
  pageInfo: PageInfo
}

interface ScheduleConnection {
  nodes: Schedule
  pageInfo: PageInfo
}

interface Schedule {
  id: string
  name: string
  description: string
  timeZone: string
  assignedTo: Target
  shifts: OnCallShift
  targets: ScheduleTarget
  target?: ScheduleTarget
  isFavorite: boolean
}

interface OnCallShift {
  userID: string
  user?: User
  start: ISOTimestamp
  end: ISOTimestamp
  truncated: boolean
}

interface ScheduleTarget {
  scheduleID: string
  target: Target
  rules: ScheduleRule
}

interface ScheduleRule {
  id: string
  scheduleID: string
  start: ClockTime
  end: ClockTime
  weekdayFilter: boolean
  target: Target
}

interface RotationConnection {
  nodes: Rotation
  pageInfo: PageInfo
}

interface CreateRotationInput {
  name: string
  description?: string
  timeZone: string
  start: ISOTimestamp
  favorite?: boolean
  type: RotationType
  shiftLength?: number
  userIDs?: string
}

interface Rotation {
  id: string
  name: string
  description: string
  isFavorite: boolean
  start: ISOTimestamp
  timeZone: string
  type: RotationType
  shiftLength: number
  activeUserIndex: number
  userIDs: string
  users: User
  nextHandoffTimes: ISOTimestamp
}

enum RotationType {
  weekly = 'weekly',
  daily = 'daily',
  hourly = 'hourly',
}

interface UpdateAlertsInput {
  alertIDs: number
  newStatus: AlertStatus
}

interface UpdateRotationInput {
  id: string
  name?: string
  description?: string
  timeZone?: string
  start?: ISOTimestamp
  type?: RotationType
  shiftLength?: number
  activeUserIndex?: number
  userIDs?: string
}

interface RotationSearchOptions {
  first?: number
  after?: string
  search?: string
  omit?: string
  favoritesOnly?: boolean
  favoritesFirst?: boolean
}

interface EscalationPolicySearchOptions {
  first?: number
  after?: string
  search?: string
  omit?: string
}

interface ScheduleSearchOptions {
  first?: number
  after?: string
  search?: string
  omit?: string
  favoritesOnly?: boolean
  favoritesFirst?: boolean
}

interface ServiceSearchOptions {
  first?: number
  after?: string
  search?: string
  omit?: string
  favoritesOnly?: boolean
  favoritesFirst?: boolean
}

interface UserSearchOptions {
  first?: number
  after?: string
  search?: string
  omit?: string
}

interface AlertSearchOptions {
  filterByStatus?: AlertStatus
  filterByServiceID?: string
  search?: string
  first?: number
  after?: string
  favoritesOnly?: boolean
  omit?: number
}

type ISOTimestamp = string

type ClockTime = string

interface Alert {
  id: string
  alertID: number
  status: AlertStatus
  summary: string
  details: string
  createdAt: ISOTimestamp
  serviceID: string
  service?: Service
  state?: AlertState
  recentEvents: AlertLogEntryConnection
}

interface AlertRecentEventsOptions {
  limit?: number
  after?: string
}

interface AlertLogEntryConnection {
  nodes: AlertLogEntry
  pageInfo: PageInfo
}

interface AlertLogEntry {
  id: number
  timestamp: ISOTimestamp
  message: string
}

interface AlertState {
  lastEscalation: ISOTimestamp
  stepNumber: number
  repeatCount: number
}

interface Service {
  id: string
  name: string
  description: string
  escalationPolicyID: string
  escalationPolicy?: EscalationPolicy
  isFavorite: boolean
  onCallUsers: ServiceOnCallUser
  integrationKeys: IntegrationKey
  labels: Label
  heartbeatMonitors: HeartbeatMonitor
}

interface CreateIntegrationKeyInput {
  serviceID?: string
  type: IntegrationKeyType
  name: string
}

interface CreateHeartbeatMonitorInput {
  serviceID: string
  name: string
  timeoutMinutes: number
}

interface UpdateHeartbeatMonitorInput {
  id: string
  name?: string
  timeoutMinutes?: number
}

enum HeartbeatMonitorState {
  inactive = 'inactive',
  healthy = 'healthy',
  unhealthy = 'unhealthy',
}

interface HeartbeatMonitor {
  id: string
  serviceID: string
  name: string
  timeoutMinutes: number
  lastState: HeartbeatMonitorState
  lastHeartbeat?: ISOTimestamp
  href: string
}

interface Label {
  key: string
  value: string
}

interface IntegrationKey {
  id: string
  serviceID: string
  type: IntegrationKeyType
  name: string
  href: string
}

enum IntegrationKeyType {
  generic = 'generic',
  grafana = 'grafana',
  site24x7 = 'site24x7',
  email = 'email',
}

interface ServiceOnCallUser {
  userID: string
  userName: string
  stepNumber: number
}

interface EscalationPolicy {
  id: string
  name: string
  description: string
  repeat: number
  assignedTo: Target
  steps: EscalationPolicyStep
}

enum AlertStatus {
  StatusAcknowledged = 'StatusAcknowledged',
  StatusClosed = 'StatusClosed',
  StatusUnacknowledged = 'StatusUnacknowledged',
}

interface Target {
  id: string
  type: TargetType
  name?: string
}

interface TargetInput {
  id: string
  type: TargetType
}

enum TargetType {
  escalationPolicy = 'escalationPolicy',
  notificationChannel = 'notificationChannel',
  slackChannel = 'slackChannel',
  notificationPolicy = 'notificationPolicy',
  rotation = 'rotation',
  service = 'service',
  schedule = 'schedule',
  user = 'user',
  integrationKey = 'integrationKey',
  userOverride = 'userOverride',
  notificationRule = 'notificationRule',
  contactMethod = 'contactMethod',
  heartbeatMonitor = 'heartbeatMonitor',
  calendarSubscription = 'calendarSubscription',
}

interface ServiceConnection {
  nodes: Service
  pageInfo: PageInfo
}

interface UserConnection {
  nodes: User
  pageInfo: PageInfo
}

interface AuthSubjectConnection {
  nodes: AuthSubject
  pageInfo: PageInfo
}

interface PageInfo {
  endCursor?: string
  hasNextPage: boolean
}

interface UpdateUserInput {
  id: string
  name?: string
  email?: string
  role?: UserRole
  statusUpdateContactMethodID?: string
}

interface AuthSubjectInput {
  userID: string
  providerID: string
  subjectID: string
}

enum UserRole {
  unknown = 'unknown',
  user = 'user',
  admin = 'admin',
}

interface User {
  id: string
  role: UserRole
  name: string
  email: string
  contactMethods: UserContactMethod
  notificationRules: UserNotificationRule
  calendarSubscriptions: UserCalendarSubscription
  statusUpdateContactMethodID: string
  authSubjects: AuthSubject
  onCallSteps: EscalationPolicyStep
}

interface UserNotificationRule {
  id: string
  delayMinutes: number
  contactMethodID: string
  contactMethod?: UserContactMethod
}

enum ContactMethodType {
  SMS = 'SMS',
  VOICE = 'VOICE',
}

interface UserContactMethod {
  id: string
  type?: ContactMethodType
  name: string
  value: string
  formattedValue: string
  disabled: boolean
}

interface CreateUserContactMethodInput {
  userID: string
  type: ContactMethodType
  name: string
  value: string
  newUserNotificationRule?: CreateUserNotificationRuleInput
}

interface CreateUserNotificationRuleInput {
  userID?: string
  contactMethodID?: string
  delayMinutes: number
}

interface UpdateUserContactMethodInput {
  id: string
  name?: string
  value?: string
}

interface SendContactMethodVerificationInput {
  contactMethodID: string
}

interface VerifyContactMethodInput {
  contactMethodID: string
  code: number
}

interface AuthSubject {
  providerID: string
  subjectID: string
  userID: string
}
