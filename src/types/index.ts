export interface SubTask {
    description: string;
    time_estimate: string;
}

export interface StructureStep {
    step: string;
    details: string[];
    time_estimate: string;
}

export interface Task {
    description: string;
}

export interface SubTaskSelection {
    selected_subtasks: SubTask[];
}