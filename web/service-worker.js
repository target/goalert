self.addEventListener('install', () => {
  try {
    console.log('[sw] install')
  } catch (e) {}
  self.skipWaiting();
});

self.addEventListener('activate', (event) => {
  try {
    console.log('[sw] activate')
  } catch (e) {}
  event.waitUntil(self.clients.claim());
});

self.addEventListener('push', (event) => {
  let data = {}
  try {
    data = event.data ? event.data.json() : {}
    console.log('[sw] push received', { hasData: !!event.data, title: data && data.title, url: data && data.url })
  } catch (e) {
    try { console.error('[sw] push data parse error', e) } catch (_) {}
  }

  const title = (data && data.title) || 'GoAlert';
  const options = {
    body: data && data.body,
    data: data && data.url,
    icon: '/static/favicon-192.png',
    badge: '/static/favicon-128.png',
  };
  event.waitUntil(self.registration.showNotification(title, options));
});

self.addEventListener('notificationclick', (event) => {
  try { console.log('[sw] notificationclick') } catch (e) {}
  event.notification.close();
  const url = event.notification.data || '/';
  event.waitUntil(clients.openWindow(url));
});
