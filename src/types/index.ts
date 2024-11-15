// types/index.ts
export enum MentalEnergy {
    High = 'high',
    Medium = 'medium',
    Low = 'low'
}

export enum TaskType {
    Creative = 'creative',
    Analytical = 'analytical',
    DecisionMaking = 'decision_making',
    Execution = 'execution',
    Learning = 'learning'
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
    id: number;
    description: string;
    time_estimate_minutes: number;
    mental_energy: MentalEnergy;
    task_type: TaskType;
    dependencies: Dependencies;
    resources: Resources;
}

export interface Task {
    description: string;
}