import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Circle, Play, Pause, SkipForward } from 'lucide-react';
import { Button } from "@/components/ui/button";
import { useStore } from '../store/useStore';

const TIMER_TYPES = {
  FOCUS: {
    duration: 25 * 60, // 25 minutes
    label: 'Focus',
  },
  SHORT_BREAK: {
    duration: 5 * 60, // 5 minutes
    label: 'Short Break',
  },
  LONG_BREAK: {
    duration: 15 * 60, // 15 minutes
    label: 'Long Break',
  },
} as const;

type TimerType = keyof typeof TIMER_TYPES;

const TimerStack = () => {
  const [activeTimers, setActiveTimers] = useState<{ id: string; type: TimerType; completed: boolean }[]>([]);
  const [currentTimer, setCurrentTimer] = useState<{ id: string; type: TimerType; completed: boolean } | null>(null);
  const [isPaused, setIsPaused] = useState(true);
  const [timeRemaining, setTimeRemaining] = useState(0);
  const store = useStore();

  const calculateRequiredTimers = (minutes: any) => {
    const focusSessionsNeeded = Math.ceil(minutes / 25);
    const timers = [];
    
    for (let i = 0; i < focusSessionsNeeded; i++) {
      timers.push({
        id: `timer-${i}`,
        type: 'FOCUS',
        completed: false,
      });
      
      // Add breaks between focus sessions
      if (i < focusSessionsNeeded - 1) {
        timers.push({
          id: `break-${i}`,
          type: i % 2 === 0 ? 'SHORT_BREAK' : 'LONG_BREAK',
          completed: false,
        });
      }
    }
    
    return timers;
  };

  useEffect(() => {
	let interval: NodeJS.Timeout;
    
    if (!isPaused && timeRemaining > 0) {
      interval = setInterval(() => {
        setTimeRemaining((prev) => {
          if (prev <= 1) {
            handleTimerComplete();
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
    }

    return () => clearInterval(interval);
  }, [isPaused, timeRemaining]);

  const handleTimerComplete = () => {
    if (currentTimer) {
      const currentIndex = activeTimers.findIndex(t => t.id === currentTimer.id);
      const updatedTimers = [...activeTimers];
      updatedTimers[currentIndex].completed = true;
      
      if (currentIndex < updatedTimers.length - 1) {
        const nextTimer = updatedTimers[currentIndex + 1];
        setTimeRemaining(TIMER_TYPES[nextTimer.type as TimerType].duration);
        setCurrentTimer(nextTimer);
        setTimeRemaining(TIMER_TYPES[nextTimer.type].duration);
      } else {
        setCurrentTimer(null);
        setTimeRemaining(0);
        setIsPaused(true);
      }
      
      setActiveTimers(updatedTimers);
    }
  };

  const formatTime = (seconds: any) => {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    return `${String(minutes).padStart(2, '0')}:${String(remainingSeconds).padStart(2, '0')}`;
  };

  const progressPercent = currentTimer
    ? ((TIMER_TYPES[currentTimer.type].duration - timeRemaining) / TIMER_TYPES[currentTimer.type].duration) * 100
    : 0;

  // Get appropriate color for timer type
  const getTimerColor = (type: TimerType) => {
    switch(type) {
      case 'FOCUS':
        return 'text-primary'; // Crimson
      case 'SHORT_BREAK':
        return 'text-secondary'; // Soft teal
      case 'LONG_BREAK':
        return 'text-secondary/80'; // Lighter teal
      default:
        return 'text-primary';
    }
  };

  return (
    <motion.div
      initial={{ x: -300, opacity: 0 }}
      animate={{ x: 0, opacity: 1 }}
      exit={{ x: -300, opacity: 0 }}
      className="space-y-6"
    >
      <AnimatePresence>
        {activeTimers.map((timer, index) => (
          <motion.div
            key={timer.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            className={`relative w-48 h-48 ${timer.completed ? 'opacity-50' : ''}`}
          >
            {timer === currentTimer && (
              <motion.div
                className="absolute inset-0"
                animate={{
                  rotate: [0, 360],
                }}
                transition={{
                  duration: 60,
                  repeat: Infinity,
                  ease: "linear"
                }}
              >
                <Circle
                  className={`w-full h-full ${getTimerColor(timer.type)}`}
                  strokeWidth={2}
                  style={{
                    strokeDasharray: '1000',
                    strokeDashoffset: 1000 - (progressPercent * 10),
                  }}
                />
              </motion.div>
            )}
            <div className="absolute inset-0 flex flex-col items-center justify-center">
              <span className={`text-sm font-medium ${timer.type === 'FOCUS' ? 'text-primary' : 'text-secondary'}`}>
                {TIMER_TYPES[timer.type].label}
              </span>
              <span className={`text-3xl font-bold ${getTimerColor(timer.type)}`}>
                {timer === currentTimer ? formatTime(timeRemaining) : formatTime(TIMER_TYPES[timer.type].duration)}
              </span>
            </div>
          </motion.div>
        ))}
      </AnimatePresence>

      {currentTimer && (
        <div className="flex justify-center space-x-2">
          <Button
            variant="outline"
            size="icon"
            onClick={() => setIsPaused(!isPaused)}
            className="border-primary hover:bg-primary/10"
          >
            {isPaused ? <Play className="h-4 w-4 text-primary" /> : <Pause className="h-4 w-4 text-primary" />}
          </Button>
          <Button
            variant="outline"
            size="icon"
            onClick={handleTimerComplete}
            className="border-secondary hover:bg-secondary/10"
          >
            <SkipForward className="h-4 w-4 text-secondary" />
          </Button>
        </div>
      )}
    </motion.div>
  );
};

export default TimerStack;