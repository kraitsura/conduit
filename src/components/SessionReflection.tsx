import React, { useState } from 'react';
import { BookOpen } from 'lucide-react';
import { motion } from 'framer-motion';
import { useStore } from '../store/useStore';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

export const SessionReflection: React.FC = () => {
  const [isOpen, setIsOpen] = useState(false);
  const { currentSession, setReflection } = useStore();
  const [reflectionText, setReflectionText] = useState('');

  if (!currentSession) return null;

  return (
    <motion.div
      className="fixed bottom-4 right-4"
      initial={{ y: 100 }}
      animate={{ y: 0 }}
    >
      <Button
        onClick={() => setIsOpen(!isOpen)}
        className="bg-primary text-primary-foreground p-3 rounded-full shadow-lg hover:bg-primary/90 transition-colors"
        size="icon"
      >
        <BookOpen className="w-6 h-6" />
      </Button>

      {isOpen && (
        <motion.div
          className="absolute bottom-16 right-0 w-80"
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
        >
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-lg">Session Reflection</CardTitle>
            </CardHeader>
            <CardContent>
              <textarea
                value={reflectionText}
                onChange={(e) => {
                  setReflectionText(e.target.value);
                  setReflection(e.target.value);
                }}
                placeholder="What went well? What could be improved?"
                className="w-full h-32 p-2 border rounded resize-none bg-background text-foreground"
              />
            </CardContent>
          </Card>
        </motion.div>
      )}
    </motion.div>
  );
};