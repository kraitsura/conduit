import React from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import RPGTaskCards from './rpg-task-cards'
import TaskDetails from './task-details'  // You'll need to create this component for the right side
import { SubTask } from '@/types'

interface TaskCardModalProps {
  isOpen: boolean
  onClose: () => void
  subtasks: SubTask[]
  selectedTasks: SubTask[]
  onSelect: (task: SubTask) => void
  onConfirm: () => void
}

export const TaskCardModal: React.FC<TaskCardModalProps> = ({
  isOpen,
  onClose,
  subtasks,
  selectedTasks,
  onSelect,
  onConfirm
}) => {
  const [hoveredTask, setHoveredTask] = React.useState<SubTask | null>(null);

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[900px]">
        <DialogHeader>
          <DialogTitle>Select Your Tasks</DialogTitle>
        </DialogHeader>
        
        <div className="grid grid-cols-2 gap-6 mt-4">
          {/* Left side - RPG Task Cards */}
          <RPGTaskCards 
            subtasks={subtasks}
            selectedTasks={selectedTasks}
            onSelect={(task) => {
              onSelect(task);
              setHoveredTask(task);
            }}
          />

          {/* Right side - Task Details */}
          <TaskDetails hoveredTask={hoveredTask} />
        </div>

        <div className="flex justify-between mt-6 border-t pt-4">
          <div className="text-muted-foreground">
            Selected: {selectedTasks.length} tasks
          </div>
          <div className="space-x-2">
            <Button variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button 
              onClick={onConfirm} 
              disabled={selectedTasks.length === 0}
              className="bg-primary hover:bg-primary/90 text-primary-foreground"
            >
              Start Session
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}