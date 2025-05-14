// types/index.ts
export enum MentalEnergy {
	High = "high",
	Medium = "medium",
	Low = "low",
}

export enum TaskType {
	Creative = "creative",
	Analytical = "analytical",
	DecisionMaking = "decision_making",
	Execution = "execution",
	Learning = "learning",
}

export interface Dependencies {
	prerequisite_tasks: number[];
	unlocks: number[];
	critical_path: boolean;
}

export interface Resources {
	tools: string[];
	references: string[];
	collaborators: string[];
}

export interface SubTask {
	id: string;
	description: string;
	time_estimate_minutes: number;
	mental_energy: MentalEnergy;
	task_type: TaskType;
	dependencies: Dependencies;
	resources: Resources;
	completed: boolean;
	date: Date;
}

export interface Task {
	id: string;
	description: string;
	subtasks: SubTask[];
	completed: boolean;
	date: Date;
}

export interface Session {
	id: string;
	startTime: Date;
	endTime?: Date;
	task: Task;
	subTasks: SubTask[];
	distractions?: Distraction[];
	reflections?: Reflection[];
}

export interface Distraction {
	id: string;
	timestamp: Date;
	description: string;
	type: "internal" | "external";
}

export interface Reflection {
	id: string;
	timestamp: Date;
	content: string;
}
