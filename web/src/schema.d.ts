// Code generated by devtools/gqltsgen DO NOT EDIT.

export interface Action {
  dest: Destination
  params: ExprStringMap
}

export interface ActionInput {
  dest: DestinationInput
  params: ExprStringMap
}

export interface Alert {
  alertID: number
  createdAt: ISOTimestamp
  details: string
  id: string
  meta?: null | AlertMetadata[]
  metaValue: string
  metrics?: null | AlertMetric
  noiseReason?: null | string
  pendingNotifications: AlertPendingNotification[]
  recentEvents: AlertLogEntryConnection
  service?: null | Service
  serviceID: string
  state?: null | AlertState
  status: AlertStatus
  summary: string
}

export interface AlertConnection {
  nodes: Alert[]
  pageInfo: PageInfo
}

export interface AlertDataPoint {
  alertCount: number
  timestamp: ISOTimestamp
}

export interface AlertLogEntry {
  id: number
  message: string
  state?: null | NotificationState
  timestamp: ISOTimestamp
}

export interface AlertLogEntryConnection {
  nodes: AlertLogEntry[]
  pageInfo: PageInfo
}

export interface AlertMetadata {
  key: string
  value: string
}

export interface AlertMetadataInput {
  key: string
  value: string
}

export interface AlertMetric {
  closedAt: ISOTimestamp
  escalated: boolean
  timeToAck: ISODuration
  timeToClose: ISODuration
}

export interface AlertMetricsOptions {
  filterByServiceID?: null | string[]
  rInterval: ISORInterval
}

export interface AlertPendingNotification {
  destination: string
}

export interface AlertRecentEventsOptions {
  after?: null | string
  limit?: null | number
}

export interface AlertSearchOptions {
  after?: null | string
  closedBefore?: null | ISOTimestamp
  createdBefore?: null | ISOTimestamp
  favoritesOnly?: null | boolean
  filterByServiceID?: null | string[]
  filterByStatus?: null | AlertStatus[]
  first?: null | number
  includeNotified?: null | boolean
  notClosedBefore?: null | ISOTimestamp
  notCreatedBefore?: null | ISOTimestamp
  omit?: null | number[]
  search?: null | string
  sort?: null | AlertSearchSort
}

export type AlertSearchSort = 'dateID' | 'dateIDReverse' | 'statusID'

export interface AlertState {
  lastEscalation: ISOTimestamp
  repeatCount: number
  stepNumber: number
}

export type AlertStatus =
  | 'StatusAcknowledged'
  | 'StatusClosed'
  | 'StatusUnacknowledged'

export interface AuthSubject {
  providerID: string
  subjectID: string
  userID: string
}

export interface AuthSubjectConnection {
  nodes: AuthSubject[]
  pageInfo: PageInfo
}

export interface AuthSubjectInput {
  providerID: string
  subjectID: string
  userID: string
}

export type Boolean = string

export interface CalcRotationHandoffTimesInput {
  count: number
  from?: null | ISOTimestamp
  handoff: ISOTimestamp
  shiftLength?: null | ISODuration
  shiftLengthHours?: null | number
  timeZone: string
}

export interface Clause {
  field: ExprIdentifier
  negate: boolean
  operator: ExprOperator
  value: ExprValue
}

export interface ClauseInput {
  field: ExprIdentifier
  negate: boolean
  operator: ExprOperator
  value: ExprValue
}

export interface ClearTemporarySchedulesInput {
  end: ISOTimestamp
  scheduleID: string
  start: ISOTimestamp
}

export type ClockTime = string

export interface CloseMatchingAlertInput {
  dedup?: null | string
  details?: null | string
  serviceID: string
  summary?: null | string
}

export interface Condition {
  clauses: Clause[]
}

export interface ConditionInput {
  clauses: ClauseInput[]
}

export interface ConditionToExprInput {
  condition: ConditionInput
}

export interface ConfigHint {
  id: string
  value: string
}

export type ConfigType = 'boolean' | 'integer' | 'string' | 'stringList'

export interface ConfigValue {
  deprecated: string
  description: string
  id: string
  password: boolean
  type: ConfigType
  value: string
}

export interface ConfigValueInput {
  id: string
  value: string
}

export type ContactMethodType =
  | 'EMAIL'
  | 'SLACK_DM'
  | 'SMS'
  | 'VOICE'
  | 'WEBHOOK'

export interface CreateAlertInput {
  dedup?: null | string
  details?: null | string
  meta?: null | AlertMetadataInput[]
  sanitize?: null | boolean
  serviceID: string
  summary: string
}

export interface CreateBasicAuthInput {
  password: string
  userID: string
  username: string
}

export interface CreateEscalationPolicyInput {
  description?: null | string
  favorite?: null | boolean
  name: string
  repeat?: null | number
  steps?: null | CreateEscalationPolicyStepInput[]
}

export interface CreateEscalationPolicyStepInput {
  actions?: null | DestinationInput[]
  delayMinutes: number
  escalationPolicyID?: null | string
  newRotation?: null | CreateRotationInput
  newSchedule?: null | CreateScheduleInput
  targets?: null | TargetInput[]
}

export interface CreateGQLAPIKeyInput {
  description: string
  expiresAt: ISOTimestamp
  name: string
  query: string
  role: UserRole
}

export interface CreateHeartbeatMonitorInput {
  additionalDetails?: null | string
  name: string
  serviceID?: null | string
  timeoutMinutes: number
}

export interface CreateIntegrationKeyInput {
  externalSystemName?: null | string
  name: string
  serviceID?: null | string
  type: IntegrationKeyType
}

export interface CreateRotationInput {
  description?: null | string
  favorite?: null | boolean
  name: string
  shiftLength?: null | number
  start: ISOTimestamp
  timeZone: string
  type: RotationType
  userIDs?: null | string[]
}

export interface CreateScheduleInput {
  description?: null | string
  favorite?: null | boolean
  name: string
  newUserOverrides?: null | CreateUserOverrideInput[]
  targets?: null | ScheduleTargetInput[]
  timeZone: string
}

export interface CreateServiceInput {
  description?: null | string
  escalationPolicyID?: null | string
  favorite?: null | boolean
  labels?: null | SetLabelInput[]
  name: string
  newEscalationPolicy?: null | CreateEscalationPolicyInput
  newHeartbeatMonitors?: null | CreateHeartbeatMonitorInput[]
  newIntegrationKeys?: null | CreateIntegrationKeyInput[]
}

export interface CreateUserCalendarSubscriptionInput {
  disabled?: null | boolean
  fullSchedule?: null | boolean
  name: string
  reminderMinutes?: null | number[]
  scheduleID: string
}

export interface CreateUserContactMethodInput {
  dest?: null | DestinationInput
  enableStatusUpdates?: null | boolean
  name: string
  newUserNotificationRule?: null | CreateUserNotificationRuleInput
  type?: null | ContactMethodType
  userID: string
  value?: null | string
}

export interface CreateUserInput {
  email?: null | string
  favorite?: null | boolean
  name?: null | string
  password: string
  role?: null | UserRole
  username: string
}

export interface CreateUserNotificationRuleInput {
  contactMethodID?: null | string
  delayMinutes: number
  userID?: null | string
}

export interface CreateUserOverrideInput {
  addUserID?: null | string
  end: ISOTimestamp
  removeUserID?: null | string
  scheduleID?: null | string
  start: ISOTimestamp
}

export interface CreatedGQLAPIKey {
  id: string
  token: string
}

export interface DebugCarrierInfo {
  mobileCountryCode: string
  mobileNetworkCode: string
  name: string
  type: string
}

export interface DebugCarrierInfoInput {
  number: string
}

export interface DebugMessage {
  alertID?: null | number
  createdAt: ISOTimestamp
  destination: string
  id: string
  providerID?: null | string
  retryCount: number
  sentAt?: null | ISOTimestamp
  serviceID?: null | string
  serviceName?: null | string
  source?: null | string
  status: string
  type: string
  updatedAt: ISOTimestamp
  userID?: null | string
  userName?: null | string
}

export interface DebugMessageStatusInfo {
  state: NotificationState
}

export interface DebugMessageStatusInput {
  providerMessageID: string
}

export interface DebugMessagesInput {
  createdAfter?: null | ISOTimestamp
  createdBefore?: null | ISOTimestamp
  first?: null | number
}

export interface DebugSendSMSInfo {
  fromNumber: string
  id: string
  providerURL: string
}

export interface DebugSendSMSInput {
  body: string
  from: string
  to: string
}

export interface Destination {
  displayInfo: InlineDisplayInfo
  type: DestinationType
  values: FieldValuePair[]
}

export interface DestinationDisplayInfo {
  iconAltText: string
  iconURL: string
  linkURL: string
  text: string
}

export interface DestinationDisplayInfoError {
  error: string
}

export interface DestinationFieldConfig {
  fieldID: string
  hint: string
  hintURL: string
  inputType: string
  label: string
  placeholderText: string
  prefix: string
  supportsSearch: boolean
  supportsValidation: boolean
}

export interface DestinationFieldSearchInput {
  after?: null | string
  destType: DestinationType
  fieldID: string
  first?: null | number
  omit?: null | string[]
  search?: null | string
}

export interface DestinationFieldValidateInput {
  destType: DestinationType
  fieldID: string
  value: string
}

export interface DestinationInput {
  type: DestinationType
  values: FieldValueInput[]
}

export type DestinationType = string

export interface DestinationTypeInfo {
  dynamicParams: DynamicParamConfig[]
  enabled: boolean
  iconAltText: string
  iconURL: string
  isContactMethod: boolean
  isDynamicAction: boolean
  isEPTarget: boolean
  isSchedOnCallNotify: boolean
  name: string
  requiredFields: DestinationFieldConfig[]
  statusUpdatesRequired: boolean
  supportsStatusUpdates: boolean
  type: DestinationType
  userDisclaimer: string
}

export interface DynamicParamConfig {
  hint: string
  hintURL: string
  label: string
  paramID: string
}

export type ErrorCode =
  | 'EXPR_TOO_COMPLEX'
  | 'INVALID_DEST_FIELD_VALUE'
  | 'INVALID_INPUT_VALUE'
  | 'INVALID_MAP_FIELD_VALUE'

export interface EscalationPolicy {
  assignedTo: Target[]
  description: string
  id: string
  isFavorite: boolean
  name: string
  notices: Notice[]
  repeat: number
  steps: EscalationPolicyStep[]
}

export interface EscalationPolicyConnection {
  nodes: EscalationPolicy[]
  pageInfo: PageInfo
}

export interface EscalationPolicySearchOptions {
  after?: null | string
  favoritesFirst?: null | boolean
  favoritesOnly?: null | boolean
  first?: null | number
  omit?: null | string[]
  search?: null | string
}

export interface EscalationPolicyStep {
  actions: Destination[]
  delayMinutes: number
  escalationPolicy?: null | EscalationPolicy
  id: string
  stepNumber: number
  targets: Target[]
}

export interface Expr {
  conditionToExpr: string
  exprToCondition: Condition
}

export type ExprBooleanExpression = string

export type ExprExpression = string

export type ExprIdentifier = string

export type ExprOperator = string

export type ExprStringExpression = string

export type ExprStringMap = Record<string, string>

export interface ExprToConditionInput {
  expr: ExprBooleanExpression
}

export type ExprValue = string

export interface FieldSearchConnection {
  nodes: FieldSearchResult[]
  pageInfo: PageInfo
}

export interface FieldSearchResult {
  fieldID: string
  isFavorite: boolean
  label: string
  value: string
}

export interface FieldValueInput {
  fieldID: string
  value: string
}

export interface FieldValuePair {
  fieldID: string
  value: string
}

export type Float = string

export interface GQLAPIKey {
  createdAt: ISOTimestamp
  createdBy?: null | User
  description: string
  expiresAt: ISOTimestamp
  id: string
  lastUsed?: null | GQLAPIKeyUsage
  name: string
  query: string
  role: UserRole
  updatedAt: ISOTimestamp
  updatedBy?: null | User
}

export interface GQLAPIKeyUsage {
  ip: string
  time: ISOTimestamp
  ua: string
}

export interface HeartbeatMonitor {
  additionalDetails: string
  href: string
  id: string
  lastHeartbeat?: null | ISOTimestamp
  lastState: HeartbeatMonitorState
  name: string
  serviceID: string
  timeoutMinutes: number
}

export type HeartbeatMonitorState = 'healthy' | 'inactive' | 'unhealthy'

export type ID = string

export type ISODuration = string

export type ISORInterval = string

export type ISOTimestamp = string

export type InlineDisplayInfo =
  | DestinationDisplayInfo
  | DestinationDisplayInfoError

export type Int = string

export interface IntegrationKey {
  config: KeyConfig
  externalSystemName?: null | string
  href: string
  id: string
  name: string
  serviceID: string
  tokenInfo: TokenInfo
  type: IntegrationKeyType
}

export interface IntegrationKeyConnection {
  nodes: IntegrationKey[]
  pageInfo: PageInfo
}

export interface IntegrationKeySearchOptions {
  after?: null | string
  first?: null | number
  omit?: null | string[]
  search?: null | string
}

export type IntegrationKeyType =
  | 'email'
  | 'generic'
  | 'grafana'
  | 'prometheusAlertmanager'
  | 'site24x7'
  | 'universal'

export interface IntegrationKeyTypeInfo {
  enabled: boolean
  id: string
  label: string
  name: string
}

export interface KeyConfig {
  defaultActions: Action[]
  oneRule?: null | KeyRule
  rules: KeyRule[]
}

export interface KeyRule {
  actions: Action[]
  conditionExpr: ExprBooleanExpression
  continueAfterMatch: boolean
  description: string
  id: string
  name: string
}

export interface KeyRuleInput {
  actions: ActionInput[]
  conditionExpr: ExprBooleanExpression
  continueAfterMatch: boolean
  description: string
  id?: null | string
  name: string
}

export interface Label {
  key: string
  value: string
}

export interface LabelConnection {
  nodes: Label[]
  pageInfo: PageInfo
}

export interface LabelKeySearchOptions {
  after?: null | string
  first?: null | number
  omit?: null | string[]
  search?: null | string
}

export interface LabelSearchOptions {
  after?: null | string
  first?: null | number
  omit?: null | string[]
  search?: null | string
  uniqueKeys?: null | boolean
}

export interface LabelValueSearchOptions {
  after?: null | string
  first?: null | number
  key: string
  omit?: null | string[]
  search?: null | string
}

export interface LinkAccountInfo {
  alertID?: null | number
  alertNewStatus?: null | AlertStatus
  userDetails: string
}

export interface MessageLogConnection {
  nodes: DebugMessage[]
  pageInfo: PageInfo
  stats: MessageLogConnectionStats
}

export interface MessageLogConnectionStats {
  timeSeries: TimeSeriesBucket[]
}

export interface MessageLogSearchOptions {
  after?: null | string
  createdAfter?: null | ISOTimestamp
  createdBefore?: null | ISOTimestamp
  first?: null | number
  omit?: null | string[]
  search?: null | string
}

export interface Mutation {
  addAuthSubject: boolean
  clearTemporarySchedules: boolean
  closeMatchingAlert: boolean
  createAlert?: null | Alert
  createBasicAuth: boolean
  createEscalationPolicy?: null | EscalationPolicy
  createEscalationPolicyStep?: null | EscalationPolicyStep
  createGQLAPIKey: CreatedGQLAPIKey
  createHeartbeatMonitor?: null | HeartbeatMonitor
  createIntegrationKey?: null | IntegrationKey
  createRotation?: null | Rotation
  createSchedule?: null | Schedule
  createService?: null | Service
  createUser?: null | User
  createUserCalendarSubscription: UserCalendarSubscription
  createUserContactMethod?: null | UserContactMethod
  createUserNotificationRule?: null | UserNotificationRule
  createUserOverride?: null | UserOverride
  debugCarrierInfo: DebugCarrierInfo
  debugSendSMS?: null | DebugSendSMSInfo
  deleteAll: boolean
  deleteAuthSubject: boolean
  deleteGQLAPIKey: boolean
  deleteSecondaryToken: boolean
  endAllAuthSessionsByCurrentUser: boolean
  escalateAlerts?: null | Alert[]
  generateKeyToken: string
  linkAccount: boolean
  promoteSecondaryToken: boolean
  sendContactMethodVerification: boolean
  setAlertNoiseReason: boolean
  setConfig: boolean
  setFavorite: boolean
  setLabel: boolean
  setScheduleOnCallNotificationRules: boolean
  setSystemLimits: boolean
  setTemporarySchedule: boolean
  swoAction: boolean
  testContactMethod: boolean
  updateAlerts?: null | Alert[]
  updateAlertsByService: boolean
  updateBasicAuth: boolean
  updateEscalationPolicy: boolean
  updateEscalationPolicyStep: boolean
  updateGQLAPIKey: boolean
  updateHeartbeatMonitor: boolean
  updateKeyConfig: boolean
  updateRotation: boolean
  updateSchedule: boolean
  updateScheduleTarget: boolean
  updateService: boolean
  updateUser: boolean
  updateUserCalendarSubscription: boolean
  updateUserContactMethod: boolean
  updateUserOverride: boolean
  verifyContactMethod: boolean
}

export interface Notice {
  details: string
  message: string
  type: NoticeType
}

export type NoticeType = 'ERROR' | 'INFO' | 'WARNING'

export interface NotificationState {
  details: string
  formattedSrcValue: string
  status?: null | NotificationStatus
}

export type NotificationStatus = 'ERROR' | 'OK' | 'WARN'

export interface OnCallNotificationRule {
  dest: Destination
  id: string
  target: Target
  time?: null | ClockTime
  weekdayFilter?: null | WeekdayFilter
}

export interface OnCallNotificationRuleInput {
  dest?: null | DestinationInput
  id?: null | string
  target?: null | TargetInput
  time?: null | ClockTime
  weekdayFilter?: null | WeekdayFilter
}

export interface OnCallOverview {
  serviceAssignments: OnCallServiceAssignment[]
  serviceCount: number
}

export interface OnCallServiceAssignment {
  escalationPolicyID: string
  escalationPolicyName: string
  serviceID: string
  serviceName: string
  stepNumber: number
}

export interface OnCallShift {
  end: ISOTimestamp
  start: ISOTimestamp
  truncated: boolean
  user?: null | User
  userID: string
}

export interface PageInfo {
  endCursor?: null | string
  hasNextPage: boolean
}

export interface PhoneNumberInfo {
  countryCode: string
  error: string
  formatted: string
  id: string
  regionCode: string
  valid: boolean
}

export interface Query {
  __schema: __Schema
  __type?: null | __Type
  actionInputValidate: boolean
  alert?: null | Alert
  alerts: AlertConnection
  authSubjectsForProvider: AuthSubjectConnection
  calcRotationHandoffTimes: ISOTimestamp[]
  config: ConfigValue[]
  configHints: ConfigHint[]
  debugMessageStatus: DebugMessageStatusInfo
  debugMessages: DebugMessage[]
  destinationDisplayInfo: DestinationDisplayInfo
  destinationFieldSearch: FieldSearchConnection
  destinationFieldValidate: boolean
  destinationFieldValueName: string
  destinationTypes: DestinationTypeInfo[]
  escalationPolicies: EscalationPolicyConnection
  escalationPolicy?: null | EscalationPolicy
  experimentalFlags: string[]
  expr: Expr
  generateSlackAppManifest: string
  gqlAPIKeys: GQLAPIKey[]
  heartbeatMonitor?: null | HeartbeatMonitor
  integrationKey?: null | IntegrationKey
  integrationKeyTypes: IntegrationKeyTypeInfo[]
  integrationKeys: IntegrationKeyConnection
  labelKeys: StringConnection
  labelValues: StringConnection
  labels: LabelConnection
  linkAccountInfo?: null | LinkAccountInfo
  messageLogs: MessageLogConnection
  phoneNumberInfo?: null | PhoneNumberInfo
  rotation?: null | Rotation
  rotations: RotationConnection
  schedule?: null | Schedule
  schedules: ScheduleConnection
  service?: null | Service
  services: ServiceConnection
  slackChannel?: null | SlackChannel
  slackChannels: SlackChannelConnection
  slackUserGroup?: null | SlackUserGroup
  slackUserGroups: SlackUserGroupConnection
  swoStatus: SWOStatus
  systemLimits: SystemLimit[]
  timeZones: TimeZoneConnection
  user?: null | User
  userCalendarSubscription?: null | UserCalendarSubscription
  userContactMethod?: null | UserContactMethod
  userOverride?: null | UserOverride
  userOverrides: UserOverrideConnection
  users: UserConnection
}

export interface Rotation {
  activeUserIndex: number
  description: string
  id: string
  isFavorite: boolean
  name: string
  nextHandoffTimes: ISOTimestamp[]
  shiftLength: number
  start: ISOTimestamp
  timeZone: string
  type: RotationType
  userIDs: string[]
  users: User[]
}

export interface RotationConnection {
  nodes: Rotation[]
  pageInfo: PageInfo
}

export interface RotationSearchOptions {
  after?: null | string
  favoritesFirst?: null | boolean
  favoritesOnly?: null | boolean
  first?: null | number
  omit?: null | string[]
  search?: null | string
}

export type RotationType = 'daily' | 'hourly' | 'monthly' | 'weekly'

export type SWOAction = 'execute' | 'reset'

export interface SWOConnection {
  count: number
  isNext: boolean
  name: string
  type: string
  version: string
}

export interface SWONode {
  canExec: boolean
  configError: string
  connections?: null | SWOConnection[]
  id: string
  isLeader: boolean
  uptime: string
}

export type SWOState =
  | 'done'
  | 'executing'
  | 'idle'
  | 'pausing'
  | 'resetting'
  | 'syncing'
  | 'unknown'

export interface SWOStatus {
  lastError: string
  lastStatus: string
  mainDBVersion: string
  nextDBVersion: string
  nodes: SWONode[]
  state: SWOState
}

export interface Schedule {
  assignedTo: Target[]
  description: string
  id: string
  isFavorite: boolean
  name: string
  onCallNotificationRules: OnCallNotificationRule[]
  shifts: OnCallShift[]
  target?: null | ScheduleTarget
  targets: ScheduleTarget[]
  temporarySchedules: TemporarySchedule[]
  timeZone: string
}

export interface ScheduleConnection {
  nodes: Schedule[]
  pageInfo: PageInfo
}

export interface ScheduleRule {
  end: ClockTime
  id: string
  scheduleID: string
  start: ClockTime
  target: Target
  weekdayFilter: WeekdayFilter
}

export interface ScheduleRuleInput {
  end?: null | ClockTime
  id?: null | string
  start?: null | ClockTime
  weekdayFilter?: null | WeekdayFilter
}

export interface ScheduleSearchOptions {
  after?: null | string
  favoritesFirst?: null | boolean
  favoritesOnly?: null | boolean
  first?: null | number
  omit?: null | string[]
  search?: null | string
}

export interface ScheduleTarget {
  rules: ScheduleRule[]
  scheduleID: string
  target: Target
}

export interface ScheduleTargetInput {
  newRotation?: null | CreateRotationInput
  rules: ScheduleRuleInput[]
  scheduleID?: null | string
  target?: null | TargetInput
}

export interface SendContactMethodVerificationInput {
  contactMethodID: string
}

export interface Service {
  description: string
  escalationPolicy?: null | EscalationPolicy
  escalationPolicyID: string
  heartbeatMonitors: HeartbeatMonitor[]
  id: string
  integrationKeys: IntegrationKey[]
  isFavorite: boolean
  labels: Label[]
  maintenanceExpiresAt?: null | ISOTimestamp
  name: string
  notices: Notice[]
  onCallUsers: ServiceOnCallUser[]
}

export interface ServiceConnection {
  nodes: Service[]
  pageInfo: PageInfo
}

export interface ServiceOnCallUser {
  stepNumber: number
  userID: string
  userName: string
}

export interface ServiceSearchOptions {
  after?: null | string
  favoritesFirst?: null | boolean
  favoritesOnly?: null | boolean
  first?: null | number
  omit?: null | string[]
  search?: null | string
}

export interface SetAlertNoiseReasonInput {
  alertID: number
  noiseReason: string
}

export interface SetFavoriteInput {
  favorite: boolean
  target: TargetInput
}

export interface SetLabelInput {
  key: string
  target?: null | TargetInput
  value: string
}

export interface SetScheduleOnCallNotificationRulesInput {
  rules: OnCallNotificationRuleInput[]
  scheduleID: string
}

export interface SetScheduleShiftInput {
  end: ISOTimestamp
  start: ISOTimestamp
  userID: string
}

export interface SetTemporaryScheduleInput {
  clearEnd?: null | ISOTimestamp
  clearStart?: null | ISOTimestamp
  end: ISOTimestamp
  scheduleID: string
  shifts: SetScheduleShiftInput[]
  start: ISOTimestamp
}

export interface SlackChannel {
  id: string
  name: string
  teamID: string
}

export interface SlackChannelConnection {
  nodes: SlackChannel[]
  pageInfo: PageInfo
}

export interface SlackChannelSearchOptions {
  after?: null | string
  first?: null | number
  omit?: null | string[]
  search?: null | string
}

export interface SlackUserGroup {
  handle: string
  id: string
  name: string
}

export interface SlackUserGroupConnection {
  nodes: SlackUserGroup[]
  pageInfo: PageInfo
}

export interface SlackUserGroupSearchOptions {
  after?: null | string
  first?: null | number
  omit?: null | string[]
  search?: null | string
}

export type StatusUpdateState =
  | 'DISABLED'
  | 'DISABLED_FORCED'
  | 'ENABLED'
  | 'ENABLED_FORCED'

export type String = string

export interface StringConnection {
  nodes: string[]
  pageInfo: PageInfo
}

export interface SystemLimit {
  description: string
  id: SystemLimitID
  value: number
}

export type SystemLimitID =
  | 'CalendarSubscriptionsPerUser'
  | 'ContactMethodsPerUser'
  | 'EPActionsPerStep'
  | 'EPStepsPerPolicy'
  | 'HeartbeatMonitorsPerService'
  | 'IntegrationKeysPerService'
  | 'NotificationRulesPerUser'
  | 'ParticipantsPerRotation'
  | 'RulesPerSchedule'
  | 'TargetsPerSchedule'
  | 'UnackedAlertsPerService'
  | 'UserOverridesPerSchedule'

export interface SystemLimitInput {
  id: SystemLimitID
  value: number
}

export interface Target {
  id: string
  name: string
  type: TargetType
}

export interface TargetInput {
  id: string
  type: TargetType
}

export type TargetType =
  | 'calendarSubscription'
  | 'chanWebhook'
  | 'contactMethod'
  | 'escalationPolicy'
  | 'heartbeatMonitor'
  | 'integrationKey'
  | 'notificationChannel'
  | 'notificationPolicy'
  | 'notificationRule'
  | 'rotation'
  | 'schedule'
  | 'service'
  | 'slackChannel'
  | 'slackUserGroup'
  | 'user'
  | 'userOverride'
  | 'userSession'

export interface TemporarySchedule {
  end: ISOTimestamp
  shifts: OnCallShift[]
  start: ISOTimestamp
}

export interface TimeSeriesBucket {
  count: number
  end: ISOTimestamp
  start: ISOTimestamp
}

export interface TimeSeriesOptions {
  bucketDuration: ISODuration
  bucketOrigin?: null | ISOTimestamp
}

export interface TimeZone {
  id: string
}

export interface TimeZoneConnection {
  nodes: TimeZone[]
  pageInfo: PageInfo
}

export interface TimeZoneSearchOptions {
  after?: null | string
  first?: null | number
  omit?: null | string[]
  search?: null | string
}

export interface TokenInfo {
  primaryHint: string
  secondaryHint: string
}

export interface UpdateAlertsByServiceInput {
  newStatus: AlertStatus
  serviceID: string
}

export interface UpdateAlertsInput {
  alertIDs: number[]
  newStatus?: null | AlertStatus
  noiseReason?: null | string
}

export interface UpdateBasicAuthInput {
  oldPassword?: null | string
  password: string
  userID: string
}

export interface UpdateEscalationPolicyInput {
  description?: null | string
  id: string
  name?: null | string
  repeat?: null | number
  stepIDs?: null | string[]
}

export interface UpdateEscalationPolicyStepInput {
  actions?: null | DestinationInput[]
  delayMinutes?: null | number
  id: string
  targets?: null | TargetInput[]
}

export interface UpdateGQLAPIKeyInput {
  description?: null | string
  id: string
  name?: null | string
}

export interface UpdateHeartbeatMonitorInput {
  additionalDetails?: null | string
  id: string
  name?: null | string
  timeoutMinutes?: null | number
}

export interface UpdateKeyConfigInput {
  defaultActions?: null | ActionInput[]
  deleteRule?: null | string
  keyID: string
  rules?: null | KeyRuleInput[]
  setRule?: null | KeyRuleInput
}

export interface UpdateRotationInput {
  activeUserIndex?: null | number
  description?: null | string
  id: string
  name?: null | string
  shiftLength?: null | number
  start?: null | ISOTimestamp
  timeZone?: null | string
  type?: null | RotationType
  userIDs?: null | string[]
}

export interface UpdateScheduleInput {
  description?: null | string
  id: string
  name?: null | string
  timeZone?: null | string
}

export interface UpdateServiceInput {
  description?: null | string
  escalationPolicyID?: null | string
  id: string
  maintenanceExpiresAt?: null | ISOTimestamp
  name?: null | string
}

export interface UpdateUserCalendarSubscriptionInput {
  disabled?: null | boolean
  fullSchedule?: null | boolean
  id: string
  name?: null | string
  reminderMinutes?: null | number[]
}

export interface UpdateUserContactMethodInput {
  enableStatusUpdates?: null | boolean
  id: string
  name?: null | string
  value?: null | string
}

export interface UpdateUserInput {
  email?: null | string
  id: string
  name?: null | string
  role?: null | UserRole
  statusUpdateContactMethodID?: null | string
}

export interface UpdateUserOverrideInput {
  addUserID?: null | string
  end?: null | ISOTimestamp
  id: string
  removeUserID?: null | string
  start?: null | ISOTimestamp
}

export interface User {
  assignedSchedules: Schedule[]
  authSubjects: AuthSubject[]
  calendarSubscriptions: UserCalendarSubscription[]
  contactMethods: UserContactMethod[]
  email: string
  id: string
  isFavorite: boolean
  name: string
  notificationRules: UserNotificationRule[]
  onCallOverview: OnCallOverview
  onCallSteps: EscalationPolicyStep[]
  role: UserRole
  sessions: UserSession[]
  statusUpdateContactMethodID: string
}

export interface UserCalendarSubscription {
  disabled: boolean
  fullSchedule: boolean
  id: string
  lastAccess: ISOTimestamp
  name: string
  reminderMinutes: number[]
  schedule?: null | Schedule
  scheduleID: string
  url?: null | string
}

export interface UserConnection {
  nodes: User[]
  pageInfo: PageInfo
}

export interface UserContactMethod {
  dest: Destination
  disabled: boolean
  formattedValue: string
  id: string
  lastTestMessageState?: null | NotificationState
  lastTestVerifyAt?: null | ISOTimestamp
  lastVerifyMessageState?: null | NotificationState
  name: string
  pending: boolean
  statusUpdates: StatusUpdateState
  type?: null | ContactMethodType
  value: string
}

export interface UserNotificationRule {
  contactMethod?: null | UserContactMethod
  contactMethodID: string
  delayMinutes: number
  id: string
}

export interface UserOverride {
  addUser?: null | User
  addUserID: string
  end: ISOTimestamp
  id: string
  removeUser?: null | User
  removeUserID: string
  start: ISOTimestamp
  target: Target
}

export interface UserOverrideConnection {
  nodes: UserOverride[]
  pageInfo: PageInfo
}

export interface UserOverrideSearchOptions {
  after?: null | string
  end?: null | ISOTimestamp
  filterAddUserID?: null | string[]
  filterAnyUserID?: null | string[]
  filterRemoveUserID?: null | string[]
  first?: null | number
  omit?: null | string[]
  scheduleID?: null | string
  start?: null | ISOTimestamp
}

export type UserRole = 'admin' | 'unknown' | 'user'

export interface UserSearchOptions {
  CMType?: null | ContactMethodType
  CMValue?: null | string
  after?: null | string
  dest?: null | DestinationInput
  favoritesFirst?: null | boolean
  favoritesOnly?: null | boolean
  first?: null | number
  omit?: null | string[]
  search?: null | string
}

export interface UserSession {
  createdAt: ISOTimestamp
  current: boolean
  id: string
  lastAccessAt: ISOTimestamp
  userAgent: string
}

export interface VerifyContactMethodInput {
  code: number
  contactMethodID: string
}

export type WeekdayFilter = [
  boolean,
  boolean,
  boolean,
  boolean,
  boolean,
  boolean,
  boolean,
]

export interface __Directive {
  args: __InputValue[]
  description?: null | string
  isRepeatable: boolean
  locations: __DirectiveLocation[]
  name: string
}

export type __DirectiveLocation =
  | 'ARGUMENT_DEFINITION'
  | 'ENUM'
  | 'ENUM_VALUE'
  | 'FIELD'
  | 'FIELD_DEFINITION'
  | 'FRAGMENT_DEFINITION'
  | 'FRAGMENT_SPREAD'
  | 'INLINE_FRAGMENT'
  | 'INPUT_FIELD_DEFINITION'
  | 'INPUT_OBJECT'
  | 'INTERFACE'
  | 'MUTATION'
  | 'OBJECT'
  | 'QUERY'
  | 'SCALAR'
  | 'SCHEMA'
  | 'SUBSCRIPTION'
  | 'UNION'
  | 'VARIABLE_DEFINITION'

export interface __EnumValue {
  deprecationReason?: null | string
  description?: null | string
  isDeprecated: boolean
  name: string
}

export interface __Field {
  args: __InputValue[]
  deprecationReason?: null | string
  description?: null | string
  isDeprecated: boolean
  name: string
  type: __Type
}

export interface __InputValue {
  defaultValue?: null | string
  description?: null | string
  name: string
  type: __Type
}

export interface __Schema {
  description?: null | string
  directives: __Directive[]
  mutationType?: null | __Type
  queryType: __Type
  subscriptionType?: null | __Type
  types: __Type[]
}

export interface __Type {
  description?: null | string
  enumValues?: null | __EnumValue[]
  fields?: null | __Field[]
  inputFields?: null | __InputValue[]
  interfaces?: null | __Type[]
  kind: __TypeKind
  name?: null | string
  ofType?: null | __Type
  possibleTypes?: null | __Type[]
  specifiedByURL?: null | string
}

export type __TypeKind =
  | 'ENUM'
  | 'INPUT_OBJECT'
  | 'INTERFACE'
  | 'LIST'
  | 'NON_NULL'
  | 'OBJECT'
  | 'SCALAR'
  | 'UNION'

type ConfigID =
  | 'General.ApplicationName'
  | 'General.PublicURL'
  | 'General.GoogleAnalyticsID'
  | 'General.NotificationDisclaimer'
  | 'General.DisableMessageBundles'
  | 'General.ShortURL'
  | 'General.DisableSMSLinks'
  | 'General.DisableLabelCreation'
  | 'General.DisableCalendarSubscriptions'
  | 'Services.RequiredLabels'
  | 'Maintenance.AlertCleanupDays'
  | 'Maintenance.AlertAutoCloseDays'
  | 'Maintenance.AutoCloseAckedAlerts'
  | 'Maintenance.APIKeyExpireDays'
  | 'Maintenance.ScheduleCleanupDays'
  | 'Auth.RefererURLs'
  | 'Auth.DisableBasic'
  | 'GitHub.Enable'
  | 'GitHub.NewUsers'
  | 'GitHub.ClientID'
  | 'GitHub.ClientSecret'
  | 'GitHub.AllowedUsers'
  | 'GitHub.AllowedOrgs'
  | 'GitHub.EnterpriseURL'
  | 'OIDC.Enable'
  | 'OIDC.NewUsers'
  | 'OIDC.OverrideName'
  | 'OIDC.IssuerURL'
  | 'OIDC.ClientID'
  | 'OIDC.ClientSecret'
  | 'OIDC.Scopes'
  | 'OIDC.UserInfoEmailPath'
  | 'OIDC.UserInfoEmailVerifiedPath'
  | 'OIDC.UserInfoNamePath'
  | 'Mailgun.Enable'
  | 'Mailgun.APIKey'
  | 'Mailgun.EmailDomain'
  | 'Slack.Enable'
  | 'Slack.ClientID'
  | 'Slack.ClientSecret'
  | 'Slack.AccessToken'
  | 'Slack.SigningSecret'
  | 'Slack.InteractiveMessages'
  | 'Twilio.Enable'
  | 'Twilio.VoiceName'
  | 'Twilio.VoiceLanguage'
  | 'Twilio.AccountSID'
  | 'Twilio.AuthToken'
  | 'Twilio.AlternateAuthToken'
  | 'Twilio.FromNumber'
  | 'Twilio.MessagingServiceSID'
  | 'Twilio.DisableTwoWaySMS'
  | 'Twilio.SMSCarrierLookup'
  | 'Twilio.SMSFromNumberOverride'
  | 'SMTP.Enable'
  | 'SMTP.From'
  | 'SMTP.Address'
  | 'SMTP.DisableTLS'
  | 'SMTP.SkipVerify'
  | 'SMTP.Username'
  | 'SMTP.Password'
  | 'Webhook.Enable'
  | 'Webhook.AllowedURLs'
  | 'Feedback.Enable'
  | 'Feedback.OverrideURL'
