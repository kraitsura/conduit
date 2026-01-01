import React from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Progress } from "@/components/ui/progress"
import { Separator } from "@/components/ui/separator"
import { Clock, Wrench, Link2, ArrowRight, Brain, Target } from 'lucide-react'
import { MentalEnergy, SubTask } from '@/types'

interface TaskDetailsProps {
  hoveredTask: SubTask | null
}

const getEnergyColor = (energy: MentalEnergy) => {
  switch (energy) {
    case MentalEnergy.Low: return 'bg-secondary'
    case MentalEnergy.Medium: return 'bg-secondary/70'
    case MentalEnergy.High: return 'bg-primary'
    default: return 'bg-muted'
  }
}

const formatTime = (minutes: number) => {
  const hours = Math.floor(minutes / 60)
  const remainingMinutes = minutes % 60
  return hours > 0
    ? `${hours}h ${remainingMinutes}m`
    : `${remainingMinutes}m`
}

export default function TaskDetails({ hoveredTask }: TaskDetailsProps) {
  return (
    <AnimatePresence mode="wait">
      {hoveredTask ? (
        <motion.div
          key={hoveredTask.id}
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -20 }}
          transition={{ duration: 0.3 }}
        >
          <Card className="overflow-hidden">
            <CardHeader className="bg-muted pb-4">
              <CardTitle className="text-2xl font-bold">{hoveredTask.description}</CardTitle>
            </CardHeader>
            <CardContent className="pt-6 space-y-6">
              <div className="flex items-center justify-between">
                <Badge variant="outline" className="text-lg px-3 py-1">
                  {hoveredTask.task_type.replace('_', ' ')}
                </Badge>
                <div className="flex items-center space-x-2">
                  <Brain className="w-5 h-5 text-muted-foreground" />
                  <span className="text-lg font-semibold">{hoveredTask.mental_energy} Energy</span>
                </div>
              </div>

              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Energy Level</span>
                  <span className="text-sm text-muted-foreground">{hoveredTask.mental_energy}</span>
                </div>
                <Progress 
                  value={
                    hoveredTask.mental_energy === MentalEnergy.Low ? 33 :
                    hoveredTask.mental_energy === MentalEnergy.Medium ? 66 : 100
                  } 
                  className={`h-2 ${getEnergyColor(hoveredTask.mental_energy)}`}
                />
              </div>

              <div className="flex items-center space-x-2">
                <Clock className="w-5 h-5 text-muted-foreground" />
                <span className="text-lg">Estimated time: {formatTime(hoveredTask.time_estimate_minutes)}</span>
              </div>

              {hoveredTask.dependencies.critical_path && (
                <div className="flex items-center space-x-2">
                  <Target className="w-5 h-5 text-primary" />
                  <span className="text-lg text-primary font-semibold">Critical Path Task</span>
                </div>
              )}

              <Separator />

              {(hoveredTask.dependencies.prerequisite_tasks.length > 0 || 
                hoveredTask.dependencies.unlocks.length > 0) && (
                <div className="space-y-2">
                  <h4 className="font-semibold text-lg">Dependencies</h4>
                  {hoveredTask.dependencies.prerequisite_tasks.length > 0 && (
                    <div className="flex items-start space-x-2">
                      <ArrowRight className="w-5 h-5 text-muted-foreground mt-1 flex-shrink-0" />
                      <span>
                        Requires: {hoveredTask.dependencies.prerequisite_tasks.join(", ")}
                      </span>
                    </div>
                  )}
                  {hoveredTask.dependencies.unlocks.length > 0 && (
                    <div className="flex items-start space-x-2">
                      <ArrowRight className="w-5 h-5 text-muted-foreground mt-1 flex-shrink-0" />
                      <span>
                        Unlocks: {hoveredTask.dependencies.unlocks.join(", ")}
                      </span>
                    </div>
                  )}
                </div>
              )}

              {(hoveredTask.resources.tools.length > 0 || 
                hoveredTask.resources.references.length > 0) && (
                <div className="space-y-2">
                  <h4 className="font-semibold text-lg">Resources</h4>
                  {hoveredTask.resources.tools.length > 0 && (
                    <div className="flex items-start space-x-2">
                      <Wrench className="w-5 h-5 text-muted-foreground mt-1 flex-shrink-0" />
                      <span>
                        Tools: {hoveredTask.resources.tools.join(", ")}
                      </span>
                    </div>
                  )}
                  {hoveredTask.resources.references.length > 0 && (
                    <div className="flex items-start space-x-2">
                      <Link2 className="w-5 h-5 text-muted-foreground mt-1 flex-shrink-0" />
                      <span>
                        References: {hoveredTask.resources.references.join(", ")}
                      </span>
                    </div>
                  )}
                </div>
              )}
            </CardContent>
          </Card>
        </motion.div>
      ) : (
        <motion.div
          key="empty"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          className="h-full flex items-center justify-center"
        >
          <p className="text-muted-foreground text-lg">Hover over a task card to see details</p>
        </motion.div>
      )}
    </AnimatePresence>
  )
}