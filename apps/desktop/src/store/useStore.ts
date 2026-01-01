import { create } from "zustand";
import { SubTask, Task, Session, Distraction, Reflection, MentalEnergy } from "../types";

interface DeepWorkStore {
  // State
  currentSession: Session | null;
  sessions: Session[];
  isTimerActive: boolean;
  remainingTime: number;
  
  // Actions
  startSession: (mainTask: string) => void;
  endSession: () => void;
  addSubTasks: (subtasks: SubTask[]) => void;
  completeSubTask: (subtaskId: string) => void;
  logDistraction: (distraction: Omit<Distraction, "id">) => void;
  setReflection: (content: string) => void;
}

export const useStore = create<DeepWorkStore>((set) => ({
  currentSession: null,
  sessions: [],
  isTimerActive: false,
  remainingTime: 25 * 60, // 25 minutes in seconds

  startSession: (mainTask: string) => 
    set((state) => {
      const newSession: Session = {
        id: Date.now().toString(),
        startTime: new Date(),
        task: {
          id: Date.now().toString(),
          description: mainTask,
          subtasks: [],
          completed: false,
          date: new Date()
        },
        subTasks: [],
        distractions: [],
        reflections: []
      };

      return {
        currentSession: newSession,
        sessions: [...state.sessions, newSession],
        isTimerActive: true
      };
    }),

  endSession: () =>
    set((state) => {
      if (!state.currentSession) return state;

      const updatedSession = {
        ...state.currentSession,
        endTime: new Date()
      };

      return {
        currentSession: null,
        sessions: state.sessions.map(session =>
          session.id === updatedSession.id ? updatedSession : session
        ),
        isTimerActive: false
      };
    }),

  addSubTasks: (subtasks: SubTask[]) =>
    set((state) => {
      if (!state.currentSession) return state;

      const updatedSession = {
        ...state.currentSession,
        subTasks: [...state.currentSession.subTasks, ...subtasks],
        task: {
          ...state.currentSession.task,
          subtasks: [...state.currentSession.task.subtasks, ...subtasks]
        }
      };

      return {
        currentSession: updatedSession,
        sessions: state.sessions.map(session =>
          session.id === updatedSession.id ? updatedSession : session
        )
      };
    }),

  completeSubTask: (subtaskId: string) =>
    set((state) => {
      if (!state.currentSession) return state;

      const updatedSubTasks = state.currentSession.subTasks.map(subtask =>
        subtask.id === subtaskId ? { ...subtask, completed: true } : subtask
      );

      const allSubtasksCompleted = updatedSubTasks.every(subtask => subtask.completed);

      const updatedSession = {
        ...state.currentSession,
        subTasks: updatedSubTasks,
        task: {
          ...state.currentSession.task,
          subtasks: updatedSubTasks,
          completed: allSubtasksCompleted
        }
      };

      return {
        currentSession: updatedSession,
        sessions: state.sessions.map(session =>
          session.id === updatedSession.id ? updatedSession : session
        )
      };
    }),

  logDistraction: (distraction: Omit<Distraction, "id">) =>
    set((state) => {
      if (!state.currentSession) return state;

      const updatedSession = {
        ...state.currentSession,
        distractions: [
          ...(state.currentSession.distractions || []),
          { ...distraction, id: Date.now().toString() }
        ]
      };

      return {
        currentSession: updatedSession,
        sessions: state.sessions.map(session =>
          session.id === updatedSession.id ? updatedSession : session
        )
      };
    }),

  setReflection: (content: string) =>
    set((state) => {
      if (!state.currentSession) return state;

      const newReflection: Reflection = {
        id: Date.now().toString(),
        timestamp: new Date(),
        content
      };

      const updatedSession = {
        ...state.currentSession,
        reflections: [
          ...(state.currentSession.reflections || []),
          newReflection
        ]
      };

      return {
        currentSession: updatedSession,
        sessions: state.sessions.map(session =>
          session.id === updatedSession.id ? updatedSession : session
        )
      };
    })
}));