import React from 'react';
import { Circle } from 'lucide-react';
import { motion } from 'framer-motion';
import { useStore } from '../store/useStore';

export const Timer: React.FC = () => {
  const { isTimerActive, remainingTime } = useStore();

  const minutes = Math.floor(remainingTime / 60);
  const seconds = remainingTime % 60;

  return (
    <div className="relative w-48 h-48">
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
          className="w-full h-full text-indigo-500 opacity-20"
          strokeWidth={2}
        />
      </motion.div>
      <div className="absolute inset-0 flex items-center justify-center">
        <span className="text-4xl font-bold text-indigo-600">
          {String(minutes).padStart(2, '0')}:{String(seconds).padStart(2, '0')}
        </span>
      </div>
    </div>
  );
};