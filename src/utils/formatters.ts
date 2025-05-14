// utils/formatters.ts
import { MentalEnergy } from '@/types'

export const getEnergyColor = (energy: MentalEnergy) => {
  switch (energy) {
    case MentalEnergy.Low: 
      return {
        badge: 'bg-secondary text-secondary-foreground',
        card: 'from-secondary/70 to-secondary'
      }
    case MentalEnergy.Medium: 
      return {
        badge: 'bg-secondary/80 text-secondary-foreground',
        card: 'from-secondary/30 to-primary/50'
      }
    case MentalEnergy.High: 
      return {
        badge: 'bg-primary text-primary-foreground',
        card: 'from-primary/60 to-primary'
      }
    default: 
      return {
        badge: 'bg-muted text-muted-foreground',
        card: 'from-muted/70 to-muted'
      }
  }
}

export const formatTime = (minutes: number) => {
  const hours = Math.floor(minutes / 60)
  const remainingMinutes = minutes % 60
  return hours > 0
    ? `${hours}h ${remainingMinutes}m`
    : `${remainingMinutes}m`
}