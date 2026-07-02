package api

// openAPISpec is served at GET /openapi.json.
// ChatGPT Custom GPT Actions, Gemini plugins, and any OpenAPI-aware AI client
// reads this to discover and call flow's endpoints.
const openAPISpec = `{
  "openapi": "3.1.0",
  "info": {
    "title": "flow",
    "description": "Neurodivergent-aware task and mood tracker. Matches tasks to your current energy level. One answer at a time.",
    "version": "1.0.0"
  },
  "servers": [{ "url": "http://localhost:7777" }],
  "security": [{ "bearerAuth": [] }],
  "paths": {
    "/context": {
      "get": {
        "operationId": "getContext",
        "summary": "Get current state",
        "description": "Returns all pending tasks, the latest mood/energy check-in, and recent journal notes. Call this first before giving the user any advice or task suggestions.",
        "responses": {
          "200": {
            "description": "Current state",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Context" }
              }
            }
          }
        }
      }
    },
    "/tasks": {
      "get": {
        "operationId": "listTasks",
        "summary": "List tasks",
        "parameters": [
          {
            "name": "all",
            "in": "query",
            "description": "Set to true to include completed tasks",
            "schema": { "type": "boolean" }
          }
        ],
        "responses": {
          "200": {
            "description": "List of tasks",
            "content": {
              "application/json": {
                "schema": { "type": "array", "items": { "$ref": "#/components/schemas/Task" } }
              }
            }
          }
        }
      },
      "post": {
        "operationId": "addTask",
        "summary": "Add a task",
        "description": "Capture a new task. Infer size and energy from context if the user doesn't specify.",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/AddTaskRequest" }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Task created",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Task" }
              }
            }
          }
        }
      }
    },
    "/tasks/suggest": {
      "get": {
        "operationId": "suggestTask",
        "summary": "Suggest the best task right now",
        "description": "Returns the single best task based on the user's latest energy check-in. Use this when the user asks what to work on. Never return a list — one answer only.",
        "responses": {
          "200": {
            "description": "Suggested task",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/SuggestResponse" }
              }
            }
          }
        }
      }
    },
    "/tasks/{id}/complete": {
      "put": {
        "operationId": "completeTask",
        "summary": "Mark a task as done",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "responses": {
          "200": { "description": "Task completed" }
        }
      }
    },
    "/tasks/{id}/breakdown": {
      "post": {
        "operationId": "breakdownTask",
        "summary": "Break a task into micro-steps",
        "description": "Generate 3-5 concrete micro-steps for the task using your own reasoning, then store them as subtasks.",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/BreakdownRequest" }
            }
          }
        },
        "responses": {
          "201": { "description": "Subtasks created" }
        }
      }
    },
    "/checkins": {
      "post": {
        "operationId": "checkin",
        "summary": "Log mood and energy",
        "description": "Record the user's current mood (1-5) and energy (1-5). Ask them naturally in conversation then log it. This affects which tasks get suggested.",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/CheckinRequest" }
            }
          }
        },
        "responses": {
          "201": { "description": "Check-in saved" }
        }
      }
    },
    "/notes": {
      "post": {
        "operationId": "addNote",
        "summary": "Save a quick note",
        "description": "Capture a thought, feeling, or observation into the journal.",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/NoteRequest" }
            }
          }
        },
        "responses": {
          "201": { "description": "Note saved" }
        }
      }
    },
    "/insights": {
      "get": {
        "operationId": "getInsights",
        "summary": "Mood trends and task stats",
        "description": "Returns average mood, average energy, and task completion rate for the past 7 days.",
        "responses": {
          "200": {
            "description": "Insights",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Insights" }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "securitySchemes": {
      "bearerAuth": {
        "type": "http",
        "scheme": "bearer"
      }
    },
    "schemas": {
      "Task": {
        "type": "object",
        "properties": {
          "id":     { "type": "string" },
          "title":  { "type": "string" },
          "size":   { "type": "string", "enum": ["xs","s","m","l","xl"] },
          "energy": { "type": "string", "enum": ["low","med","high"] },
          "status": { "type": "string", "enum": ["todo","doing","done"] },
          "tags":   { "type": "array", "items": { "type": "string" } },
          "created_at":   { "type": "string", "format": "date-time" },
          "completed_at": { "type": "string", "format": "date-time", "nullable": true }
        }
      },
      "AddTaskRequest": {
        "type": "object",
        "required": ["title"],
        "properties": {
          "title":     { "type": "string" },
          "size":      { "type": "string", "enum": ["xs","s","m","l","xl"], "default": "m" },
          "energy":    { "type": "string", "enum": ["low","med","high"], "default": "med" },
          "tags":      { "type": "array", "items": { "type": "string" } },
          "parent_id": { "type": "string", "nullable": true }
        }
      },
      "BreakdownRequest": {
        "type": "object",
        "required": ["steps"],
        "properties": {
          "steps": { "type": "array", "items": { "type": "string" }, "minItems": 1 }
        }
      },
      "CheckinRequest": {
        "type": "object",
        "required": ["mood","energy"],
        "properties": {
          "mood":   { "type": "integer", "minimum": 1, "maximum": 5 },
          "energy": { "type": "integer", "minimum": 1, "maximum": 5 },
          "note":   { "type": "string" }
        }
      },
      "NoteRequest": {
        "type": "object",
        "required": ["content"],
        "properties": {
          "content": { "type": "string" },
          "tags":    { "type": "array", "items": { "type": "string" } }
        }
      },
      "SuggestResponse": {
        "type": "object",
        "properties": {
          "task":             { "$ref": "#/components/schemas/Task", "nullable": true },
          "based_on_energy":  { "type": "integer" },
          "message":          { "type": "string" }
        }
      },
      "Context": {
        "type": "object",
        "properties": {
          "tasks":          { "type": "array", "items": { "$ref": "#/components/schemas/Task" } },
          "latest_checkin": { "type": "object", "nullable": true },
          "recent_notes":   { "type": "array", "items": { "type": "object" } }
        }
      },
      "Insights": {
        "type": "object",
        "properties": {
          "period_days":     { "type": "integer" },
          "checkin_count":   { "type": "integer" },
          "avg_mood":        { "type": "number" },
          "avg_energy":      { "type": "number" },
          "tasks_total":     { "type": "integer" },
          "tasks_completed": { "type": "integer" }
        }
      }
    }
  }
}`
