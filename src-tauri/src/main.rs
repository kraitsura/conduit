use serde::{Deserialize, Serialize};
use tauri::State;
use reqwest::Client;
use dotenv::dotenv;
use std::env;

#[derive(Debug, Serialize, Deserialize)]
struct RawSubTask {
    description: String,
    #[serde(deserialize_with = "deserialize_time_estimate")]
    time_estimate: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct SubTask {
    description: String,
    time_estimate: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct StructureStep {
    step: String,
    details: Vec<String>,
    time_estimate: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct Task {
    description: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct SubTaskSelection {
    selected_subtasks: Vec<SubTask>,
}

struct AppState {
    client: Client,
    api_key: String,
}

// Custom error type for better error handling
#[derive(Debug, Serialize)]
enum AppError {
    MissingApiKey,
    ApiError(String),
    NetworkError(String),
    ParseError(String),
}

impl std::fmt::Display for AppError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            AppError::MissingApiKey => write!(f, "OpenAI API key is missing"),
            AppError::ApiError(msg) => write!(f, "API Error: {}", msg),
            AppError::NetworkError(msg) => write!(f, "Network Error: {}", msg),
            AppError::ParseError(msg) => write!(f, "Parse Error: {}", msg),
        }
    }
}

impl AppState {
    fn new() -> Result<Self, AppError> {
        let api_key = env::var("OPENAI_API_KEY")
            .map_err(|_| AppError::MissingApiKey)?;
        
        println!("API Key found: {}", if api_key.is_empty() { "No" } else { "Yes" });
        
        Ok(Self {
            client: Client::new(),
            api_key,
        })
    }
}

// Test command to verify Tauri is working
#[tauri::command]
fn test_backend() -> Result<String, String> {
    Ok("Backend is working!".to_string())
}

// Custom deserializer to handle both string and number time estimates
fn deserialize_time_estimate<'de, D>(deserializer: D) -> Result<String, D::Error>
where
    D: serde::Deserializer<'de>,
{
    use serde::de::Error;
    
    #[derive(Deserialize)]
    #[serde(untagged)]
    enum TimeEstimate {
        String(String),
        Number(i64),
    }

    match TimeEstimate::deserialize(deserializer)? {
        TimeEstimate::String(s) => Ok(s),
        TimeEstimate::Number(n) => Ok(format!("{} minutes", n)),
    }
}

#[tauri::command]
async fn get_subtasks(
    task: Task,
    state: State<'_, AppState>,
) -> Result<Vec<SubTask>, String> {
    println!("Received task: {:?}", task);

    let response = state.client
        .post("https://api.openai.com/v1/chat/completions")
        .header("Authorization", format!("Bearer {}", state.api_key))
        .json(&serde_json::json!({
            "model": "gpt-3.5-turbo",
            "messages": [
                {
                    "role": "system",
                    "content": "You are an expert assistant with a talent for deconstructing complex tasks into clear, manageable subtasks. You generate multiple options for each step, allowing users to select the best approach and create a personalized task list to achieve their goals efficiently. Provide your response in JSON format."
                },
                {
                    "role": "user",
                    "content": format!(
                        "Break down the following task into subtasks, including an estimated time for each subtask in minutes. \
                        Respond with a JSON array of objects, each with 'description' and 'time_estimate' fields. \
                        Task: {}", 
                        task.description
                    )
                }
            ],
            "max_tokens": 1000  // Increased token limit
        }))
        .send()
        .await
        .map_err(|e| {
            println!("Network error: {:?}", e);
            format!("Network error: {}", e)
        })?;

    println!("OpenAI API Response Status: {}", response.status());

    if !response.status().is_success() {
        let error_text = response.text().await.unwrap_or_else(|_| "Unknown error".to_string());
        println!("API error response: {}", error_text);
        return Err(format!("API error: {}", error_text));
    }

    let response_data: serde_json::Value = response.json().await.map_err(|e| {
        println!("JSON parse error: {:?}", e);
        format!("Failed to parse response: {}", e)
    })?;

    println!("Raw API Response: {:?}", response_data);

    let content = response_data["choices"][0]["message"]["content"]
        .as_str()
        .ok_or_else(|| {
            println!("Failed to get content from response");
            "Failed to get response content".to_string()
        })?;

    let cleaned_content = content
        .trim()
        .trim_start_matches("```json")
        .trim_end_matches("```")
        .trim();

    println!("Cleaned content: {}", cleaned_content);

    let raw_subtasks: Vec<RawSubTask> = serde_json::from_str(cleaned_content).map_err(|e| {
        println!("Failed to parse subtasks: {:?}", e);
        format!("Failed to parse subtasks: {}", e)
    })?;

    // Convert RawSubTask to SubTask
    let subtasks = raw_subtasks.into_iter()
        .map(|raw| SubTask {
            description: raw.description,
            time_estimate: raw.time_estimate,
        })
        .collect();

    println!("Successfully parsed subtasks");
    Ok(subtasks)
}

fn main() {
    dotenv().ok();
    
    println!("Starting Tauri application...");
    
    match AppState::new() {
        Ok(state) => {
            println!("AppState initialized successfully");
            tauri::Builder::default()
                .manage(state)
                .invoke_handler(tauri::generate_handler![test_backend, get_subtasks])
                .run(tauri::generate_context!())
                .expect("error while running tauri application");
        }
        Err(e) => {
            println!("Failed to initialize AppState: {}", e);
            std::process::exit(1);
        }
    }
}