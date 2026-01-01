use serde::{Deserialize, Serialize};
use tauri::State;
use reqwest::Client;
use dotenv::dotenv;
use std::env;

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
enum MentalEnergy {
    High,
    Medium,
    Low,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
enum TaskType {
    Creative,
    Analytical,
    DecisionMaking,
    Execution,
    Learning,
}

#[derive(Debug, Serialize, Deserialize)]
struct Dependencies {
    prerequisite_tasks: Vec<u32>,
    unlocks: Vec<u32>,
    critical_path: bool,
}

#[derive(Debug, Serialize, Deserialize)]
struct Resources {
    tools: Vec<String>,
    references: Vec<String>,
}

#[derive(Debug, Serialize, Deserialize)]
struct SubTask {
    id: u32, 
    description: String,
    time_estimate_minutes: u32,
    mental_energy: MentalEnergy,
    task_type: TaskType,
    dependencies: Dependencies,
    resources: Resources,
}

#[derive(Debug, Serialize, Deserialize)]
struct Task {
    description: String,
}

struct AppState {
    client: Client,
    api_key: String,
}

impl AppState {
    fn new() -> Result<Self, String> {
        let api_key = env::var("OPENAI_API_KEY")
            .map_err(|_| "OPENAI_API_KEY not found in environment".to_string())?;
        
        Ok(Self {
            client: Client::new(),
            api_key,
        })
    }
}

#[tauri::command]
async fn test_backend() -> Result<String, String> {
    Ok("Backend connection successful".to_string())
}


#[tauri::command]
async fn get_subtasks(
    task: Task,
    state: State<'_, AppState>,
) -> Result<Vec<SubTask>, String> {
    println!("Received task: {:?}", task);

    let system_prompt = r#"You are an AI expert in task decomposition and project management. Your role is to break down complex tasks into logical, actionable subtasks while maintaining dependencies and providing detailed metadata for optimal execution.

    Principles to Follow:
    1. Each subtask must be specific, actionable, and self-contained
    2. Time estimates should be realistic and account for complexity
    3. Mental energy levels must accurately reflect cognitive load
    4. Task types must align with the actual cognitive requirements
    5. Dependencies must form a logical sequence
    6. Resource recommendations should be specific and relevant
    
    Validate your response before sending:
    - All subtasks have clear, actionable descriptions
    - Time estimates are reasonable (15-240 minutes)
    - Dependencies form a valid graph without cycles
    - All fields use the exact specified format"#;
    
        let user_prompt = format!(
            r#"Break down this task into 2-6 subtasks: "{}"
    
    Provide a JSON response with this exact structure:
    {{
        "subtasks": [
            {{
                "id": <unique number>,
                "description": "Clear, actionable subtask description",
                "time_estimate_minutes": <number>,
                "mental_energy": "high"|"medium"|"low",
                "task_type": "creative"|"analytical"|"decision_making"|"execution"|"learning",
                "dependencies": {{
                    "prerequisite_tasks": [<ID numbers of required previous tasks>],
                    "unlocks": [<ID numbers of tasks this enables>],
                    "critical_path": <boolean>
                }},
                "resources": {{
                    "tools": ["Specific tools, software, or platforms needed"],
                    "references": ["Relevant documentation, guides, or materials"]
                }}
            }}
        ]
    }}
    
    Requirements:
    1. Each subtask must take 15-240 minutes
    2. Dependencies must form a valid sequence
    3. Critical path should be marked for key subtasks
    4. Resource lists must be specific and actionable"#,
            task.description
        );

    let response = state.client
        .post("https://api.openai.com/v1/chat/completions")
        .header("Authorization", format!("Bearer {}", state.api_key))
        .json(&serde_json::json!({
            "model": "gpt-4-turbo-preview",
            "messages": [
                {
                    "role": "system",
                    "content": system_prompt
                },
                {
                    "role": "user",
                    "content": user_prompt
                }
            ],
            "max_tokens": 2000,
            "temperature": 0.7, 
            "response_format": { "type": "json_object" }
        }))
        .send()
        .await
        .map_err(|e| format!("Network error: {}", e))?;

    if !response.status().is_success() {
        let error_text = response.text().await.unwrap_or_else(|_| "Unknown error".to_string());
        return Err(format!("API error: {}", error_text));
    }

    let response_data: serde_json::Value = response.json().await
        .map_err(|e| format!("Failed to parse OpenAI response: {}", e))?;

    let content = response_data["choices"][0]["message"]["content"]
        .as_str()
        .ok_or("Failed to get response content")?;

    println!("Raw GPT response: {}", content);

    let json_value: serde_json::Value = serde_json::from_str(content)
        .map_err(|e| {
            println!("Invalid JSON received: {}", content);
            format!("Failed to parse content as JSON: {}", e)
        })?;

    // Extract the subtasks array from the wrapper object
    let subtasks_value = json_value.get("subtasks")
        .ok_or("Response missing 'subtasks' array")?;

    // Validate before converting, store pretty-printed JSON for error cases
    let pretty_json = serde_json::to_string_pretty(subtasks_value)
        .unwrap_or_else(|_| subtasks_value.to_string());

    if let Err(validation_error) = validate_subtask_json(subtasks_value) {
        println!("JSON Validation Failed. Received JSON: {}", pretty_json);
        return Err(format!("Invalid JSON structure: {}", validation_error));
    }

    // Now try to parse into our structs
    match serde_json::from_value(subtasks_value.clone()) {
        Ok(tasks) => Ok(tasks),
        Err(e) => {
            println!("JSON to Struct Conversion Failed. JSON structure:\n{}", pretty_json);
            Err(format!("Failed to convert JSON to SubTask structs: {}", e))
        }
    }
}

fn validate_subtask_json(json: &serde_json::Value) -> Result<(), String> {
    let subtasks = json.as_array()
        .ok_or_else(|| "Expected an array of subtasks".to_string())?;

    if subtasks.is_empty() || subtasks.len() > 6 {
        return Err(format!("Expected 2-6 subtasks, got {}", subtasks.len()));
    }

    for (index, subtask) in subtasks.iter().enumerate() {
        let subtask_obj = subtask.as_object()
            .ok_or_else(|| format!("Subtask {} is not an object", index))?;

        // Validate time estimate
        if let Some(time) = subtask_obj.get("time_estimate_minutes") {
            let minutes = time.as_i64()
                .ok_or_else(|| format!("Invalid time estimate in subtask {}", index))?;
            if minutes < 15 || minutes > 240 {
                return Err(format!("Time estimate must be between 15 and 240 minutes in subtask {}", index));
            }
        }

        // Check required fields
        for field in ["description", "time_estimate_minutes", "mental_energy", "task_type", "dependencies", "resources"] {
            if !subtask_obj.contains_key(field) {
                return Err(format!("Subtask {} missing required field: {}", index, field));
            }
        }


        // Validate mental_energy
        if let Some(energy) = subtask_obj.get("mental_energy") {
            let energy_str = match energy {
                serde_json::Value::String(s) => s.as_str(),
                serde_json::Value::Array(arr) if arr.len() == 1 => {
                    arr[0].as_str().ok_or_else(|| format!("Invalid mental_energy array value in subtask {}", index))?
                },
                _ => return Err(format!("Invalid mental_energy format in subtask {}", index))
            };
            
            if !["high", "medium", "low"].contains(&energy_str) {
                return Err(format!("Invalid mental_energy value in subtask {}: {}", index, energy_str));
            }
        }

        // Validate task_type
        if let Some(task_type) = subtask_obj.get("task_type") {
            let task_types = match task_type {
                serde_json::Value::String(s) => vec![s.as_str()],
                serde_json::Value::Array(arr) => arr.iter()
                    .map(|v| v.as_str().ok_or_else(|| format!("Invalid task_type array value in subtask {}", index)))
                    .collect::<Result<Vec<_>, String>>()?,
                _ => return Err(format!("Invalid task_type format in subtask {}", index))
            };

            let valid_types = ["creative", "analytical", "decision_making", "execution", "learning"];
            for t in task_types {
                if !valid_types.contains(&t) {
                    return Err(format!("Invalid task_type value in subtask {}: {}", index, t));
                }
            }
        }
    }

    validate_dependency_graph(subtasks)?;

    Ok(())
}

fn validate_dependency_graph(subtasks: &[serde_json::Value]) -> Result<(), String> {
    let mut dependency_map: std::collections::HashMap<u32, Vec<u32>> = std::collections::HashMap::new();
    
    // Build dependency graph
    for subtask in subtasks {
        if let Some(deps) = subtask["dependencies"]["prerequisite_tasks"].as_array() {
            let task_id = subtask["id"].as_u64()
                .ok_or_else(|| "Invalid task ID".to_string())? as u32;
            
            let prerequisites: Vec<u32> = deps.iter()
                .filter_map(|d| d.as_u64().map(|n| n as u32))
                .collect();
            
            dependency_map.insert(task_id, prerequisites);
        }
    }

    // Check for cycles using DFS
    let mut visited = std::collections::HashSet::new();
    let mut path = std::collections::HashSet::new();

    for task in dependency_map.keys() {
        if has_cycle(&dependency_map, *task, &mut visited, &mut path) {
            return Err("Dependency cycle detected".to_string());
        }
    }

    Ok(())
}

fn has_cycle(
    graph: &std::collections::HashMap<u32, Vec<u32>>,
    current: u32,
    visited: &mut std::collections::HashSet<u32>,
    path: &mut std::collections::HashSet<u32>,
) -> bool {
    if path.contains(&current) {
        return true;
    }
    if visited.contains(&current) {
        return false;
    }

    visited.insert(current);
    path.insert(current);

    if let Some(deps) = graph.get(&current) {
        for &dep in deps {
            if has_cycle(graph, dep, visited, path) {
                return true;
            }
        }
    }

    path.remove(&current);
    false
}

fn main() {
    dotenv().ok();
    
    match AppState::new() {
        Ok(state) => {
            tauri::Builder::default()
                .manage(state)
                .invoke_handler(tauri::generate_handler![test_backend, get_subtasks])
                .run(tauri::generate_context!())
                .expect("error while running tauri application");
        }
        Err(e) => {
            eprintln!("Failed to initialize AppState: {}", e);
            std::process::exit(1);
        }
    }
}


