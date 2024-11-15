import { Input } from "./ui/input"
import { useState } from 'react'
import { useTaskBreakdown } from '../hooks/useTaskBreakdown'

export function TaskBreak() {
  const [task, setTask] = useState<string>('')
  const [selectedSubtasks, setSelectedSubtasks] = useState<number[]>([])
  const { loading, error, subtasks, structure, getSubtasks, getOverallStructure } = useTaskBreakdown()

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    try {
      await getSubtasks(task)
      setSelectedSubtasks([])
    } catch (error) {
      // Error handling is managed by the hook
    }
  }

  const handleSubtaskSelection = (index: number) => {
    setSelectedSubtasks(prev => 
      prev.includes(index) ? prev.filter(i => i !== index) : [...prev, index]
    )
  }

  const handleGetStructure = async () => {
    try {
      const selectedSubtaskData = selectedSubtasks.map(index => subtasks[index])
      await getOverallStructure(selectedSubtaskData)
    } catch (error) {
      // Error handling is managed by the hook
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-900 to-purple-900 p-4">
      <div className="container mx-auto max-w-3xl">
        <h1 className="text-4xl font-bold mb-8 text-white drop-shadow-lg text-center">
          TaskForce
          <span className="block text-lg font-normal mt-2 text-blue-200">AI-Powered Task Breakdown</span>
        </h1>

        {error && (
          <div className="mb-6 animate-fadeIn">
            <p className="text-red-500 text-center bg-white/90 backdrop-blur-sm rounded-lg p-3 shadow-lg">
              {error}
            </p>
          </div>
        )}

        <div className="bg-white/10 backdrop-blur-md rounded-xl p-6 shadow-2xl mb-6">
          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              type="text"
              value={task}
              onChange={(e) => setTask(e.target.value)}
              placeholder="Enter your task here..."
              className="w-full px-4 py-3 rounded-lg border-2 border-transparent bg-white/90 backdrop-blur-sm focus:outline-none focus:border-blue-400 transition-all duration-300"
            />
            <button 
              type="submit" 
              className="w-full px-4 py-3 bg-blue-600 text-white rounded-lg disabled:bg-blue-400 hover:bg-blue-700 transition-all duration-300 transform hover:scale-[1.02] disabled:hover:scale-100"
              disabled={loading || !task.trim()}
            >
              {loading ? (
                <span className="flex items-center justify-center">
                  <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Processing...
                </span>
              ) : 'Get Subtasks'}
            </button>
          </form>
        </div>

        {subtasks.length > 0 && (
          <div className="bg-white/10 backdrop-blur-md rounded-xl p-6 shadow-2xl mb-6 animate-fadeIn">
            <h2 className="text-xl font-semibold mb-4 text-white">Select Subtasks</h2>
            <div className="space-y-3">
              {subtasks.map((subtask, index) => (
                <label key={index} className="flex items-center p-3 rounded-lg bg-white/20 hover:bg-white/30 transition-colors duration-200 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={selectedSubtasks.includes(index)}
                    onChange={() => handleSubtaskSelection(index)}
                    className="mr-3 h-4 w-4"
                  />
                  <span className="text-white">{subtask.description} - {subtask.time_estimate}</span>
                </label>
              ))}
            </div>
            <button 
              onClick={handleGetStructure}
              className="mt-4 w-full px-4 py-3 bg-emerald-600 text-white rounded-lg disabled:bg-emerald-400 hover:bg-emerald-700 transition-all duration-300 transform hover:scale-[1.02] disabled:hover:scale-100"
              disabled={loading || selectedSubtasks.length === 0}
            >
              {loading ? 'Generating Structure...' : 'Generate Plan'}
            </button>
          </div>
        )}

        {structure.length > 0 && (
          <div className="bg-white/10 backdrop-blur-md rounded-xl p-6 shadow-2xl animate-fadeIn">
            <h2 className="text-xl font-semibold mb-4 text-white">Project Structure</h2>
            <div className="space-y-6">
              {structure.map((step, index) => (
                <div key={index} className="bg-white/20 rounded-lg p-4">
                  <h3 className="font-semibold text-white">{index + 1}. {step.step} - {step.time_estimate}</h3>
                  <ul className="mt-2 space-y-2 list-disc list-inside">
                    {step.details.map((detail, detailIndex) => (
                      <li key={detailIndex} className="text-blue-100 ml-4">{detail}</li>
                    ))}
                  </ul>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}