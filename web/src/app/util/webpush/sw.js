// // Register event listener for the 'push' event.
self.addEventListener('push', (event) => {
  const decoder = new TextDecoder()
  const buff = event.data.arrayBuffer()
  const messageData = JSON.parse(decoder.decode(buff))

  console.log(messageData)

  // Keep the service worker alive until the notification is created.
  event.waitUntil(
    self.registration.showNotification('GoAlert', {
      badge:
        'https://raw.githubusercontent.com/target/goalert/master/web/src/app/public/favicon-64.png',
      body: messageData.message,
      data: messageData,
      icon:
        'https://raw.githubusercontent.com/target/goalert/master/web/src/app/public/favicon.ico',
      lang: 'en',
      requireInteraction: true,
      silent: false,
      actions: [
        { action: 'ack', title: 'Acknowledge' },
        { action: 'close', title: 'Close' },
      ],
    }),
  )
})

self.addEventListener(
  'notificationclick',
  (event) => {
    const data = event.notification.data

    event.notification.close()

    if (event.action === 'ack') {
      console.log(data.ackCode)
    } else if (event.action === 'close') {
      console.log(data.closeCode)
    } else {
      // clients.openWindow(data.url)
    }
  },
  false,
)
