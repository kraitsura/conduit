import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Checkbox } from "@/components/ui/checkbox"
import { useTaskBreakdown } from "@/hooks/useTaskBreakdown"
import { Clock, Wrench, Link2, ArrowRight } from "lucide-react"
import { MentalEnergy, SubTask } from '@/types'

function TaskBreak() {
  const [task, setTask] = useState("")
  const [selectedTasks, setSelectedTasks] = useState<SubTask[]>([])
  const { loading, error, subtasks, getSubtasks } = useTaskBreakdown()

  const handleBreakdown = async () => {
    if (task.trim()) {
      await getSubtasks(task)
    }
  }

  const handleTaskSelection = (subtask: SubTask) => {
    setSelectedTasks((prev) =>
      prev.some((t) => t.id === subtask.id)
        ? prev.filter((t) => t.id !== subtask.id)
        : [...prev, subtask]
    )
  }

  const formatTime = (minutes: number) => {
    const hours = Math.floor(minutes / 60)
    const remainingMinutes = minutes % 60
    return hours > 0
      ? `${hours}h ${remainingMinutes}m`
      : `${remainingMinutes}m`
  }

  const getEnergyColor = (energy: MentalEnergy) => {
    switch (energy) {
      case MentalEnergy.Low: return 'bg-green-600'
      case MentalEnergy.Medium: return 'bg-yellow-600'
      case MentalEnergy.High: return 'bg-red-600'
      default: return 'bg-gray-600'
    }
  }

  const handleGeneratePlan = () => {
    console.log("Generating plan for:", selectedTasks)
    // Implement plan generation logic here
  }

  return (
    <div className="min-h-screen bg-gray-900 text-gray-100 p-4">
      <div className="max-w-4xl mx-auto space-y-6">
        <Card className="bg-gray-800 border-gray-700">
          <CardHeader>
            <CardTitle>Task Breakdown</CardTitle>
            <CardDescription>Enter your task and click 'Breakdown' to get started</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex space-x-2">
              <Input
                placeholder="Enter your task here..."
                value={task}
                onChange={(e) => setTask(e.target.value)}
                className="bg-gray-700 border-gray-600 text-gray-100"
              />
              <Button onClick={handleBreakdown} disabled={loading || !task.trim()}>
                {loading ? "Breaking down..." : "Breakdown"}
              </Button>
            </div>
          </CardContent>
        </Card>

        {error && (
          <Card className="bg-red-900 border-red-700">
            <CardContent>
              <p className="text-red-100">{error}</p>
            </CardContent>
          </Card>
        )}

        {subtasks.length > 0 && (
          <Card className="bg-gray-800 border-gray-700">
            <CardHeader>
              <CardTitle>Subtasks</CardTitle>
              <CardDescription>Select tasks to include in your plan</CardDescription>
            </CardHeader>
            <CardContent>
              <ul className="space-y-6">
                {subtasks.map((subtask) => (
                  <li key={subtask.id} className="border border-gray-700 rounded-lg p-4">
                    <div className="flex items-start space-x-3">
                      <Checkbox
                        id={`task-${subtask.id}`}
                        checked={selectedTasks.some((t) => t.id === subtask.id)}
                        onCheckedChange={() => handleTaskSelection(subtask)}
                      />
                      <div className="flex-1 space-y-3">
                        <div>
                          <label htmlFor={`task-${subtask.id}`} className="font-medium cursor-pointer">
                            {subtask.description}
                          </label>
                          <div className="flex flex-wrap gap-2 mt-2">
                            <Badge variant="secondary" className="flex items-center gap-1">
                              <Clock className="w-3 h-3" />
                              {formatTime(subtask.time_estimate_minutes)}
                            </Badge>
                            <Badge
                              variant="secondary"
                              className={`${getEnergyColor(subtask.mental_energy)}`}
                            >
                              {subtask.mental_energy} Energy
                            </Badge>
                            <Badge variant="outline">{subtask.task_type.replace('_', ' ')}</Badge>
                            {subtask.dependencies.critical_path && (
                              <Badge variant="destructive">Critical Path</Badge>
                            )}
                          </div>
                        </div>

                        {subtask.dependencies.prerequisite_tasks.length > 0 && (
                          <div className="flex items-center gap-1 text-gray-400">
                            <ArrowRight className="w-4 h-4" />
                            Requires: {subtask.dependencies.prerequisite_tasks.map(String).join(", ")}
                          </div>
                        )}
                        {subtask.dependencies.unlocks.length > 0 && (
                          <div className="flex items-center gap-1 text-gray-400">
                            <ArrowRight className="w-4 h-4" />
                            Unlocks: {subtask.dependencies.unlocks.map(String).join(", ")}
                          </div>
                        )}


                        <div className="flex flex-wrap gap-4 text-sm">
                          {subtask.resources.tools.length > 0 && (
                            <div className="flex items-center gap-1">
                              <Wrench className="w-4 h-4 text-gray-400" />
                              <span className="text-gray-400">
                                Tools: {subtask.resources.tools.join(", ")}
                              </span>
                            </div>
                          )}
                          {subtask.resources.references.length > 0 && (
                            <div className="flex items-center gap-1">
                              <Link2 className="w-4 h-4 text-gray-400" />
                              <span className="text-gray-400">
                                References: {subtask.resources.references.join(", ")}
                              </span>
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  </li>
                ))}
              </ul>
            </CardContent>
            <CardFooter>
              <Button onClick={handleGeneratePlan} disabled={selectedTasks.length === 0}>
                Generate Plan
              </Button>
            </CardFooter>
          </Card>
        )}
      </div>
    </div>
  )
}

export default TaskBreak