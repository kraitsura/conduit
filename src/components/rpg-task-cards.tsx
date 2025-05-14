"use client"

import React, { useState } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Clock, Check } from "lucide-react"
import { MentalEnergy, SubTask } from '@/types'

interface RPGTaskCardsProps {
  subtasks: SubTask[]
  selectedTasks: SubTask[]
  onSelect: (task: SubTask) => void
}

const getEnergyColor = (energy: MentalEnergy) => {
  switch (energy) {
    case MentalEnergy.Low: return 'from-secondary/50 to-secondary'
    case MentalEnergy.Medium: return 'from-secondary-foreground/30 to-primary/50'
    case MentalEnergy.High: return 'from-primary/50 to-primary'
    default: return 'from-muted to-muted-foreground'
  }
}

const formatTime = (minutes: number) => {
  const hours = Math.floor(minutes / 60)
  const remainingMinutes = minutes % 60
  return hours > 0
    ? `${hours}h ${remainingMinutes}m`
    : `${remainingMinutes}m`
}

export default function RPGTaskCards({ subtasks, selectedTasks, onSelect }: RPGTaskCardsProps) {
  const [hoveredTask, setHoveredTask] = useState<SubTask | null>(null)

  return (
    <div className="h-full w-full overflow-y-auto p-4">
      <div className="grid grid-cols-2 gap-4 auto-rows-max">
        {subtasks.map((task) => {
          const isSelected = selectedTasks.some(t => t.id === task.id)

          return (
            <Card 
              key={task.id}
              className={`
                w-full cursor-pointer overflow-hidden
                transition-all duration-300 ease-in-out
                ${isSelected ? 'ring-4 ring-primary shadow-lg shadow-primary/30' : ''}
                ${hoveredTask?.id === task.id ? 'shadow-xl shadow-primary/20' : ''}
              `}
              onClick={() => {
                onSelect(task);
                setHoveredTask(task);
              }}
              onMouseEnter={() => setHoveredTask(task)}
              onMouseLeave={() => setHoveredTask(null)}
            >
              <CardContent className="p-4 flex flex-col justify-between relative overflow-hidden">
                <div 
                  className={`
                    absolute inset-0 bg-gradient-to-br ${getEnergyColor(task.mental_energy)} opacity-75
                    transition-opacity duration-300 ease-in-out
                    ${hoveredTask?.id === task.id ? 'opacity-100' : ''}
                  `}
                />
                <div className="relative z-10">
                  <div className="text-sm font-semibold mb-2 text-foreground">
                    {task.task_type.replace('_', ' ')}
                  </div>
                  <h3 className="text-lg font-bold mb-2 line-clamp-3 text-foreground">
                    {task.description}
                  </h3>
                </div>
                <div className="relative z-10 space-y-2">
                  <Badge variant="secondary" className="flex items-center gap-1 w-fit bg-background/60 text-foreground">
                    <Clock className="w-3 h-3" />
                    {formatTime(task.time_estimate_minutes)}
                  </Badge>
                  {task.dependencies.critical_path && (
                    <Badge variant="destructive" className="w-fit">
                      Critical Path
                    </Badge>
                  )}
                </div>
                {isSelected && (
                  <div className="absolute top-2 right-2 z-10">
                    <Check className="w-5 h-5 text-primary" />
                  </div>
                )}
              </CardContent>
            </Card>
          )
        })}
      </div>
    </div>
  )
}