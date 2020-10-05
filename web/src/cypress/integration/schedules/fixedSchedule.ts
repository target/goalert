import { Chance } from 'chance'
import { testScreen } from '../../support'
import { Schedule } from '../../../schema'

const c = new Chance()

function testFixedSchedule(screen: ScreenFormat): void {
  it('should create a fixed schedule', () => {})
  it('should create a fixed schedule overlapping existing shifts', () => {})
  it('should edit a fixed schedule', () => {})
  it('should delete a fixed schedule', () => {})

  it('should toggle timezone', () => {})
  it('should toggle duration field', () => {})
  it('should delete a shift in step 2', () => {})
  it('should go back and forth between steps', () => {})
  it('should cancel and close form', () => {})
}

testScreen('Fixed Schedule', testFixedSchedule)
