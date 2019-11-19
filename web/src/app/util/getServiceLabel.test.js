import getServiceLabel from './getServiceLabel'

test('it should split service labels correctly', () => {
  expect(getServiceLabel('wcbn.fm/rfaa=88.3 ann arbor')).toStrictEqual({
    labelKey: 'wcbn.fm/rfaa',
    labelValue: '88.3 ann arbor',
  })
})
