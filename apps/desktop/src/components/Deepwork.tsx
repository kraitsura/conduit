import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { useTaskBreakdown } from "@/hooks/useTaskBreakdown";
import { Clock, Pause } from "lucide-react";
import { SubTask } from "@/types";
import TimerStack from "./TimerStack";
import { SessionReflection } from "./SessionReflection";
import { useStore } from "../store/useStore";
import { TaskCardModal } from "./TaskCardModal";
import { formatTime, getEnergyColor } from "@/utils/formatters";
import { motion, AnimatePresence } from "framer-motion";

const DeepWorkAssistant = () => {
  const [task, setTask] = useState("");
  const [selectedTasks, setSelectedTasks] = useState<SubTask[]>([]);
  const [isTaskModalOpen, setIsTaskModalOpen] = useState(false);
  const [activeSubtask, setActiveSubtask] = useState<SubTask | null>(null);
  const { loading, error, subtasks, getSubtasks } = useTaskBreakdown();

  const {
    startSession,
    endSession,
    addSubTasks,
    completeSubTask,
    isTimerActive,
    currentSession,
  } = useStore();

  const handleBreakdown = async () => {
    if (task.trim()) {
      await getSubtasks(task);
      setIsTaskModalOpen(true);
    }
  };

  const handleTaskSelection = (subtask: SubTask) => {
    setSelectedTasks((prev) =>
      prev.some((t) => t.id === subtask.id)
        ? prev.filter((t) => t.id !== subtask.id)
        : [...prev, subtask]
    );
  };

  const handleSubtaskClick = (subtask: SubTask) => {
    if (!subtask.completed) {
      setActiveSubtask(subtask);
    }
  };

  const handleGeneratePlan = () => {
    startSession(task);
    addSubTasks(selectedTasks);
    setSelectedTasks([]);
    setTask("");
    setIsTaskModalOpen(false);
  };

  return (
    <div className="w-full">
      <div className="mx-auto">
        <AnimatePresence>
          {!currentSession && (
            <motion.div
              initial={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: 300 }}
              className="w-full"
            >
              <Card>
                <CardHeader>
                  <CardTitle>Task Breakdown</CardTitle>
                  <CardDescription>
                    Enter your task and click 'Breakdown' to get started
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="flex space-x-2">
                    <Input
                      placeholder="Enter your task here..."
                      value={task}
                      onChange={(e) => setTask(e.target.value)}
                    />
                    <Button 
                      onClick={handleBreakdown} 
                      disabled={loading || !task.trim()}
                      className="bg-primary hover:bg-primary/90 text-primary-foreground"
                    >
                      {loading ? "Breaking down..." : "Breakdown"}
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </motion.div>
          )}

          {currentSession && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="grid grid-cols-1 md:grid-cols-2 gap-6"
            >
              {/* Timer Stack Column */}
              <div>
                <TimerStack />
              </div>

              {/* Tasks Column */}
              <motion.div
                initial={{ x: 0 }}
                animate={{ x: 0 }}
                className="space-y-6"
              >
                <Card>
                  <CardHeader>
                    <CardTitle>Current Session: {currentSession.task.description}</CardTitle>
                    <CardDescription>Track your progress for this session</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <ul className="space-y-4">
                      {currentSession.subTasks.map((subtask) => (
                        <li
                          key={subtask.id}
                          className={`flex items-center space-x-3 p-2 border rounded cursor-pointer transition-colors ${
                            activeSubtask?.id === subtask.id ? 'bg-muted' : ''
                          }`}
                          onClick={() => handleSubtaskClick(subtask)}
                        >
                          <Checkbox
                            checked={subtask.completed}
                            onCheckedChange={() => completeSubTask(subtask.id)}
                            id={`session-task-${subtask.id}`}
                          />
                          <div>
                            <label
                              htmlFor={`session-task-${subtask.id}`}
                              className="font-medium"
                            >
                              {subtask.description}
                            </label>
                            <div className="flex gap-2 mt-1">
                              <Badge
                                variant="secondary"
                                className="flex items-center gap-1"
                              >
                                <Clock className="w-3 h-3" />
                                {formatTime(subtask.time_estimate_minutes)}
                              </Badge>
                              <Badge
                                variant="secondary"
                                className={`${getEnergyColor(subtask.mental_energy)}`}
                              >
                                {subtask.mental_energy} Energy
                              </Badge>
                            </div>
                          </div>
                        </li>
                      ))}
                    </ul>
                  </CardContent>
                </Card>

                <Button
                  onClick={endSession}
                  className="bg-primary hover:bg-primary/90 text-primary-foreground flex items-center space-x-2"
                >
                  <Pause className="w-4 h-4" />
                  <span>End Session</span>
                </Button>
              </motion.div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      <TaskCardModal
        isOpen={isTaskModalOpen}
        onClose={() => setIsTaskModalOpen(false)}
        subtasks={subtasks}
        selectedTasks={selectedTasks}
        onSelect={handleTaskSelection}
        onConfirm={handleGeneratePlan}
      />

      <SessionReflection />
    </div>
  );
};

export default DeepWorkAssistant;