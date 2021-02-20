export function isSafari(): boolean {
  return 'safari' in window
}

export function isNotificationSupported(): boolean {
  return 'serviceWorker' in navigator && 'Notification' in window
}

/*
 * asks user consent to receive push notifications and returns the response of the user, one of granted, default, denied
 */
export function askNotificationPermission(
  callback: NotificationPermissionCallback,
): void {
  try {
    // Safari doesn't return a promise for requestPermissions and it
    // throws a TypeError. It takes a callback as the first argument
    // instead.
    Notification.requestPermission((perm) => {
      callback(perm)
    }).catch((err) => console.log(err))
  } catch (error) {
    // Firefox, Chrome etc.
    Notification.requestPermission()
      .then((perm) => callback(perm))
      .catch((err) => console.log(err))
  }
}

/*
 * checks if at least one service worker is present
 */
export async function isAnyServiceWorkerRegistered(): Promise<boolean> {
  return navigator.serviceWorker.getRegistrations().then((registrations) => {
    return registrations.length > 0
  })
}

export async function isServiceWorkerRegistered(
  clientURL = '/',
): Promise<boolean | void> {
  return navigator.serviceWorker
    .getRegistration(clientURL)
    .then((r) => r !== undefined)
    .catch((err) =>
      console.log(`Error getting SW registration at "${clientURL}":`, err),
    )
}

/*
 * shows a notification
 */
// TODO refactor
export function sendNotification(): void {
  const img = '/images/jason-leung-HM6TMmevbZQ-unsplash.jpg'
  const text = 'Take a look at this brand new t-shirt!'
  const title = 'New Product Available'
  const options = {
    body: text,
    icon: '/images/jason-leung-HM6TMmevbZQ-unsplash.jpg',
    vibrate: [200, 100, 200],
    tag: 'new-product',
    image: img,
    badge: 'https://spyna.it/icons/android-icon-192x192.png',
    actions: [
      {
        action: 'Detail',
        title: 'View',
        icon: 'https://via.placeholder.com/128/ff0000',
      },
    ],
  }
  navigator.serviceWorker.ready.then(function (serviceWorker) {
    serviceWorker.showNotification(title, options)
  })
}

/*
 * using the registered service worker creates a push notification subscription and returns it
 */
export async function createNotificationSubscription(
  registration: ServiceWorkerRegistration,
  vapidPublicKey: PushSubscriptionOptionsInit['applicationServerKey'],
): Promise<PushSubscription> {
  console.log('subscribing with ', vapidPublicKey)
  // const registration = await navigator.serviceWorker.ready

  // fetch VAPID key
  // const response = await fetch('/getVapidPublicKey') // TODO replace fetch route with this
  // const response = await fetch(
  //   'https://blooming-ocean-51906.herokuapp.com/getVapidPublicKey',
  // )
  // const json = await response.json()
  // const vapidPublicKey = json.value

  return registration.pushManager.subscribe({
    userVisibleOnly: true,
    applicationServerKey: vapidPublicKey,
  })
}

/*
 * returns the subscription if present or nothing
 */
export async function getUserSubscription(): Promise<PushSubscription | null> {
  const registration = await navigator.serviceWorker.ready
  return registration.pushManager.getSubscription()
}

// TODO
// export function registerForSafariRemoteNotifications(): void {
//   const websitePushID = 'web.com.target.GoAlert' // TODO config

//   const permissionData = window.safari.pushNotification.permission(
//     websitePushID,
//   )

//   const checkRemotePermission = (permissionData) => {
//     console.log(permissionData)
//     if (permissionData.permission === 'default') {
//       console.log('default, requesting permission...')
//       // This is a new web service URL and its validity is unknown.
//       window.safari.pushNotification.requestPermission(
//         'https://blooming-ocean-51906.herokuapp.com/push', // The web service URL.
//         websitePushID, // The Website Push ID.
//         {}, // Data that you choose to send to your server to help you identify the user.
//         checkRemotePermission, // The callback function.
//       )
//     } else if (permissionData.permission === 'denied') {
//       // The user said no.
//     } else if (permissionData.permission === 'granted') {
//       // The web service URL is a valid push provider, and the user said yes.
//       // permissionData.deviceToken is now available to use.
//     }
//   }
//   checkRemotePermission(permissionData)
// }

export async function registerSW(): Promise<ServiceWorkerRegistration> {
  // clean slate?
  // await navigator.serviceWorker.getRegistrations().then((registrations) => {
  //   for (const registration of registrations) {
  //     registration.unregister()
  //   }
  // })

  let registration = await navigator.serviceWorker.register('/static/sw.js')

  // TODO fix race condition here b/w state of installing vs activated
  while (registration.waiting || registration.installing) {
    // wait
    console.log('waiting/installing')
    registration = await navigator.serviceWorker.register('/static/sw.js')
  }

  return registration
}

export async function unregisterSW(): Promise<void> {
  const registration = await navigator.serviceWorker.ready
  registration.unregister()
}
